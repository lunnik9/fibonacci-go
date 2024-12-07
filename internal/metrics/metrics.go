package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	FibonacciCalculationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fibonacci_calculation_duration_nanoseconds",
			Help:    "Time spent calculating Fibonacci sequences, labeled by n size.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	FibonacciStreamCalculationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fibonacci_stream_calculation_duration_nanoseconds",
			Help:    "Time spent calculating Fibonacci sequences, labeled by n and chunk size.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	FibonacciCalculationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fibonacci_calculations_total",
			Help: "Total number of Fibonacci calculations performed, labeled by n size.",
		},
		[]string{"n"},
	)

	FibonacciStreamCalculationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fibonacci_stream_calculations_total",
			Help: "Total number of Fibonacci calculations performed, labeled by n and chunk size.",
		},
		[]string{"n", "chunk_size"},
	)
)

func init() {
	prometheus.MustRegister(FibonacciCalculationDuration)
	prometheus.MustRegister(FibonacciStreamCalculationDuration)
	prometheus.MustRegister(FibonacciCalculationsTotal)
	prometheus.MustRegister(FibonacciStreamCalculationsTotal)
}
