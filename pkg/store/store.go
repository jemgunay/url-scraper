package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

var _ ports.Storer = (*Store)(nil)

// Store is a concurrency-safe store with key counting functionality. If the
// number of stored values exceeds the defined store capacity, the store's
// dataset is truncated.
type Store struct {
	logger         config.Logger
	recordCapacity int

	mu           *sync.RWMutex
	recordLookup map[string]*ports.Record
	recordList   []*ports.Record
}

// New initialises a new Store ready to be read from/written to.
func New(logger config.Logger, recordCapacity int) *Store {
	return &Store{
		logger:         logger,
		recordCapacity: recordCapacity,

		mu:           &sync.RWMutex{},
		recordLookup: make(map[string]*ports.Record, recordCapacity),
		// capacity+1 as we will append one slot over capacity before truncating
		recordList: make([]*ports.Record, 0, recordCapacity+1),
	}
}

// Store stores a key into the store, or bumps the count if it has been
// previously stored. Store is concurrency safe.
func (s *Store) Store(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// store in uniqueness map if previously unseen, or bump existing entry
	// so that we don't have to search the ports.Record list
	val, ok := s.recordLookup[key]
	if !ok {
		val = &ports.Record{
			Key: key,
		}
	}
	val.LastUpserted = time.Now().UTC()
	val.SubmitCount++
	s.recordLookup[key] = val

	if ok {
		// if we've seen the value before, we don't want to re-append it to
		// the record list as it's already in there
		return
	}

	// append to key list, sort by upsert date, and truncate to maximum
	// capacity; if at capacity, we'll truncate the oldest ports.Record
	s.recordList = append(s.recordList, val)

	// sort by last upserted timestamp
	sort.Slice(s.recordList, func(i, j int) bool {
		return s.recordList[i].LastUpserted.UnixNano() > s.recordList[j].LastUpserted.UnixNano()
	})

	if len(s.recordList) > s.recordCapacity {
		// remove purged item from lookup
		recordToPurge := s.recordList[s.recordCapacity]
		delete(s.recordLookup, recordToPurge.Key)
		// truncate slice to purge item from list
		s.recordList = s.recordList[:s.recordCapacity]
	}

	fmt.Println(len(s.recordList), len(s.recordLookup))
}

// Fetch fetches records given the specified criteria. Records are truncated to
// the required limit, and are sorted as requested by sortBy and sortOrder.
func (s *Store) Fetch(limit int, sortBy ports.SortBy, sortOrder ports.SortOrder) []ports.Record {
	s.mu.RLock()

	// copy so that the consumer doesn't mutate the original store list
	recordListCopy := make([]ports.Record, 0, len(s.recordList))
	for _, record := range s.recordList {
		recordListCopy = append(recordListCopy, *record)
	}

	s.mu.RUnlock()

	if limit > s.recordCapacity {
		limit = s.recordCapacity
	}

	// by default, records will be sorted by age descending on insertion
	if sortBy == ports.Count {
		sort.Slice(recordListCopy, func(i, j int) bool {
			return recordListCopy[i].SubmitCount > recordListCopy[j].SubmitCount
		})
	}

	if sortOrder == ports.Ascending {
		// reverse slice ordering
		for i, j := 0, len(recordListCopy)-1; i < j; i, j = i+1, j-1 {
			recordListCopy[i], recordListCopy[j] = recordListCopy[j], recordListCopy[i]
		}
	}

	return recordListCopy[:limit]
}
