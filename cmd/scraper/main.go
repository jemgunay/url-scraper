package main

import (
	"flag"
	"log"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ingest"
	"jemgunay/url-scraper/pkg/server"
	"jemgunay/url-scraper/pkg/store"
)

func main() {
	confPath := flag.String("config-path", "config.yaml", "the path to the yaml config file")
	flag.Parse()

	conf, err := config.New(*confPath)
	if err != nil {
		log.Fatalf("failed to initialise config: %s", err)
	}

	logger := conf.Logger

	storage := store.New(logger, 50)

	ingester := ingest.New(logger, storage)

	// start HTTP server
	httpServer := server.New(logger, conf.Port, ingester, storage)
	if err := httpServer.Run(); err != nil {
		logger.Fatal("failed to initialise")
	}
}
