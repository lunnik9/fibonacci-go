package service_test

import (
	"context"
	"errors"
	"testing"

	"fibonacci/internal/domain"
	"fibonacci/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestGetFibonacci(t *testing.T) {
	s := service.NewService(10, 2, 100, 200)

	t.Run("valid input", func(t *testing.T) {
		ctx := context.Background()
		result, err := s.GetFibonacci(ctx, 10)

		assert.NoError(t, err)
		assert.Equal(t, []string{"0", "1", "1", "2", "3", "5", "8", "13", "21", "34"}, result)
	})

	t.Run("negative n", func(t *testing.T) {
		ctx := context.Background()
		_, err := s.GetFibonacci(ctx, -5)

		assert.ErrorIs(t, err, domain.ErrNegativeN)
	})

	t.Run("n exceeds limit", func(t *testing.T) {
		ctx := context.Background()
		_, err := s.GetFibonacci(ctx, 150)

		assert.ErrorIs(t, err, domain.ErrTooLargeN)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.GetFibonacci(ctx, 10)

		assert.ErrorIs(t, err, domain.ErrContextCanceled)
	})
}

func TestGetFibonacciStream(t *testing.T) {
	s := service.NewService(10, 2, 50, 100)

	t.Run("valid input", func(t *testing.T) {
		ctx := context.Background()
		chunks := [][]string{}
		sendFunc := func(values []string, index int) error {
			temp := make([]string, len(values))
			copy(temp, values)
			chunks = append(chunks, temp)

			return nil
		}

		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         10,
			ChunkSize: 4,
			SendFunc:  sendFunc,
		})

		assert.NoError(t, err)
		assert.Equal(t, [][]string{
			{"0", "1", "1", "2"},
			{"3", "5", "8", "13"},
			{"21", "34"},
		}, chunks)
	})

	t.Run("n exceeds limit", func(t *testing.T) {
		ctx := context.Background()
		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         150,
			ChunkSize: 20,
			SendFunc:  func([]string, int) error { return nil },
		})

		assert.ErrorIs(t, err, domain.ErrTooLargeN)
	})

	t.Run("chunk size too large", func(t *testing.T) {
		ctx := context.Background()
		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         10,
			ChunkSize: 20,
			SendFunc:  func([]string, int) error { return nil },
		})

		assert.ErrorIs(t, err, domain.ErrInvalidChunkSize)
	})

	t.Run("chunk size too small", func(t *testing.T) {
		ctx := context.Background()
		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         10,
			ChunkSize: 1,
			SendFunc:  func([]string, int) error { return nil },
		})

		assert.ErrorIs(t, err, domain.ErrInvalidChunkSize)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         10,
			ChunkSize: 4,
			SendFunc:  func([]string, int) error { return nil },
		})

		assert.ErrorIs(t, err, domain.ErrContextCanceled)
	})

	t.Run("send function error", func(t *testing.T) {
		ctx := context.Background()
		err := s.GetFibonacciStream(ctx, domain.FibonacciStreamRequest{
			N:         10,
			ChunkSize: 4,
			SendFunc: func([]string, int) error {
				return errors.New("send error")
			},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "send error")
	})
}
