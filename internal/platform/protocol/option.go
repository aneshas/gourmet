package protocol

import "time"

// HTTPOption represents http protocol config option
type HTTPOption func(*Config)

// WithHTTPHeaders set's provided headers to every request handled
func WithHTTPHeaders(h map[string]string) HTTPOption {
	return func(cfg *Config) {
		cfg.passHeaders = h
	}
}

// WithHTTPRequestTimeout sets a timeout for every upstream request
func WithHTTPRequestTimeout(d time.Duration) HTTPOption {
	return func(cfg *Config) {
		cfg.requestTimeout = d
	}
}
