package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fibonacci/config"
	"fibonacci/internal/server"
	"fibonacci/internal/service"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	// Config setup
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	// Logs setup
	logger := logrus.New()
	level, err := logrus.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		panic(err)
	}
	logger.SetLevel(level)

	// Global context setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start Prometheus metrics server
	startMetricsServer(ctx, cfg.MetricsPort, logger)

	// Create Fibonacci service and gRPC server
	fibService := service.NewService(cfg.MaxChunkSize, cfg.MinChunkSize, cfg.NLimit, cfg.StreamNLimit)
	grpcServer := grpc.NewServer()
	fibServer := server.NewFibonacciServer(ctx, grpcServer, fibService, logger)
	if fibServer == nil {
		logger.Fatal("Failed to create Fibonacci server")
	}

	lis, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: :%v", cfg.AppPort, err)
	}

	logger.Infof("Starting gRPC server on :%s", cfg.AppPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("gRPC server stopped with error: %v", err)
			cancel()
		}
	}()

	shutdownInitiated := false

ShutdownLoop:
	for {
		select {
		case <-ctx.Done():
			break ShutdownLoop

		case sig := <-sigCh:
			switch sig {
			case syscall.SIGINT:
				if !shutdownInitiated {
					logger.Info("Received SIGINT. Performing soft shutdown...")
					shutdownInitiated = true

					go func() {
						grpcServer.GracefulStop()

						time.Sleep(20 * time.Second)

						cancel()
					}()
				} else {
					logger.Info("Received second SIGINT. Performing hard shutdown...")
					cancel()
				}
			case syscall.SIGTERM:
				logger.Info("Received SIGTERM. Triggering shutdown...")
				cancel()
			}
		}
	}

	logger.Info("Exiting...")
}

func startMetricsServer(ctx context.Context, port string, logger *logrus.Logger) {
	router := mux.NewRouter()

	router.Path("/metrics").Handler(promhttp.Handler())

	s := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting metrics server at :%s/metrics", port)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Metrics server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := s.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down metrics server: %v", err)
		} else {
			logger.Info("Metrics server stopped.")
		}
	}()
}
