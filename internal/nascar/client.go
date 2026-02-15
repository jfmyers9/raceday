package nascar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

const baseURL = "https://cf.nascar.com/cacher"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// FetchCupSchedule returns the Cup Series (series_id=1) race schedule for the
// given year. Results are served from a local file cache when fresh.
func FetchCupSchedule(year int) ([]Race, error) {
	cacheKey := fmt.Sprintf("schedule_%d.json", year)

	if data, ok := readCache(cacheKey); ok {
		return parseCupSchedule(data)
	}

	url := fmt.Sprintf("%s/%d/race_list_basic.json", baseURL, year)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schedule API returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	_ = writeCache(cacheKey, data)
	return parseCupSchedule(data)
}

func parseCupSchedule(data []byte) ([]Race, error) {
	var resp ScheduleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing schedule: %w", err)
	}

	sort.Slice(resp.Series1, func(i, j int) bool {
		return resp.Series1[i].DateScheduled < resp.Series1[j].DateScheduled
	})
	return resp.Series1, nil
}

// NextRace returns the next upcoming (incomplete) Cup Series race.
func NextRace(races []Race) *Race {
	now := time.Now().UTC()
	for i := range races {
		if races[i].IsComplete() {
			continue
		}
		start, err := races[i].RaceStartUTC()
		if err != nil {
			continue
		}
		// Include races that started up to 6 hours ago (may still be running)
		if start.After(now.Add(-6 * time.Hour)) {
			return &races[i]
		}
	}
	return nil
}
