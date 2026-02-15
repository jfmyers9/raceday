package nascar

import (
	"testing"
)

func TestTrackCoords_KnownTrack(t *testing.T) {
	lat, lon, ok := TrackCoords(105) // Daytona
	if !ok {
		t.Fatal("expected Daytona (105) to have coordinates")
	}
	if lat < 29.0 || lat > 30.0 || lon > -80.0 || lon < -82.0 {
		t.Errorf("Daytona coords out of range: %f, %f", lat, lon)
	}
}

func TestTrackCoords_UnknownTrack(t *testing.T) {
	_, _, ok := TrackCoords(999999)
	if ok {
		t.Error("expected unknown track_id to return ok=false")
	}
}

func TestAllScheduleTracksHaveCoords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	races, err := FetchCupSchedule(2025)
	if err != nil {
		t.Fatalf("fetching 2025 schedule: %v", err)
	}
	if len(races) == 0 {
		t.Fatal("schedule returned 0 races")
	}

	seen := make(map[int]string)
	for _, r := range races {
		if r.TrackID == 0 {
			t.Errorf("race %q has track_id=0", r.RaceName)
			continue
		}
		seen[r.TrackID] = r.TrackName
	}

	for trackID, trackName := range seen {
		if _, _, ok := TrackCoords(trackID); !ok {
			t.Errorf("track_id %d (%s) missing from trackLocations", trackID, trackName)
		}
	}
}
