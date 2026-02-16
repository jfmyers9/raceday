package f1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.openf1.org/v1"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchJSON(url string, v any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openf1: %s returned %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// FetchMeetings returns all meetings for a given year.
func FetchMeetings(year int) ([]Meeting, error) {
	cacheKey := fmt.Sprintf("meetings_%d.json", year)
	if data, ok := readCache(cacheKey); ok {
		var m []Meeting
		if err := json.Unmarshal(data, &m); err == nil {
			return m, nil
		}
	}

	url := fmt.Sprintf("%s/meetings?year=%d", baseURL, year)
	var meetings []Meeting
	if err := fetchJSON(url, &meetings); err != nil {
		return nil, err
	}

	if data, err := json.Marshal(meetings); err == nil {
		writeCache(cacheKey, data)
	}
	return meetings, nil
}

// FetchRaceSessions returns race sessions for a given year.
func FetchRaceSessions(year int) ([]Session, error) {
	cacheKey := fmt.Sprintf("race_sessions_%d.json", year)
	if data, ok := readCache(cacheKey); ok {
		var s []Session
		if err := json.Unmarshal(data, &s); err == nil {
			return s, nil
		}
	}

	url := fmt.Sprintf("%s/sessions?year=%d&session_name=Race", baseURL, year)
	var sessions []Session
	if err := fetchJSON(url, &sessions); err != nil {
		return nil, err
	}

	if data, err := json.Marshal(sessions); err == nil {
		writeCache(cacheKey, data)
	}
	return sessions, nil
}

// FetchLatestSession returns the current or most recent session.
func FetchLatestSession() (*Session, error) {
	url := baseURL + "/sessions?session_key=latest"
	var sessions []Session
	if err := fetchJSON(url, &sessions); err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// FetchPositions returns position data for a session.
func FetchPositions(sessionKey int) ([]Position, error) {
	url := fmt.Sprintf("%s/position?session_key=%d", baseURL, sessionKey)
	var positions []Position
	if err := fetchJSON(url, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// FetchDrivers returns driver info for a session.
func FetchDrivers(sessionKey int) ([]DriverInfo, error) {
	url := fmt.Sprintf("%s/drivers?session_key=%d", baseURL, sessionKey)
	var drivers []DriverInfo
	if err := fetchJSON(url, &drivers); err != nil {
		return nil, err
	}
	return drivers, nil
}

// FetchRaceControl returns race control messages for a session.
func FetchRaceControl(sessionKey int) ([]RaceControlMessage, error) {
	url := fmt.Sprintf("%s/race_control?session_key=%d", baseURL, sessionKey)
	var msgs []RaceControlMessage
	if err := fetchJSON(url, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

// FetchStints returns stint data for a session.
func FetchStints(sessionKey int) ([]Stint, error) {
	url := fmt.Sprintf("%s/stints?session_key=%d", baseURL, sessionKey)
	var stints []Stint
	if err := fetchJSON(url, &stints); err != nil {
		return nil, err
	}
	return stints, nil
}
