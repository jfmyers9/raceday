package nascar

import (
	"fmt"
	"sort"

	"github.com/jfmyers/tmux-raceday/internal/series"
)

// NASCARSeries implements series.Series for the NASCAR Cup Series.
type NASCARSeries struct{}

func NewSeries() *NASCARSeries { return &NASCARSeries{} }

func (s *NASCARSeries) Name() string      { return "NASCAR Cup" }
func (s *NASCARSeries) ShortName() string  { return "NASCAR" }

func (s *NASCARSeries) FetchSchedule(year int) ([]series.Race, error) {
	races, err := FetchCupSchedule(year)
	if err != nil {
		return nil, fmt.Errorf("nascar schedule: %w", err)
	}

	out := make([]series.Race, 0, len(races))
	for _, r := range races {
		sr := series.Race{
			SeriesName:  s.Name(),
			ShortName:   s.ShortName(),
			RaceName:    r.RaceName,
			TrackName:   r.TrackName,
			Broadcaster: r.TelevisionBroadcaster,
			Complete:    r.IsComplete(),
		}
		if t, err := r.RaceStartUTC(); err == nil {
			sr.StartTime = t
		}
		if lat, lon, ok := TrackCoords(r.TrackID); ok {
			sr.Lat, sr.Lon = lat, lon
		}
		out = append(out, sr)
	}
	return out, nil
}

func (s *NASCARSeries) FetchLiveState() (*series.LiveState, error) {
	feed, err := FetchLiveFeed()
	if err != nil {
		return nil, nil
	}
	if !feed.IsLiveCupRace() {
		return nil, nil
	}

	state := &series.LiveState{
		SeriesName: s.Name(),
		ShortName:  s.ShortName(),
		RaceName:   feed.RunName,
		TrackName:  feed.TrackName,
		CurrentLap: feed.LapNumber,
		TotalLaps:  feed.LapsInRace,
		FlagSymbol: FlagSymbol(feed.FlagState),
		FlagName:   flagName(feed.FlagState),
	}

	if lat, lon, ok := TrackCoords(feed.TrackID); ok {
		state.Lat, state.Lon = lat, lon
	}

	if leader := feed.Leader(); leader != nil {
		state.Leader = vehicleToDriver(leader)
	}

	sorted := make([]Vehicle, len(feed.Vehicles))
	copy(sorted, feed.Vehicles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RunningPosition < sorted[j].RunningPosition
	})
	state.Positions = make([]series.Driver, len(sorted))
	for i := range sorted {
		state.Positions[i] = vehicleToDriver(&sorted[i])
	}

	return state, nil
}

func vehicleToDriver(v *Vehicle) series.Driver {
	return series.Driver{
		Number:   v.VehicleNumber,
		Name:     v.Driver.LastName,
		FullName: v.Driver.FullName,
		Position: v.RunningPosition,
		Delta:    float64(v.RunningPosition - v.StartingPosition),
	}
}

func flagName(state int) string {
	switch state {
	case FlagGreen:
		return "Green"
	case FlagCaution:
		return "Caution"
	case FlagRed:
		return "Red"
	case FlagWhite:
		return "White"
	case FlagFinished:
		return "Checkered"
	default:
		return "Unknown"
	}
}
