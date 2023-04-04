package ingest

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

var _ ports.Ingester = (*Processor)(nil)

type Processor struct {
	logger  config.Logger
	storage ports.Storer

	httpClient  *http.Client // TODO: interface this?
	insertQueue chan string
}

func New(logger config.Logger, storage ports.Storer) *Processor {
	processor := &Processor{
		logger:  logger,
		storage: storage,

		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		insertQueue: make(chan string),
	}

	go processor.startPoller()

	return processor
}

func (s *Processor) startPoller() {
	ticker := time.NewTicker(time.Second * 60)

	for {
		select {
		case key := <-s.insertQueue:
			s.logger.Debug("inserting URL into store", zap.String("key", key))
			// attempt to download Key
			// if download fails, discard it
			// if success, persist in store\
			s.storage.Store(key)

		case <-ticker.C:
			s.logger.Debug("triggering URL refresh")
			// get 10 most submitted URLs from store

			// create worker pool of capacity 3
			// fan out and re-fetch each the URLs, measuring time to scrape

			// fan back in results & collect all the download times,
			// successful downloads counter and failed downloads
			// counter and log them all on the stdout
		}
	}
}

func (s *Processor) Ingest() {

}

type scrapeDuration struct {
	URL      string        `json:"url"`
	Duration time.Duration `json:"duration"`
}

type summary struct {
	Durations    []scrapeDuration `json:"scrape_durations"`
	SuccessCount int              `json:"success_count"`
	FailureCount int              `json:"failure_count"`
}