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

func main() {
	configFile := flag.String("config", "/etc/gourmet/gourmet.toml", "path to configuration file")
	logFile := flag.String("log", "/var/log/gourmet/access.log", "path to log file")
	flag.Parse()

	r, err := os.Open(*configFile)
	checkErr(err)

	cfg, err := config.Parse(r)
	checkErr(err)

	err = os.MkdirAll("/var/log/gourmet", 0766)
	checkErr(err)

	file, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0766)
	checkErr(err)

	// TODO - Create app or gourmet type or package pass it config and logger
	// and it should do the stubing and running
	logger := log.New(file, "gourmet => ", log.Ldate|log.Ltime)
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
