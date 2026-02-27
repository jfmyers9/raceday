package cache

import (
	"os"
	"path/filepath"
	"time"
)

// Cache provides file-based caching with mtime-based TTL.
type Cache struct {
	dir string
}

// New creates a Cache that stores files under $UserCacheDir/raceday/{subdir}.
func New(subdir string) *Cache {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return &Cache{dir: filepath.Join(base, "raceday", subdir)}
}

// Read returns cached data if it exists and is younger than ttl.
func (c *Cache) Read(key string, ttl time.Duration) ([]byte, bool) {
	path := filepath.Join(c.dir, key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > ttl {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Write stores data to the cache file.
func (c *Cache) Write(key string, data []byte) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, key), data, 0o644)
}

// ReadStale returns cached data regardless of TTL, as long as the file exists.
// The second return value indicates whether the data was stale (TTL expired).
func (c *Cache) ReadStale(key string, ttl time.Duration) ([]byte, bool) {
	path := filepath.Join(c.dir, key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	stale := time.Since(info.ModTime()) > ttl
	return data, stale
}

// Invalidate removes a cache entry.
func (c *Cache) Invalidate(key string) {
	_ = os.Remove(filepath.Join(c.dir, key))
}

// TimeNow is a seam for testing time-dependent behavior.
var TimeNow = time.Now

// TTLForProximity returns a cache duration based on how close now is to
// the given event start time. Closer events get shorter (or zero) TTLs
// so data stays fresh when it matters most.
func TTLForProximity(eventStart time.Time) time.Duration {
	now := TimeNow()

	// Event in progress (within 6h of start) — no caching.
	sinceStart := now.Sub(eventStart)
	if sinceStart >= 0 && sinceStart < 6*time.Hour {
		return 0
	}

	untilStart := eventStart.Sub(now)

	// Starting within an hour — very short cache.
	if untilStart > 0 && untilStart <= time.Hour {
		return 30 * time.Second
	}

	// Same calendar day (UTC).
	y1, m1, d1 := now.UTC().Date()
	y2, m2, d2 := eventStart.UTC().Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		return 2 * time.Minute
	}

	// Within a week.
	if untilStart > 0 && untilStart <= 7*24*time.Hour {
		return 10 * time.Minute
	}

	return 30 * time.Minute
}
