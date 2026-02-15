package nascar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const liveFeedURL = "https://cf.nascar.com/live/feeds/live-feed.json"

// LiveFeed represents the real-time race data from NASCAR's CDN.
type LiveFeed struct {
	LapNumber              int       `json:"lap_number"`
	LapsInRace             int       `json:"laps_in_race"`
	LapsToGo               int       `json:"laps_to_go"`
	FlagState              int       `json:"flag_state"`
	RaceID                 int       `json:"race_id"`
	TrackID                int       `json:"track_id"`
	RunType                int       `json:"run_type"`
	SeriesID               int       `json:"series_id"`
	TrackName              string    `json:"track_name"`
	RunName                string    `json:"run_name"`
	Stage                  StageInfo `json:"stage"`
	NumberOfCautionLaps    int       `json:"number_of_caution_laps"`
	NumberOfCautions       int       `json:"number_of_caution_segments"`
	NumberOfLeadChanges    int       `json:"number_of_lead_changes"`
	NumberOfLeaders        int       `json:"number_of_leaders"`
	Vehicles               []Vehicle `json:"vehicles"`
}

type StageInfo struct {
	StageNum    int `json:"stage_num"`
	FinishAtLap int `json:"finish_at_lap"`
	LapsInStage int `json:"laps_in_stage"`
}

type Vehicle struct {
	RunningPosition     int        `json:"running_position"`
	VehicleNumber       string     `json:"vehicle_number"`
	Driver              DriverInfo `json:"driver"`
	Delta               float64    `json:"delta"`
	LapsCompleted       int        `json:"laps_completed"`
	LapsLed             []LapRange `json:"laps_led"`
	LastLapSpeed        float64    `json:"last_lap_speed"`
	BestLapSpeed        float64    `json:"best_lap_speed"`
	BestLapTime         float64    `json:"best_lap_time"`
	Status              int        `json:"status"`
	IsOnTrack           bool       `json:"is_on_track"`
	IsOnDVP             bool       `json:"is_on_dvp"`
	StartingPosition    int        `json:"starting_position"`
	VehicleManufacturer string     `json:"vehicle_manufacturer"`
	SponsorName         string     `json:"sponsor_name"`
	PassesMade          int        `json:"passes_made"`
	PassingDiff         int        `json:"passing_differential"`
	PitStops            []PitStop  `json:"pit_stops"`
}

type PitStop struct {
	PitInLapCount       int     `json:"pit_in_lap_count"`
	PitInRank           int     `json:"pit_in_rank"`
	PitOutRank          int     `json:"pit_out_rank"`
	PositionsGainedLost int     `json:"positions_gained_lossed"`
	PitInElapsedTime    float64 `json:"pit_in_elapsed_time"`
	PitOutElapsedTime   float64 `json:"pit_out_elapsed_time"`
}

// PitCount returns the number of pit stops with actual lap data.
func (v Vehicle) PitCount() int {
	count := 0
	for _, p := range v.PitStops {
		if p.PitInLapCount > 0 {
			count++
		}
	}
	return count
}

// LastPitLap returns the lap of the most recent pit stop, or 0 if none.
func (v Vehicle) LastPitLap() int {
	for i := len(v.PitStops) - 1; i >= 0; i-- {
		if v.PitStops[i].PitInLapCount > 0 {
			return v.PitStops[i].PitInLapCount
		}
	}
	return 0
}

// LapsSinceLastPit returns laps since last pit stop based on current lap.
func (v Vehicle) LapsSinceLastPit(currentLap int) int {
	last := v.LastPitLap()
	if last == 0 {
		return currentLap
	}
	return currentLap - last
}

type DriverInfo struct {
	DriverID  int    `json:"driver_id"`
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LapRange struct {
	StartLap int `json:"start_lap"`
	EndLap   int `json:"end_lap"`
}

// TotalLapsLed returns the total number of laps led by this vehicle.
func (v Vehicle) TotalLapsLed() int {
	total := 0
	for _, lr := range v.LapsLed {
		total += lr.EndLap - lr.StartLap + 1
	}
	return total
}

// FetchLiveFeed retrieves the current live race feed.
func FetchLiveFeed() (*LiveFeed, error) {
	resp, err := httpClient.Get(liveFeedURL)
	if err != nil {
		return nil, fmt.Errorf("fetching live feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("live feed returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading live feed: %w", err)
	}

	var feed LiveFeed
	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("parsing live feed: %w", err)
	}

	return &feed, nil
}

// Flag state constants matching NASCAR's API.
const (
	FlagGreen    = 1
	FlagCaution  = 2
	FlagRed      = 4
	FlagFinished = 8
	FlagWhite    = 9
)

// FlagSymbol returns a colored symbol for the current flag state.
func FlagSymbol(state int) string {
	switch state {
	case FlagGreen:
		return "🟢"
	case FlagCaution:
		return "🟡"
	case FlagRed:
		return "🔴"
	case FlagWhite:
		return "🏳"
	case FlagFinished:
		return "🏁"
	default:
		return "⚪"
	}
}

// IsLiveCupRace returns true if the feed represents a live Cup Series race.
func (f *LiveFeed) IsLiveCupRace() bool {
	return f.RunType == 3 && f.SeriesID == 1 && f.LapNumber > 0
}

// FindDriver returns the vehicle matching the given car number, or nil.
func (f *LiveFeed) FindDriver(carNumber string) *Vehicle {
	for i := range f.Vehicles {
		if f.Vehicles[i].VehicleNumber == carNumber {
			return &f.Vehicles[i]
		}
	}
	return nil
}

// Leader returns the vehicle in P1, or nil if no vehicles.
func (f *LiveFeed) Leader() *Vehicle {
	for i := range f.Vehicles {
		if f.Vehicles[i].RunningPosition == 1 {
			return &f.Vehicles[i]
		}
	}
	return nil
}
