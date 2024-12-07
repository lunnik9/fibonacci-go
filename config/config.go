package config

type Config struct {
	MaxChunkSize int `env:"MAX_CHUNK_SIZE" envDefault:"100"`
	MinChunkSize int `env:"MIN_CHUNK_SIZE"  envDefault:"5"`
	NLimit       int `env:"N_LIMIT"  envDefault:"500"`
	StreamNLimit int `env:"STREAM_N_LIMIT"  envDefault:"1000"`

	AppPort     string `env:"APP_PORT" envDefault:"50051"`
	MetricsPort string `env:"PORT" envDefault:"8080"`

	LogLevel string `env:"PORT" envDefault:"info"`
}
