package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"fibonacci/internal/domain"
	"fibonacci/internal/metrics"
)

//go:generate mockery --name=Service --with-expecter --output=../mock --outpkg=mock --case=underscore

// Service defines the interface for Fibonacci calculations.
type Service interface {
	// GetFibonacci calculates the first n Fibonacci numbers.
	GetFibonacci(ctx context.Context, n int) ([]string, error)

	// GetFibonacciStream streams chunks of Fibonacci numbers based on the request.
	GetFibonacciStream(ctx context.Context, req domain.FibonacciStreamRequest) error
}

// FibonacciService implements the Service interface with additional constraints.
type fibonacciService struct {
	MaxChunkSize int // Maximum allowed chunk size for streaming
	MinChunkSize int // Minimum allowed chunk size for streaming
	NLimit       int // Maximum limit for the Fibonacci sequence length
	StreamNLimit int // Maximum limit for the Fibonacci streaming sequence length
}

func NewService(maxChunkSize int, minChunkSize int, nLimit, streamNLimit int) Service {
	return &fibonacciService{
		MaxChunkSize: maxChunkSize,
		MinChunkSize: minChunkSize,
		NLimit:       nLimit,
		StreamNLimit: streamNLimit,
	}
}

func (s *fibonacciService) GetFibonacci(ctx context.Context, n int) ([]string, error) {
	if n < 0 {
		return nil, domain.ErrNegativeN
	}

	if n > s.NLimit {
		return nil, fmt.Errorf("%w: must not exceed %d", domain.ErrTooLargeN, s.NLimit)
	}

	start := time.Now()

	res, err := getFibonacci(ctx, n)
	if err != nil {
		return nil, err
	}

	metrics.FibonacciCalculationDuration.WithLabelValues().Observe(float64(time.Since(start).Nanoseconds()))
	metrics.FibonacciCalculationsTotal.WithLabelValues(strconv.Itoa(n)).Inc()

	return res, nil
}

// getFibonacci generates the Fibonacci sequence up to n using dp.
func getFibonacci(ctx context.Context, n int) ([]string, error) {
	seq := make([]string, n)

	if n > 0 {
		seq[0] = "0"
	}

	if n > 1 {
		seq[1] = "1"
	}

	for i := 2; i < n; i++ {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil, domain.ErrContextCanceled
			}
			return nil, ctx.Err()
		default:
			seq[i] = addStrings(seq[i-1], seq[i-2])
		}
	}

	return seq, nil
}

func (s *fibonacciService) GetFibonacciStream(ctx context.Context, req domain.FibonacciStreamRequest) error {
	if req.N < 0 {
		return domain.ErrNegativeN
	}

	if req.N > s.StreamNLimit {
		return fmt.Errorf("%w: must not exceed %d", domain.ErrTooLargeN, s.NLimit)
	}

	if req.ChunkSize > s.MaxChunkSize {
		return fmt.Errorf("%w: must not exceed %d", domain.ErrInvalidChunkSize, s.MaxChunkSize)
	}

	if req.ChunkSize < s.MinChunkSize {
		return fmt.Errorf("%w: must be at least %d", domain.ErrInvalidChunkSize, s.MinChunkSize)
	}

	start := time.Now()

	err := processChunks(ctx, req.N, req.ChunkSize, req.SendFunc)
	if err != nil {
		return err
	}

	metrics.FibonacciStreamCalculationDuration.WithLabelValues().Observe(float64(time.Since(start).Nanoseconds()))
	metrics.FibonacciStreamCalculationsTotal.WithLabelValues(strconv.Itoa(req.N), strconv.Itoa(req.ChunkSize)).Inc()

	return nil
}

// processChunks divides the Fibonacci sequence into chunks and streams each chunk.
// it also utilizes only one array we
func processChunks(ctx context.Context, n, chunkSize int, send func([]string, int) error) error {
	var (
		prev1, prev2 = "0", "1"
		iteration    = 0
		chunk        = make([]string, chunkSize) // Reused chunk array
	)

	for i := 0; i < n; i += chunkSize {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return domain.ErrContextCanceled
			}

			return ctx.Err()
		default:
			iteration++
			end := i + chunkSize
			if end > n {
				end = n
			}

			currentChunkSize := end - i

			for j := 0; j < currentChunkSize; j++ {
				chunk[j] = prev1
				prev1, prev2 = prev2, addStrings(prev1, prev2)
			}

			err := send(chunk[:currentChunkSize], i)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addStrings(num1, num2 string) string {
	var result []byte
	carry := false
	i, j := len(num1)-1, len(num2)-1

	for i >= 0 || j >= 0 || carry {
		sum := 0

		if carry {
			sum++
		}

		if i >= 0 {
			sum += int(num1[i] - '0')
			i--
		}
		if j >= 0 {
			sum += int(num2[j] - '0')
			j--
		}

		if sum > 9 {
			carry = true
			sum -= 10
		} else {
			carry = false
		}

		result = append(result, byte(sum)+'0')
	}

	for k, l := 0, len(result)-1; k < l; k, l = k+1, l-1 {
		result[k], result[l] = result[l], result[k]
	}

	return string(result)
}
