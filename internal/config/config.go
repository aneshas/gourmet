// Package config provides gourmet configuration parsing functionality
package config

import (
	"errors"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/upstream"
)

const (
	roundRobinAlg = "round_robin"
)

const (
	staticProvider = "static"
)

const (
	defaultPort = 8080
)

var (
	errUpstreamMismatch  = errors.New("upstreams in server location list don't match up with upstreams.")
	errNoUpstreams       = errors.New("upstream block missing or no upstreams listed.")
	errNoServers         = errors.New("if using static upstream server provider (default) server list should not be empty.")
	errNoServerPath      = errors.New("upstream server path must not be empty")
	errNoServer          = errors.New("server block not present")
	errNoServerLocations = errors.New("no server locations block present")
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

// New creates new config instance
func New(r io.Reader) (*Config, error) {
	cfg, err := parse(r)
	if err != nil {
		return nil, err
	}

	//cfg.setUpstreamDefaults()

	err = cfg.validate()
	if err != nil {
		return nil, err
	}

	return &Config{cfg.Server.Port, cfg}, nil
}

type LocationBalancers map[string]balancer.Balancer

// Config represents gourmet config struct and provides
// balancer instantiation methods
type Config struct {
	ServerPort int
	config     *config
}

func (cfg *Config) BalanceLocations() (LocationBalancers, error) {
	m := make(map[string]balancer.Balancer)

	for _, loc := range cfg.config.Server.Locations {
		ups := cfg.config.Upstreams[loc.Upstream]
		servers, err := getServers(ups)
		if err != nil {
			return nil, err
		}
		bl := makeBalancer(ups.Balancer, servers)
		m[loc.RegEx] = bl
	}

	return m, nil
}

func makeBalancer(alg string, s []*upstream.Server) balancer.Balancer {
	switch alg {
	case roundRobinAlg:
		return balancer.NewRoundRobin(s)
	}
	return nil
}

func getServers(ups *cfgUpstream) ([]*upstream.Server, error) {
	var servers []*upstream.Server

	switch ups.Provider {
	case staticProvider:
		for _, s := range ups.Servers {
			servers = append(servers, upstream.NewServer(s.Path, s.Weight))
		}
	}

	return servers, nil
}

type config struct {
	// TODO - Add SSL cert and keyfile
	Upstreams map[string]*cfgUpstream
	Server    *server
}

type cfgUpstream struct {
	Balancer string
	Provider string

	// Servers should be ignored if Provider is not static
	Servers []upstreamServer
}

type upstreamServer struct {
	Path   string
	Weight int
}

type server struct {
	Port      int
	Locations []serverLocation
}

type serverLocation struct {
	RegEx    string `toml:"location"`
	Upstream string
}

func (cfg *config) validate() error {
	if cfg.Upstreams == nil || len(cfg.Upstreams) == 0 {
		return errNoUpstreams
	}

	for _, ups := range cfg.Upstreams {
		cfg.setUpstreamDefaults(ups)
		if ups.Provider == staticProvider && (ups.Servers == nil || len(ups.Servers) == 0) {
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
		if _, ok := cfg.Upstreams[loc.Upstream]; !ok {
			return errUpstreamMismatch
		}
	}

	return nil
}

func (cfg *config) setUpstreamDefaults(u *cfgUpstream) {
	if u.Provider == "" {
		u.Provider = staticProvider
	}
	if u.Balancer == "" {
		u.Balancer = roundRobinAlg
	}
}

func parse(r io.Reader) (*config, error) {
	cfg := config{}
	_, err := toml.DecodeReader(r, &cfg)
	return &cfg, err
}
