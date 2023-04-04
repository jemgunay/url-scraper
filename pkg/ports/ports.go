package ports

import "time"

type Record struct {
	Key          string
	SubmitCount  int
	LastUpserted time.Time
}

type Storer interface {
	Store(key string)
	Fetch() []Record
}

type Ingester interface {
}
