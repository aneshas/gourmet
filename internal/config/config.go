// Package config provides gourmet configuration parsing functionality
package config

import (
	"errors"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/tonto/gourmet/internal/upstream"
)

const (
	// RoundRobinAlg represents round robin balancer config label
	RoundRobinAlg = "round_robin"
)

const (
	// StaticProvider represents static
	// upstream server provider config label
	StaticProvider = "static"
)

const (
	defaultPort = 8080
)

var (
	errUpstreamMismatch  = errors.New("upstreams in server location list don't match up with upstreams")
	errNoUpstreams       = errors.New("upstream block missing or no upstreams listed")
	errNoServers         = errors.New("if using static upstream server provider (default) server list should not be empty")
	errNoServerPath      = errors.New("upstream server path must not be empty")
	errNoServer          = errors.New("server block not present")
	errNoServerLocations = errors.New("no server locations block present")
	errInvalidTOML       = errors.New("invalid format for config file")
)

// Provider represents an interface that should be implemented
// by upstream providers in order to supply load balancing endpoints
type Provider interface {
	Servers() []*upstream.Server

	// Sync returns a channel over which Provider
	// implementer should signal a change in the hosts config.
	// After receiving the signal receiver should call Upstreams() again
	// and reset the next pointer if applicable (such as with round-robin)
	Sync() chan struct{}
}

// Parse parses config file and creates new config instance
func Parse(r io.Reader) (*Config, error) {
	cfg, err := parse(r)
	if err != nil {
		return nil, errInvalidTOML
	}

	err = cfg.validate()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func parse(r io.Reader) (*Config, error) {
	cfg := Config{}
	_, err := toml.DecodeReader(r, &cfg)
	return &cfg, err
}

// Config represents gourmet config struct and provides
// balancer instantiation methods
type Config struct {
	Upstreams map[string]*Upstream
	Server    *Server
}

// Upstream represents upstream config resource
type Upstream struct {
	Balancer string
	Provider string

	// Servers should be ignored if Provider is not static
	Servers []*UpstreamServer
}

// UpstreamServer represents upstream server config resource
type UpstreamServer struct {
	Path        string
	Weight      int
	MaxFail     int `toml:"max_fail"`
	FailTimeout int `toml:"fail_timeout"`
}

// Server represents server config resource
type Server struct {
	// TODO - Add SSL cert and keyfile
	Port      int
	Locations []ServerLocation
}

// ServerLocation represents location config resource
type ServerLocation struct {
	Path     string
	HTTPPass string `toml:"http_pass"`
}

func (cfg *Config) validate() error {
	if cfg.Upstreams == nil || len(cfg.Upstreams) == 0 {
		return errNoUpstreams
	}

	for _, ups := range cfg.Upstreams {
		cfg.setUpstreamDefaults(ups)
		if ups.Provider == StaticProvider &&
			(ups.Servers == nil || len(ups.Servers) == 0) {
			return errNoServers
		}
		for _, s := range ups.Servers {
			if s.Path == "" {
				return errNoServerPath
			}
		}
	}

	if cfg.Server == nil {
		return errNoServer
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = defaultPort
	}

	if cfg.Server.Locations == nil || len(cfg.Server.Locations) == 0 {
		return errNoServerLocations
	}

	for _, loc := range cfg.Server.Locations {
		if _, ok := cfg.Upstreams[loc.HTTPPass]; !ok {
			return errUpstreamMismatch
		}
	}

	return nil
}

func (cfg *Config) setUpstreamDefaults(u *Upstream) {
	if u.Provider == "" {
		u.Provider = StaticProvider
	}
	if u.Balancer == "" {
		u.Balancer = RoundRobinAlg
	}
	for _, s := range u.Servers {
		cfg.setUServerDefaults(s)
	}
}

func (*Config) setUServerDefaults(s *UpstreamServer) {
	if s.MaxFail == 0 {
		s.MaxFail = 10
	}
	if s.FailTimeout == 0 {
		s.FailTimeout = 1
	}
}
