package main

import (
	"time"

	"github.com/tonto/gourmet/internal/platform/ingress"

	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/config"
	"github.com/tonto/gourmet/internal/platform/protocol"
	"github.com/tonto/gourmet/internal/upstream"
)

func run(ig *ingress.Ingress, cfg *config.Config) func() {
	m := make(map[string]balancer.Balancer)
	qc := []chan struct{}{}

	for _, loc := range cfg.Server.Locations {
		ups := cfg.Upstreams[loc.HTTPPass]
		servers := getServers(ups)
		for _, s := range servers {
			c := make(chan struct{})
			qc = append(qc, c)
			go s.Run(c)
		}
		bl := getBalancer(ups.Balancer, servers)
		m[loc.Path] = bl
	}

	for path, h := range m {
		// TODO - determine type of protocol by looking at Protocol in location list
		ig.RegisterLocHandler(path, protocol.NewHTTP(h))
	}

	return func() {
		for _, c := range qc {
			c <- struct{}{}
		}
	}
}

func getBalancer(alg string, s []*upstream.Server) balancer.Balancer {
	switch alg {
	case config.RoundRobinAlg:
		return balancer.NewRoundRobin(s)
	case config.RandomAlg:
		return balancer.NewRandom(s)
	}
	return nil
}

func getServers(ups *config.Upstream) []*upstream.Server {
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

	return servers
}
