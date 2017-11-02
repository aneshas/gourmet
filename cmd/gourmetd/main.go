package main

import (
	"flag"
	"log"
	"os"

	"github.com/tonto/gourmet/internal/config"
	"github.com/tonto/gourmet/internal/platform/ingress"
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

	// TODO - config.MustParse(r, func(cfg *config.Config) {}) (apply this to stubbing?)
	cfg, err := config.Parse(r)
	checkErr(err)

	logger := log.New(os.Stdout, "gourmet => ", log.Ldate|log.Ltime)
	ig := ingress.New(logger)

	// TODO - Handle startup / gracefull shutdown better
	// eg. coordinate stop() with server shutdown
	// move server from kit to gourmet?
	stop := run(ig, cfg)
	defer stop()

	sv := http.NewServer(
		http.WithHandler(ig),
		http.WithLogger(logger),
		http.WithMiddleware(
			middleware.CORS(),
		),
	)

	err = sv.Run(cfg.Server.Port)
	if err != nil {
		logger.Println("error stopping server", err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
