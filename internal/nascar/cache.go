package nascar

import (
	"encoding/json"
	"time"

	"github.com/jfmyers/tmux-raceday/internal/cache"
)

const cacheTTL = 1 * time.Hour

var fileCache = cache.New("")

func fetchLiveFeedCached(nextRaceStart time.Time) (*LiveFeed, error) {
	const key = "live_feed.json"

	ttl := cache.TTLForProximity(nextRaceStart)
	if ttl > 0 {
		if data, ok := fileCache.Read(key, ttl); ok {
			var feed LiveFeed
			if err := json.Unmarshal(data, &feed); err == nil {
				return &feed, nil
			}
		}
	}

	feed, err := FetchLiveFeed()
	if err != nil {
		// Fall back to stale cache on API failure.
		if data, _ := fileCache.ReadStale(key, ttl); data != nil {
			var stale LiveFeed
			if json.Unmarshal(data, &stale) == nil {
				return &stale, nil
			}
		}
		return nil, err
	}

	if data, err := json.Marshal(feed); err == nil {
		_ = fileCache.Write(key, data)
	}

	return feed, nil
}
