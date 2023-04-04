package ports

import (
	"context"
	"time"
)

type Record struct {
	Key          string    `json:"key"`
	SubmitCount  int       `json:"count"`
	LastUpserted time.Time `json:"last_upserted"`
}

type Storer interface {
	Store(key string)
	Fetch() []Record
}

type Ingester interface {
	Ingest(ctx context.Context, url string) error
}
