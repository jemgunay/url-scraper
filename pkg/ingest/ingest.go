package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

var _ ports.Ingester = (*Processor)(nil)

type Processor struct {
	logger  config.Logger
	storage ports.Storer

	httpClient  ports.Client
	insertQueue chan string
}

func New(logger config.Logger, storage ports.Storer, httpClient ports.Client) *Processor {
	processor := &Processor{
		logger:     logger,
		storage:    storage,
		httpClient: httpClient,

		insertQueue: make(chan string, 10), // TODO: make capacity configurable via yaml
	}

	go processor.startPollers()

	return processor
}

func (s *Processor) startPollers() {
	f := func(url string) {
		logger := s.logger.With(zap.String("url", url))

		if _, err := s.benchmarkRequest(url); err != nil {
			// download failed so discard URL
			logger.Error("failed to validate URL", zap.Error(err))
			return
		}

		// URL is healthy so persist to store
		s.storage.Store(url)
		logger.Info("successfully validated and stored URL")
	}

	// create a long-lived worker group to fan out enqueued URL insertions
	newWorkerGroup[string](3, s.insertQueue, f)

	// continuously refresh benchmarks of the most common URLs
	ticker := time.NewTicker(time.Second * 60) // TODO: make ticker frequency configurable via yaml
	for range ticker.C {
		s.logger.Debug("triggering URL benchmark refresh")
		s.refreshBenchmarks()
	}
}

// Ingest attempts to ingest a URL into the Processor. If the processor queue is
// experiencing backpressure, the call will block until the context is
// cancelled, in which case an error is returned. It is the responsibility of
// the consumer to handle this error accordingly.
func (s *Processor) Ingest(ctx context.Context, url string) error {
	select {
	case s.insertQueue <- url:
		return nil
	case <-ctx.Done():
		return errors.New("request to enqueue expired")
	}
}

func (s *Processor) refreshBenchmarks() {
	// get 10 most submitted URLs from store and pre-queue them into a buffer
	records := s.storage.Fetch(10, ports.Age, ports.Descending)

	recordsIn := make(chan string, len(records))
	for _, record := range records {
		recordsIn <- record.Key
	}
	// close so that the worker group will terminate once all records have
	// drained
	close(recordsIn)

	// use resultsOut to fan results back in from the workers
	resultsOut := make(chan scrapeResult, len(records))

	// create worker pool of capacity 3 to fan out requests to benchmark URLs
	f := func(url string) {
		logger := s.logger.With(zap.String("url", url))

		result, err := s.benchmarkRequest(url)
		if err != nil {
			logger.Error("failed to benchmark URL", zap.Error(err))
		} else {
			s.storage.Store(url)
			logger.Info("successfully benchmarked URL", zap.Error(err))
		}

		resultsOut <- result
	}

	wg := newWorkerGroup[string](3, recordsIn, f)
	wg.Wait()

	// fan back in results & collect all the download times,
	summary := &scrapeSummary{}

	close(resultsOut)
	for result := range resultsOut {
		summary.push(result)
	}

	// log successful & failed download stats
	s.logger.Info("successfully refreshed URL benchmarks", zap.Any("summary", summary))
}

func (s *Processor) benchmarkRequest(url string) (scrapeResult, error) {
	result := scrapeResult{
		URL:    url,
		Status: failure,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// ensure we drain body and guarantee connection reuse
	io.Copy(io.Discard, resp.Body)

	// finish timing here so that we don't include validation in the benchmark
	result.Duration = time.Now().UTC().Sub(now).String()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected HTTP response status: %s", resp.Status)
	}

	result.Status = success
	return result, nil
}

type scrapeStatus string

const (
	success scrapeStatus = "success"
	failure scrapeStatus = "failure"
)

type scrapeResult struct {
	URL      string       `json:"url"`
	Duration string       `json:"duration"`
	Status   scrapeStatus `json:"status"`
}

type scrapeSummary struct {
	Durations    []scrapeResult `json:"scrape_durations"`
	SuccessCount int            `json:"success_count"`
	FailureCount int            `json:"failure_count"`
}

func (s *scrapeSummary) push(result scrapeResult) {
	if result.Status == success {
		s.SuccessCount++
	} else {
		s.FailureCount++
	}
	s.Durations = append(s.Durations, result)
}

// workerGroup facilitates the ability to fan out a stream of inputs to a
// predefined set of concurrent workers. The Wait method will wait for all
// workers to terminate. Worker termination is triggered by the closure and
// draining of the stream channel.
type workerGroup[T any] struct {
	stream <-chan T
	work   func(item T)
	wg     *sync.WaitGroup
}

// newWorkerGroup initialises a new workerGroup which reads from stream and fans
// out to "count" workers. Each worker will read from stream when unblocked and
// will execute the work function with this read value.
func newWorkerGroup[T any](count int, stream <-chan T, work func(item T)) *workerGroup[T] {
	w := &workerGroup[T]{
		stream: stream,
		work:   work,
		wg:     &sync.WaitGroup{},
	}

	w.wg.Add(count)
	for i := 0; i < count; i++ {
		go w.spawnWorker()
	}

	return w
}

func (w *workerGroup[T]) spawnWorker() {
	for item := range w.stream {
		w.work(item)
	}
	w.wg.Done()
}

// Wait blocks until all workers have terminated, i.e. stream has been closed
// and drained.
func (w *workerGroup[T]) Wait() {
	w.wg.Wait()
}
