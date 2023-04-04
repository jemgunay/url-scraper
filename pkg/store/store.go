package store

import (
	"sort"
	"sync"
	"time"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

var _ ports.Storer = (*Store)(nil)

type Store struct {
	logger         config.Logger
	recordCapacity int

	mu           *sync.RWMutex
	recordLookup map[string]*ports.Record
	recordList   []*ports.Record
}

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
		return
	}

	// append to key list, sort by upsert date, and truncate to maximum
	// capacity; if at capacity, we'll truncate the oldest ports.Record
	s.recordList = append(s.recordList, val)

	sort.Slice(s.recordList, func(i, j int) bool {
		return s.recordList[i].LastUpserted.After(s.recordList[j].LastUpserted)
	})

	if len(s.recordList) > s.recordCapacity {
		// remove purged item from map
		recordToPurge := s.recordList[s.recordCapacity]
		delete(s.recordLookup, recordToPurge.Key)
		// truncate slice to purge item from list
		s.recordList = s.recordList[:s.recordCapacity]
	}
}

func (s *Store) Fetch() []ports.Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// copy so that the consumer doesn't mutate the original store list
	recordListCopy := make([]ports.Record, 0, len(s.recordList))
	for _, record := range s.recordList {
		recordListCopy = append(recordListCopy, *record)
	}

	// TODO: implement sort bits + limit opts

	return recordListCopy
}
