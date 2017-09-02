package upstream

import "time"

type ServerOption func(*ServerConfig)

func WithWeight(w int) ServerOption {
	return func(sc *ServerConfig) {
		sc.weight = w
	}
}

func WithFailTimeout(d time.Duration) ServerOption {
	return func(sc *ServerConfig) {
		sc.failTimeout = d
	}
}

func WithMaxFail(n int) ServerOption {
	return func(sc *ServerConfig) {
		sc.maxFail = n
	}
}

func WithQueueSize(n int) ServerOption {
	return func(sc *ServerConfig) {
		sc.queueBufferSz = n
	}
}
