package config

import (
	"io"

	"github.com/BurntSushi/toml"
)

type config struct {
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
	Location string
	Upstream string
}

func parse(r io.Reader) (*config, error) {
	cfg := config{}
	_, err := toml.DecodeReader(r, &cfg)
	return &cfg, err
}
