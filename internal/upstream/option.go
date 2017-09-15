package upstream

import "time"

// ServerOption represents upstream server option
type ServerOption func(*ServerConfig)

func WithWeight(w int) ServerOption {
	return func(sc *ServerConfig) {
		sc.weight = w
	}
}

// WithFailTimeout sets fail timeout option for upstream server
func WithFailTimeout(d time.Duration) ServerOption {
	return func(sc *ServerConfig) {
		sc.failTimeout = d
	}
}

// WithMaxFail sets max fail option for upstream server
func WithMaxFail(n int) ServerOption {
	return func(sc *ServerConfig) {
		sc.maxFail = n
	}
}

// WithQueueSize sets upstream server queue size
func WithQueueSize(n int) ServerOption {
	return func(sc *ServerConfig) {
		sc.queueBufferSz = n
	}
}
