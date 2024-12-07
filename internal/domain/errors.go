package domain

import "errors"

var (
	ErrInvalidChunkSize = errors.New("invalid chunk size")
	ErrNegativeN        = errors.New("n must be positive")
	ErrTooLargeN        = errors.New("to large n")
	ErrContextCanceled  = errors.New("context canceled")
)
