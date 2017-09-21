// gourmetd is a gourmet binary that spawns
// an http server and balances requests received
package main

import (
	"flag"
	"log"
	"os"

	"github.com/tonto/gourmet/cmd/gourmetd/ingress"
	"github.com/tonto/gourmet/internal/compose"
	"github.com/tonto/gourmet/internal/config"
	"github.com/tonto/gourmet/internal/platform/protocol"
	"github.com/tonto/kit/http"
	"github.com/tonto/kit/http/middleware"
)

const (
	//configFile = "/etc/gourmet/gourmet.conf"
	configFile = "cmd/gourmetd/example.toml"
)

func main() {
	flag.Parse()
	cfile := flag.String("config", configFile, "path to configuration file")

	r, err := os.Open(*cfile)
	checkErr(err)

	cfg, err := config.Parse(r)
	checkErr(err)

	// TODO - make it recieve a func with regex and protocol params
	// rename it eg. With(func(string, http.Handler))
	bmap, err := compose.FromConfig(cfg)
	checkErr(err)

	logger := log.New(os.Stdout, "gourmet => ", log.Ldate|log.Ltime)

	ig := ingress.NewHTTP(logger)

	for rx, bl := range bmap {
		// TODO - determine type of protocol by looking at Protocol in location list
		ig.AddLocation(rx, protocol.NewHTTP(bl))
	}

	sv := http.NewServer(
		http.WithHandler(ig),
		http.WithLogger(logger),
		http.WithMiddleware(
			middleware.CORS(),
		),
	)

	log.Fatal(sv.Run(cfg.Server.Port))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
