// gourmetd is a gourmet binary that spawns
// an http server and balances requests recieved
package main

import (
	"flag"
	"log"
	"os"

	"github.com/tonto/gourmet/cmd/gourmetd/ingres"
	"github.com/tonto/gourmet/cmd/gourmetd/location"
	"github.com/tonto/gourmet/internal/balancer"
	"github.com/tonto/gourmet/internal/config"
	"github.com/tonto/gourmet/internal/upstream"
	"github.com/tonto/kit/http"
	"github.com/tonto/kit/http/middleware"
)

const (
	//configFile = "/etc/gourmetd.conf"
	configFile = "cmd/gourmetd/example.toml"
)

func main() {
	flag.Parse()
	cfile := flag.String("config", configFile, "path to configuration file")

	r, err := os.Open(*cfile)
	checkErr(err)

	cfg, err := config.New(r)
	checkErr(err)

	logger := log.New(os.Stdout, "gourmet => ", log.Ldate|log.Ltime|log.Lshortfile)

	sv := http.NewServer(
		http.WithHandler(mustMakeIg(cfg)),
		http.WithLogger(logger),
		http.WithMiddleware(
			middleware.CORS(),
		),
	)

	log.Fatal(sv.Run(cfg.Server.Port))
}

func mustMakeIg(cfg *config.Config) *ingres.RegExRouter {
	ig := ingres.NewRegEx()

	for _, loc := range cfg.Server.Locations {
		ups := cfg.Upstreams[loc.Upstream]
		ig.AddLocation(
			loc.RegEx,
			location.New(
				mustMakeBalancer(&ups),
			),
		)
	}

	return ig
}

func mustMakeBalancer(ups *config.Upstream) balancer.Balancer {
	s := mustGetSrvrs(ups)

	switch ups.Balancer {
	case config.RoundRobinAlg:
		return balancer.NewRoundRobin(s)
	}
	return nil
}

func mustGetSrvrs(ups *config.Upstream) []*upstream.Server {
	var servers []*upstream.Server

	switch ups.Provider {
	case config.StaticProvider:
		for _, s := range ups.Servers {
			servers = append(servers, upstream.NewServer(s.Path, s.Weight))
		}
	}

	return servers
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
