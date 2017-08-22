package config

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/tonto/gourmet/internal/upstream"
)

const (
	RoundRobinAlg = "round_robin"
)

const (
	StaticProvider = "static"
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
	return parse(r)
}

type Config struct {
	Upstreams map[string]Upstream
	Server    Server
}

type Upstream struct {
	Balancer string
	Provider string

	// Servers should be ignored if Provider is not static
	Servers []UpstreamServer
}

type UpstreamServer struct {
	Path   string
	Weight int
}

type Server struct {
	Port      int
	Locations []ServerLocation
}

type ServerLocation struct {
	RegEx    string `toml:"location"`
	Upstream string
}

func parse(r io.Reader) (*Config, error) {
	cfg := Config{}
	_, err := toml.DecodeReader(r, &cfg)
	return &cfg, err
}
