package nascar

import "testing"

func TestIsLiveCupRace(t *testing.T) {
	tests := []struct {
		name     string
		feed     LiveFeed
		want     bool
	}{
		{
			"green flag race",
			LiveFeed{RunType: 3, SeriesID: 1, LapNumber: 50, FlagState: FlagGreen},
			true,
		},
		{
			"finished race still reports live",
			LiveFeed{RunType: 3, SeriesID: 1, LapNumber: 200, FlagState: FlagFinished},
			true,
		},
		{
			"caution flag",
			LiveFeed{RunType: 3, SeriesID: 1, LapNumber: 100, FlagState: FlagCaution},
			true,
		},
		{
			"not a race (practice)",
			LiveFeed{RunType: 1, SeriesID: 1, LapNumber: 10, FlagState: FlagGreen},
			false,
		},
		{
			"not cup series",
			LiveFeed{RunType: 3, SeriesID: 2, LapNumber: 10, FlagState: FlagGreen},
			false,
		},
		{
			"lap 0 (pre-race)",
			LiveFeed{RunType: 3, SeriesID: 1, LapNumber: 0, FlagState: FlagGreen},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.IsLiveCupRace()
			if got != tt.want {
				t.Errorf("IsLiveCupRace() = %v, want %v", got, tt.want)
			}
		})
	}
}
