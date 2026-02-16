package f1

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
	return filepath.Join(dir, "raceday", "f1")
}

func readCache(key string) ([]byte, bool) {
	path := filepath.Join(cacheDir(), key)
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

func writeCache(key string, data []byte) error {
	dir := cacheDir()
	os.MkdirAll(dir, 0o755)
	return os.WriteFile(filepath.Join(dir, key), data, 0o644)
}
