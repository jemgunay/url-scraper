package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// validate that old records are expired from the store when capacity is
// exceeded
func TestStore_StoreStaleEviction(t *testing.T) {
	logger := zap.NewNop()
	s := New(logger, 5)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("url-%d", i+1)

		for j := 0; j < i+1; j++ {
			s.Store(key)
		}

		time.Sleep(time.Millisecond * 10)
	}

	actualRecords := s.Fetch()

	expectedKeys := []string{"url-10", "url-9", "url-8", "url-7", "url-6"}
	expectedCounts := []int{10, 9, 8, 7, 6} // 1/2/3/4/5 have been expired

	var actualKeys []string
	var actualCounts []int
	for _, k := range actualRecords {
		actualKeys = append(actualKeys, k.Key)
		actualCounts = append(actualCounts, k.SubmitCount)
	}
	require.Equal(t, expectedKeys, actualKeys)
	require.Equal(t, expectedCounts, actualCounts)
}
