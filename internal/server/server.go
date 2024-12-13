package server

import (
	"context"
	"errors"
	"net/http"

	"fibonacci/internal/domain"
	"fibonacci/internal/genproto/fibonacci-service/api"
	"fibonacci/internal/service"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// FibonacciServer handles gRPC requests for the Fibonacci service.
type FibonacciServer struct {
	api.UnimplementedFibonacciServiceServer
	service   service.Service
	globalCtx context.Context
	logger    *logrus.Logger
}

//go:generate mockery --name=FibonacciChunkStreamServer --with-expecter --output=../mock --outpkg=mock --case=underscore

// FibonacciChunkStreamServer provides an interface to enable mock generation for gRPC streams.
// This is a workaround due to limitations with generating mocks for generic gRPC interfaces.
type FibonacciChunkStreamServer interface {
	grpc.ServerStreamingServer[api.FibonacciChunk]
}

func NewFibonacciServer(ctx context.Context, s *grpc.Server, fibonacciService service.Service, logger *logrus.Logger) *FibonacciServer {
	if logger == nil {
		logger = logrus.New()
	}
	server := &FibonacciServer{
		service:   fibonacciService,
		globalCtx: ctx,
		logger:    logger,
	}

	reflection.Register(s)

	api.RegisterFibonacciServiceServer(s, server)

	return server
}

// FibonacciStream streams chunks of Fibonacci numbers to the client.
func (s *FibonacciServer) FibonacciStream(req *api.FibonacciStreamRequest, stream grpc.ServerStreamingServer[api.FibonacciChunk]) error {
	s.logger.Printf("FibonacciStream called with N=%d, ChunkSize=%d", req.GetN(), req.GetChunkSize())

	sendFunc := func(values []string, i int) error {
		return stream.Send(&api.FibonacciChunk{
			Index:  int32(i),
			Values: values,
		})
	}

	ctx, cancel := MergeContexts(stream.Context(), s.globalCtx)
	defer cancel()

	err := s.service.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
		N:         int(req.GetN()),
		ChunkSize: int(req.GetChunkSize()),
		SendFunc:  sendFunc,
	})

	if err != nil {
		s.logger.Printf("Error getting fibonacci stream: %v", err)

		if errors.Is(err, domain.ErrInvalidChunkSize) || errors.Is(err, domain.ErrNegativeN) || errors.Is(err, domain.ErrTooLargeN) {
			return status.Errorf(http.StatusBadRequest, "Bad Request: %s", err)
		} else if errors.Is(err, domain.ErrContextCanceled) && s.globalCtx.Err() != nil {
			return status.Errorf(http.StatusServiceUnavailable, "Service unavailable: %s", err)
		} else if errors.Is(err, domain.ErrContextCanceled) {
			return status.Errorf(http.StatusBadRequest, "Context canceled: %s", err)
		}

		return status.Errorf(http.StatusInternalServerError, "Internal server error: %s", err)
	}

	return nil
}

// Fibonacci calculates the entire Fibonacci sequence up to n and returns it.
func (s *FibonacciServer) Fibonacci(ctx context.Context, req *api.FibonacciRequest) (*api.FibonacciResponse, error) {
	s.logger.Printf("Fibonacci called with N=%d", req.GetN())

	ctx, cancel := MergeContexts(ctx, s.globalCtx)
	defer cancel()

	res, err := s.service.GetFibonacci(ctx, int(req.GetN()))

	if err != nil {
		s.logger.Printf("Error getting fibonacci stream: %v", err)

		if errors.Is(err, domain.ErrInvalidChunkSize) || errors.Is(err, domain.ErrNegativeN) || errors.Is(err, domain.ErrTooLargeN) {
			return nil, status.Errorf(http.StatusBadRequest, "Bad Request: %s", err)
		} else if errors.Is(err, domain.ErrContextCanceled) && s.globalCtx.Err() != nil {
			return nil, status.Errorf(http.StatusServiceUnavailable, "Service unavailable: %s", err)
		} else if errors.Is(err, domain.ErrContextCanceled) {
			return nil, status.Errorf(http.StatusBadRequest, "Context canceled: %s", err)
		}

		return nil, status.Errorf(http.StatusInternalServerError, "Internal server error: %s", err)
	}

	return &api.FibonacciResponse{Values: res}, nil
}

// MergeContexts combines two contexts into a single context.
func MergeContexts(c1, c2 context.Context) (context.Context, func()) {
	mergedCtx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-c1.Done():
			cancel()
		case <-c2.Done():
			cancel()
		case <-mergedCtx.Done():
			return
		}
	}()

	return mergedCtx, cancel
}
