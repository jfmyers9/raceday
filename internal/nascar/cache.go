package nascar

import (
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 1 * time.Hour

func cacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "raceday")
}

func cachePath(key string) string {
	return filepath.Join(cacheDir(), key)
}

// readCache returns cached data if it exists and is younger than TTL.
func readCache(key string) ([]byte, bool) {
	path := cachePath(key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > cacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// invalidateCache removes a cache entry so the next fetch hits the API.
func invalidateCache(key string) {
	_ = os.Remove(cachePath(key))
}

// writeCache stores data to the cache file.
func writeCache(key string, data []byte) error {
	dir := cacheDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(cachePath(key), data, 0o644)
}
