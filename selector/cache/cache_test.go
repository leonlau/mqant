package cache

import (
	"testing"

	"github.com/leonlau/mqant/v2/registry/mock"
	"github.com/leonlau/mqant/v2/selector"
)

func TestCacheSelector(t *testing.T) {
	counts := map[string]int{}

	cache := NewSelector(selector.Registry(mock.NewRegistry()))

	next, err := cache.Select("foo")
	if err != nil {
		t.Errorf("Unexpected error calling cache select: %v", err)
	}

	for i := 0; i < 100; i++ {
		node, err := next()
		if err != nil {
			t.Errorf("Expected node err, got err: %v", err)
		}
		counts[node.Id]++
	}

	t.Logf("Cache Counts %v", counts)
}
