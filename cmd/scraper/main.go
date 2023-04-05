package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

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
	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(conf.TimeoutSeconds),
	}
	ingester := ingest.New(logger, storage, httpClient)

	// start HTTP server
	logger.Info("starting HTTP server", zap.Int("port", conf.Port))
	httpServer := server.New(logger, conf.Port, ingester, storage)
	if err := httpServer.Run(); err != nil {
		logger.Warn("HTTP server shut down")
	}
}
