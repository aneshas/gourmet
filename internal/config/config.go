package config

import (
	"io"

	"github.com/tonto/gourmet/internal/upstream"
)

// Provider represents an interface that should be implemented
// by upstream providers in order to supply load balancing targets
type Provider interface {
	Servers() []*upstream.Server

	// Sync returns a channel over which Provider
	// implementer should signal a change in the hosts config.
	// After receiving the signal receiver should call Upstreams() again
	// and reset the next pointer if applicable (such as with round-robin)
	Sync() chan struct{}
}

// TODO - Maybe accept io.Reader and implement it under platform
func New(r io.Reader) (*Config, error) {
	cfg, err := parse(r)
	if err != nil {
		return nil, err
	}
	c := Config{cfg}
	return &c, err
}

type Config struct {
	*config
}
