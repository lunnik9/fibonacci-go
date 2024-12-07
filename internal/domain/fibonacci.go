package domain

type FibonacciStreamRequest struct {
	N         int
	ChunkSize int
	SendFunc  func([]string, int) error
}
