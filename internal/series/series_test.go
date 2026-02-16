package series

import (
	"fmt"
	"testing"
	"time"
)

type mockSeries struct {
	name      string
	shortName string
	races     []Race
	live      *LiveState
	err       error
}

func (m *mockSeries) Name() string      { return m.name }
func (m *mockSeries) ShortName() string  { return m.shortName }
func (m *mockSeries) FetchSchedule(year int) ([]Race, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.races, nil
}
func (m *mockSeries) FetchLiveState() (*LiveState, error) {
	return m.live, m.err
}

func TestNextRaceAcrossAll(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	nascarRace := Race{
		SeriesName: "NASCAR Cup", ShortName: "NASCAR",
		RaceName: "Daytona 500", TrackName: "Daytona",
		StartTime: now.Add(24 * time.Hour),
	}
	f1Race := Race{
		SeriesName: "Formula 1", ShortName: "F1",
		RaceName: "Bahrain GP", TrackName: "Bahrain",
		StartTime: now.Add(48 * time.Hour),
	}

	nascar := &mockSeries{name: "NASCAR Cup", shortName: "NASCAR", races: []Race{nascarRace}}
	f1 := &mockSeries{name: "Formula 1", shortName: "F1", races: []Race{f1Race}}

	t.Run("NASCAR race soonest", func(t *testing.T) {
		got := NextRaceAcrossAll([]Series{nascar, f1}, now)
		if got == nil {
			t.Fatal("expected a race, got nil")
		}
		if got.RaceName != "Daytona 500" {
			t.Errorf("got %s, want Daytona 500", got.RaceName)
		}
	})

	t.Run("F1 race soonest", func(t *testing.T) {
		laterNascar := &mockSeries{
			name: "NASCAR Cup", shortName: "NASCAR",
			races: []Race{{
				SeriesName: "NASCAR Cup", RaceName: "Atlanta 400",
				StartTime: now.Add(72 * time.Hour),
			}},
		}
		got := NextRaceAcrossAll([]Series{laterNascar, f1}, now)
		if got == nil {
			t.Fatal("expected a race, got nil")
		}
		if got.RaceName != "Bahrain GP" {
			t.Errorf("got %s, want Bahrain GP", got.RaceName)
		}
	})

	t.Run("all complete returns nil", func(t *testing.T) {
		done := &mockSeries{
			name: "NASCAR Cup", shortName: "NASCAR",
			races: []Race{{Complete: true, StartTime: now.Add(time.Hour)}},
		}
		got := NextRaceAcrossAll([]Series{done}, now)
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("empty series list returns nil", func(t *testing.T) {
		got := NextRaceAcrossAll([]Series{}, now)
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("skips series with fetch errors", func(t *testing.T) {
		failing := &mockSeries{name: "Broken", err: fmt.Errorf("network error")}
		got := NextRaceAcrossAll([]Series{failing, f1}, now)
		if got == nil {
			t.Fatal("expected a race, got nil")
		}
		if got.RaceName != "Bahrain GP" {
			t.Errorf("got %s, want Bahrain GP", got.RaceName)
		}
	})

	t.Run("skips past races", func(t *testing.T) {
		pastOnly := &mockSeries{
			name: "NASCAR Cup", shortName: "NASCAR",
			races: []Race{{
				SeriesName: "NASCAR Cup", RaceName: "Past Race",
				StartTime: now.Add(-24 * time.Hour),
			}},
		}
		got := NextRaceAcrossAll([]Series{pastOnly}, now)
		if got != nil {
			t.Errorf("expected nil for past races, got %+v", got)
		}
	})
}
