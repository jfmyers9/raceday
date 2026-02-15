package nascar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const pointsURL = "https://cf.nascar.com/live/feeds/live-points.json"

type PointsEntry struct {
	CarNumber        string `json:"car_number"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	DriverID         int    `json:"driver_id"`
	Points           int    `json:"points"`
	PointsPosition   int    `json:"points_position"`
	PointsThisRace   int    `json:"points_earned_this_race"`
	DeltaLeader      int    `json:"delta_leader"`
	DeltaNext        int    `json:"delta_next"`
	Wins             int    `json:"wins"`
	Top5             int    `json:"top_5"`
	Top10            int    `json:"top_10"`
	Poles            int    `json:"poles"`
	Stage1Points     int    `json:"stage_1_points"`
	Stage2Points     int    `json:"stage_2_points"`
	BonusPoints      int    `json:"bonus_points"`
	IsInChase        bool   `json:"is_in_chase"`
	IsPointsEligible bool   `json:"is_points_eligible"`
	IsRookie         bool   `json:"is_rookie"`
}

// FetchStandings retrieves the current points standings.
func FetchStandings() ([]PointsEntry, error) {
	resp, err := httpClient.Get(pointsURL)
	if err != nil {
		return nil, fmt.Errorf("fetching standings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("standings API returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading standings: %w", err)
	}

	var entries []PointsEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing standings: %w", err)
	}

	return entries, nil
}
