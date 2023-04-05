package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"jemgunay/url-scraper/pkg/ports"
)

type testIngester struct{}

func (testIngester) Ingest(ctx context.Context, url string) error {
	return nil
}

type testStorage struct{}

func (testStorage) Store(key string) {}

func (testStorage) Fetch(limit int, sortBy ports.SortBy, order ports.SortOrder) []ports.Record {
	return []ports.Record{}
}

func TestNew_InvalidPort(t *testing.T) {
	logger := zap.NewNop()
	ingester := &testIngester{}
	storage := testStorage{}

	server := New(logger, -1, ingester, storage)
	err := server.Run()
	require.ErrorContains(t, err, "listen tcp: address -1: invalid port")
}
