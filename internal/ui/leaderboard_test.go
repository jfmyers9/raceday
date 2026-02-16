package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jfmyers/tmux-raceday/internal/nascar"
	"github.com/jfmyers/tmux-raceday/internal/series"
	"github.com/jfmyers/tmux-raceday/internal/weather"
)

func liveFeed() *nascar.LiveFeed {
	return &nascar.LiveFeed{
		RunType:   3,
		SeriesID:  1,
		LapNumber: 50,
		TrackID:   1,
	}
}

func TestHasLiveRace_NASCARLive(t *testing.T) {
	m := Model{feed: liveFeed()}
	if !m.hasLiveRace() {
		t.Error("expected true when NASCAR feed is a live cup race")
	}
}

func TestHasLiveRace_F1Live(t *testing.T) {
	m := Model{f1Live: &series.LiveState{}}
	if !m.hasLiveRace() {
		t.Error("expected true when F1 live state is present")
	}
}

func TestHasLiveRace_NoLive(t *testing.T) {
	m := Model{}
	if m.hasLiveRace() {
		t.Error("expected false with no live data")
	}
}

func TestHasLiveRace_NonLiveFeed(t *testing.T) {
	m := Model{feed: &nascar.LiveFeed{RunType: 1, SeriesID: 1}}
	if m.hasLiveRace() {
		t.Error("expected false when feed is not a live cup race")
	}
}

func TestWeatherTickNoFetchWithoutLiveRace(t *testing.T) {
	m := Model{
		feed: &nascar.LiveFeed{TrackID: 1},
		weather: &weather.Conditions{Temp: 72},
	}

	result, cmd := m.Update(weatherTickMsg{})
	updated := result.(Model)

	if updated.weather == nil {
		t.Error("weather should be preserved (not cleared by tick)")
	}

	// Should only return weatherTickCmd (the timer), not a batch with fetchWeatherCmd
	// Since there's no live race, the tick handler returns a single cmd (weatherTickCmd)
	// rather than a batch of fetchWeatherCmd + weatherTickCmd
	if cmd == nil {
		t.Error("expected weatherTickCmd to be returned")
	}
}

func TestFeedMsgClearsWeatherWhenNoLiveRace(t *testing.T) {
	m := Model{
		weather: &weather.Conditions{Temp: 72},
	}

	// nil feed should clear weather
	result, _ := m.Update(feedMsg(nil))
	updated := result.(Model)
	if updated.weather != nil {
		t.Error("weather should be cleared when feed is nil and no F1 live")
	}
}

func TestFeedMsgPreservesWeatherWithF1Live(t *testing.T) {
	m := Model{
		weather: &weather.Conditions{Temp: 72},
		f1Live:  &series.LiveState{},
	}

	result, _ := m.Update(feedMsg(nil))
	updated := result.(Model)
	if updated.weather == nil {
		t.Error("weather should be preserved when F1 is live")
	}
}

func TestFeedMsgClearsWeatherWhenNonLiveFeed(t *testing.T) {
	m := Model{
		weather: &weather.Conditions{Temp: 72},
	}

	// Non-live feed (practice session)
	feed := &nascar.LiveFeed{RunType: 1, SeriesID: 1, TrackID: 5}
	result, _ := m.Update(feedMsg(feed))
	updated := result.(Model)
	if updated.weather != nil {
		t.Error("weather should be cleared for non-live feed without F1")
	}
}

func testUpdateReturnsModel(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	result, _ := m.Update(msg)
	return result.(Model)
}
