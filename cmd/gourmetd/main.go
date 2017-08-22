package main

import (
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
	configFile = "/etc/gourmetd.conf"
)

func main() {
	r, err := os.Open("cmd/gourmetd/example.toml")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.New(r)
	if err != nil {
		log.Fatal(err)
	}

	ig := ingres.New()

	for _, loc := range cfg.Server.Locations {
		ups := cfg.Upstreams[loc.Upstream]

		// check provider and fetch servers
		// if provider is static use servers provided in config
		// keep servers in sync using the provider
		var servers []*upstream.Server

		for _, s := range ups.Servers {
			servers = append(servers, upstream.NewServer(s.Path, s.Weight))
		}

		log.Println(servers)
		// instantiate the correct balancer
		bl := balancer.NewRoundRobin(servers)

		ig.AddLocation(loc.Location, location.New(bl))
		log.Println(loc.Location)
	}

	// Config flow
	// - determine what type of provider we are using
	// - instantiate the provider or just use provided servers if static
	// - get servers for upstreams from provider and spin up servers
	// - loop through upstreams create balancers and map them to location urls eg. map[string]balancer.Balancer
	// - add locations to ingres using mapped urls and balancers

	logger := log.New(os.Stdout, "gourmet => ", log.Ldate|log.Ltime|log.Lshortfile)

	sv := http.NewServer(
		http.WithHandler(ig),
		http.WithLogger(logger),
		http.WithMiddleware(
			middleware.CORS(),
		),
	)

	log.Fatal(sv.Run(cfg.Server.Port))
}
