package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"jemgunay/url-scraper/pkg/ports"
)

func TestStore_Store(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		sortBy         ports.SortBy
		sortOrder      ports.SortOrder
		expectedKeys   []string
		expectedCounts []int
	}{
		{
			name:           "sort by age descending",
			limit:          5,
			sortBy:         ports.Age,
			sortOrder:      ports.Descending,
			expectedKeys:   []string{"url-10", "url-9", "url-8", "url-7", "url-6"},
			expectedCounts: []int{10, 9, 8, 7, 6}, // 1/2/3/4/5 have been expired
		},
		{
			name:           "sort by age ascending",
			limit:          5,
			sortBy:         ports.Age,
			sortOrder:      ports.Ascending,
			expectedKeys:   []string{"url-6", "url-7", "url-8", "url-9", "url-10"},
			expectedCounts: []int{6, 7, 8, 9, 10},
		},
	}

	logger := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(logger, 5)

			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("url-%d", i+1)

				for j := 0; j < i+1; j++ {
					s.Store(key)
				}

				time.Sleep(time.Millisecond * 10)
			}

			actualRecords := s.Fetch(tt.limit, tt.sortBy, tt.sortOrder)

			var actualKeys []string
			var actualCounts []int
			for _, k := range actualRecords {
				actualKeys = append(actualKeys, k.Key)
				actualCounts = append(actualCounts, k.SubmitCount)
			}
			require.Equal(t, tt.expectedKeys, actualKeys)
			require.Equal(t, tt.expectedCounts, actualCounts)
		})
	}
}
