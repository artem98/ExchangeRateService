package worker

import (
	"sync"
	"time"
)

type RateJobsCache struct {
	mu       sync.Mutex
	requests map[string]cachedRequest
	ttl      time.Duration
}

type cachedRequest struct {
	id        uint64
	timestamp time.Time
}

func MakeRateJobsCache(ttl time.Duration) *RateJobsCache {
	return &RateJobsCache{
		requests: make(map[string]cachedRequest),
		ttl:      ttl,
	}
}

func (cache *RateJobsCache) Get(currency1, currency2 string) (uint64, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	code := currency1 + currency2

	entry, ok := cache.requests[code]
	if !ok {
		return 0, false
	}

	if time.Since(entry.timestamp) > cache.ttl {
		return 0, false
	}
	return entry.id, true
}

func (cache *RateJobsCache) Set(currency1, currency2 string, id uint64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.requests[currency1+currency2] = cachedRequest{
		id:        id,
		timestamp: time.Now(),
	}
}
