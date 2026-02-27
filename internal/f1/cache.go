package f1

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jfmyers/tmux-raceday/internal/cache"
	"github.com/jfmyers/tmux-raceday/internal/series"
)

const cacheTTL = 1 * time.Hour

var fileCache = cache.New("f1")

// sessionTTL returns cache duration for the latest-session endpoint.
// It layers F1-specific end-time awareness on top of the shared
// proximity-based TTL logic.
func sessionTTL(sess *Session) time.Duration {
	if sess == nil {
		return 10 * time.Minute
	}

	startTime, err := time.Parse(time.RFC3339, sess.DateStart)
	if err != nil {
		return 10 * time.Minute
	}

	now := timeNow()
	endTime, _ := time.Parse(time.RFC3339, sess.DateEnd)
	pastEnd := !endTime.IsZero() && endTime.Before(now)

	// Past grace period — session is settled, use long TTL.
	if pastEnd && now.After(endTime.Add(series.PostRaceGracePeriod)) {
		return 10 * time.Minute
	}

	// Session started but not yet past grace period — no caching.
	if now.After(startTime) {
		return 0
	}

	return cache.TTLForProximity(startTime)
}

// sessionDataTTL returns cache duration for session sub-data
// (positions, drivers, race control, stints).
// Finished sessions have immutable data cached for longer.
func sessionDataTTL(sess *Session) time.Duration {
	if sess == nil {
		return 10 * time.Minute
	}

	now := timeNow()
	endTime, _ := time.Parse(time.RFC3339, sess.DateEnd)

	if !endTime.IsZero() && endTime.Before(now) {
		return time.Hour
	}

	return 0
}

// cachedFetch reads from cache if fresh, otherwise calls fetch and caches result.
// TTL of 0 bypasses cache reads (always fetches).
func cachedFetch[T any](cacheKey string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	if ttl > 0 {
		if data, ok := fileCache.Read(cacheKey, ttl); ok {
			var v T
			if json.Unmarshal(data, &v) == nil {
				return v, nil
			}
		}
	}

	v, err := fetch()
	if err != nil {
		// Fall back to stale cache on API failure.
		if data, _ := fileCache.ReadStale(cacheKey, ttl); data != nil {
			var stale T
			if json.Unmarshal(data, &stale) == nil {
				return stale, nil
			}
		}
		var zero T
		return zero, err
	}

	if data, err := json.Marshal(v); err == nil {
		_ = fileCache.Write(cacheKey, data)
	}
	return v, nil
}

func cachedFetchPositions(sess *Session) ([]Position, error) {
	return cachedFetch(
		fmt.Sprintf("positions_%d.json", sess.SessionKey),
		sessionDataTTL(sess),
		func() ([]Position, error) { return FetchPositions(sess.SessionKey) },
	)
}

func cachedFetchDrivers(sess *Session) ([]DriverInfo, error) {
	return cachedFetch(
		fmt.Sprintf("drivers_%d.json", sess.SessionKey),
		sessionDataTTL(sess),
		func() ([]DriverInfo, error) { return FetchDrivers(sess.SessionKey) },
	)
}

func cachedFetchRaceControl(sess *Session) ([]RaceControlMessage, error) {
	return cachedFetch(
		fmt.Sprintf("race_control_%d.json", sess.SessionKey),
		sessionDataTTL(sess),
		func() ([]RaceControlMessage, error) { return FetchRaceControl(sess.SessionKey) },
	)
}

func cachedFetchStints(sess *Session) ([]Stint, error) {
	return cachedFetch(
		fmt.Sprintf("stints_%d.json", sess.SessionKey),
		sessionDataTTL(sess),
		func() ([]Stint, error) { return FetchStints(sess.SessionKey) },
	)
}
