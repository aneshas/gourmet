// Package compose provides gourmet app configurable initializaton
package compose

import (
	"time"

	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/config"
	"github.com/tonto/gourmet/internal/upstream"
)

// LocationBalancers represents location path regex map
// with their assigned balancer
type LocationBalancers map[string]balancer.Balancer

// FromConfig composes app from provided config
func FromConfig(cfg *config.Config) (LocationBalancers, error) {
	m := make(map[string]balancer.Balancer)

	for _, loc := range cfg.Server.Locations {
		ups := cfg.Upstreams[loc.HTTPPass]
		servers, err := getServers(ups)
		if err != nil {
			return nil, err
		}
		bl := makeBalancer(ups.Balancer, servers)
		m[loc.Path] = bl
	}

	return m, nil
}

func makeBalancer(alg string, s []*upstream.Server) balancer.Balancer {
	switch alg {
	case config.RoundRobinAlg:
		return balancer.NewRoundRobin(s)
	}
	return nil
}

func getServers(ups *config.Upstream) ([]*upstream.Server, error) {
	var servers []*upstream.Server

	switch ups.Provider {
	case config.StaticProvider:
		for _, s := range ups.Servers {
			servers = append(
				servers,
				upstream.NewServer(
					s.Path,
					upstream.WithWeight(s.Weight),
					upstream.WithFailTimeout(time.Duration(s.FailTimeout)*time.Second),
					upstream.WithMaxFail(s.MaxFail),
					upstream.WithQueueSize(100),
				),
			)
		}
	}

	return servers, nil
}
