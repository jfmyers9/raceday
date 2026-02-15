package nascar

import "time"

type ScheduleResponse struct {
	Series1 []Race `json:"series_1"`
}

type Race struct {
	RaceID                int             `json:"race_id"`
	SeriesID              int             `json:"series_id"`
	RaceSeason            int             `json:"race_season"`
	RaceName              string          `json:"race_name"`
	RaceTypeID            int             `json:"race_type_id"`
	TrackName             string          `json:"track_name"`
	DateScheduled         string          `json:"date_scheduled"`
	ScheduledLaps         int             `json:"scheduled_laps"`
	WinnerDriverID        *int            `json:"winner_driver_id"`
	TelevisionBroadcaster string          `json:"television_broadcaster"`
	Schedule              []ScheduleEvent `json:"schedule"`
}

type ScheduleEvent struct {
	EventName    string `json:"event_name"`
	Notes        string `json:"notes"`
	StartTimeUTC string `json:"start_time_utc"`
	RunType      int    `json:"run_type"`
}

// RaceStartUTC returns the UTC start time of the actual race (run_type=3).
// Falls back to date_scheduled if no race event is found.
func (r Race) RaceStartUTC() (time.Time, error) {
	for _, ev := range r.Schedule {
		if ev.RunType == 3 && ev.StartTimeUTC != "" {
			return time.Parse("2006-01-02T15:04:05", ev.StartTimeUTC)
		}
	}
	return time.Parse("2006-01-02T15:04:05", r.DateScheduled)
}

// IsComplete returns true if the race has a recorded winner.
func (r Race) IsComplete() bool {
	return r.WinnerDriverID != nil && *r.WinnerDriverID > 0
}
