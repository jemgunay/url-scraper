package ports

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Record defines a URL record.
type Record struct {
	Key          string    `json:"key"`
	SubmitCount  int       `json:"count"`
	LastUpserted time.Time `json:"last_upserted"`
}

// SortOrder is the sort ordering approach.
type SortOrder string

// Validate validates SortOrder.
func (s SortOrder) Validate() error {
	switch s {
	case Ascending, Descending:
		return nil
	default:
		return errors.New("sort order value is invalid")
	}
}

// SortBy is the sort direction approach.
type SortBy string

// Validate validates SortBy.
func (s SortBy) Validate() error {
	switch s {
	case Age, Count:
		return nil
	default:
		return errors.New("sort by value is invalid")
	}
}

const (
	Ascending  SortOrder = "asc"
	Descending SortOrder = "desc"
	Age        SortBy    = "age"
	Count      SortBy    = "count"
)

// Storer is responsible for storing and fetching Records.
type Storer interface {
	Store(key string)
	Fetch(limit int, sortBy SortBy, order SortOrder) []Record
}

// Ingester is responsible for ingesting and processing URLs.
type Ingester interface {
	Ingest(ctx context.Context, url string) error
}

// Client represents a client capable of performing HTTP requests.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
