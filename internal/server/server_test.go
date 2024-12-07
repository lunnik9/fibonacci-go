package server_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"fibonacci/internal/domain"
	"fibonacci/internal/genproto/fibonacci-service/api"
	internalMock "fibonacci/internal/mock"
	"fibonacci/internal/server"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func TestFibonacciServer_Fibonacci(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, 10).Return([]string{"0", "1", "1", "2", "3", "5", "8", "13", "21", "34"}, nil)

		req := &api.FibonacciRequest{N: 10}
		res, err := s.Fibonacci(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []string{"0", "1", "1", "2", "3", "5", "8", "13", "21", "34"}, res.Values)
	})

	t.Run("negative N", func(t *testing.T) {
		ctx := context.Background()
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, -5).Return(nil, domain.ErrNegativeN)

		req := &api.FibonacciRequest{N: -5}
		res, err := s.Fibonacci(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrNegativeN).Error(), err.Error())
	})

	t.Run("N too large", func(t *testing.T) {
		ctx := context.Background()
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, 101).Return(nil, domain.ErrTooLargeN)

		req := &api.FibonacciRequest{N: 101}
		res, err := s.Fibonacci(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrTooLargeN).Error(), err.Error())
	})

	t.Run("N too large", func(t *testing.T) {
		ctx := context.Background()
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, 101).Return(nil, domain.ErrTooLargeN)

		req := &api.FibonacciRequest{N: 101}
		res, err := s.Fibonacci(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrTooLargeN).Error(), err.Error())
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()

		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, 10).Return(nil, domain.ErrContextCanceled)

		cancel()

		req := &api.FibonacciRequest{N: 10}
		res, err := s.Fibonacci(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, status.Errorf(http.StatusBadRequest, "Context canceled: %s", domain.ErrContextCanceled).Error(), err.Error())
	})

	t.Run("service shut down", func(t *testing.T) {
		ctx := context.Background()
		globalCtx, globalCancel := context.WithCancel(context.Background())

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)

		mockService.EXPECT().GetFibonacci(mock.Anything, 10).Return(nil, domain.ErrContextCanceled)
		globalCancel()

		req := &api.FibonacciRequest{N: 10}
		res, err := s.Fibonacci(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, status.Errorf(http.StatusServiceUnavailable, "Service unavailable: %s", domain.ErrContextCanceled).Error(), err.Error())
	})
}

func TestFibonacciServer_FibonacciStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 10, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, r domain.FibonacciStreamRequest) error {
				assert.Equal(t, 10, r.N)
				assert.Equal(t, 4, r.ChunkSize)
				err := r.SendFunc([]string{"0", "1", "1", "2"}, 0)
				assert.NoError(t, err)
				err = r.SendFunc([]string{"3", "5", "8", "13"}, 4)
				assert.NoError(t, err)
				return nil
			})

		stream.EXPECT().Context().Return(context.Background())
		stream.EXPECT().Send(mock.Anything).Return(nil).Twice()

		err := s.FibonacciStream(req, stream)
		assert.NoError(t, err)
	})

	t.Run("invalid chunk size", func(t *testing.T) {
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 10, ChunkSize: 0}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(domain.ErrInvalidChunkSize)

		stream.EXPECT().Context().Return(context.Background())

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrInvalidChunkSize).Error())
	})

	t.Run("negative N", func(t *testing.T) {
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: -5, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(domain.ErrNegativeN)

		stream.EXPECT().Context().Return(context.Background())

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrNegativeN).Error())
	})

	t.Run("N too large", func(t *testing.T) {
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 101, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(domain.ErrTooLargeN)

		stream.EXPECT().Context().Return(context.Background())

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusBadRequest, "Bad Request: %s", domain.ErrTooLargeN).Error())
	})

	t.Run("context canceled by request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		globalCtx := context.Background()
		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 10, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(domain.ErrContextCanceled)

		cancel()
		stream.EXPECT().Context().Return(ctx)

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusBadRequest, "Context canceled: %s", domain.ErrContextCanceled).Error())
	})

	t.Run("service shut down", func(t *testing.T) {
		globalCtx, globalCancel := context.WithCancel(context.Background())
		defer globalCancel()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()
		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		globalCancel()

		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 10, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(domain.ErrContextCanceled)

		stream.EXPECT().Context().Return(context.Background())

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusServiceUnavailable, "Service unavailable: %s", domain.ErrContextCanceled).Error())
	})

	t.Run("internal server error", func(t *testing.T) {
		globalCtx := context.Background()

		grpcServer := grpc.NewServer()
		mockService := internalMock.NewService(t)
		log := logrus.New()

		s := server.NewFibonacciServer(globalCtx, grpcServer, mockService, log)
		stream := internalMock.NewFibonacciChunkStreamServer(t)
		req := &api.FibonacciStreamRequest{N: 10, ChunkSize: 4}

		mockService.EXPECT().
			GetFibonacciStream(mock.Anything, mock.Anything).
			Return(errors.New("some internal error"))

		stream.EXPECT().Context().Return(context.Background())

		err := s.FibonacciStream(req, stream)
		assert.EqualError(t, err, status.Errorf(http.StatusInternalServerError, "Internal server error: %s", "some internal error").Error())
	})
}
