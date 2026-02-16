package f1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jfmyers/tmux-raceday/internal/series"
)

// stubF1Server returns an httptest.Server that responds to OpenF1 API calls
// with the provided session and race control messages. Positions and drivers
// are minimal but valid.
func stubF1Server(sess Session, rcMsgs []RaceControlMessage) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/sessions":
			json.NewEncoder(w).Encode([]Session{sess})
		case r.URL.Path == "/v1/position":
			json.NewEncoder(w).Encode([]Position{
				{DriverNumber: 1, Position: 1, Date: "2026-03-15T14:00:00"},
			})
		case r.URL.Path == "/v1/drivers":
			json.NewEncoder(w).Encode([]DriverInfo{
				{DriverNumber: 1, FullName: "Max Verstappen", NameAcronym: "VER", TeamName: "Red Bull"},
			})
		case r.URL.Path == "/v1/race_control":
			json.NewEncoder(w).Encode(rcMsgs)
		case r.URL.Path == "/v1/stints":
			json.NewEncoder(w).Encode([]Stint{})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestFetchLiveState_GracePeriod(t *testing.T) {
	raceStart := time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)
	raceEnd := raceStart.Add(2 * time.Hour)

	sess := Session{
		SessionKey:       9999,
		SessionName:      "Race",
		DateStart:        raceStart.Format(time.RFC3339),
		DateEnd:          raceEnd.Format(time.RFC3339),
		CircuitShortName: "Bahrain",
		Location:         "Bahrain",
	}

	chequered := []RaceControlMessage{
		{Category: "Flag", Flag: "GREEN", LapNumber: 1},
		{Category: "Flag", Flag: "CHEQUERED", LapNumber: 57},
	}

	tests := []struct {
		name         string
		now          time.Time
		rcMsgs       []RaceControlMessage
		wantNil      bool
		wantFinished bool
	}{
		{
			name:         "mid-race returns live state with Finished=false",
			now:          raceStart.Add(1 * time.Hour),
			rcMsgs:       []RaceControlMessage{{Category: "Flag", Flag: "GREEN", LapNumber: 30}},
			wantNil:      false,
			wantFinished: false,
		},
		{
			name:         "chequered flag sets Finished=true",
			now:          raceEnd.Add(-5 * time.Minute),
			rcMsgs:       chequered,
			wantNil:      false,
			wantFinished: true,
		},
		{
			name:         "just past end time still returns state (grace period)",
			now:          raceEnd.Add(10 * time.Minute),
			rcMsgs:       chequered,
			wantNil:      false,
			wantFinished: true,
		},
		{
			name:         "within grace period shows Finished=true",
			now:          raceEnd.Add(series.PostRaceGracePeriod - 1*time.Minute),
			rcMsgs:       chequered,
			wantNil:      false,
			wantFinished: true,
		},
		{
			name:    "after grace period returns nil",
			now:     raceEnd.Add(series.PostRaceGracePeriod + 1*time.Minute),
			rcMsgs:  chequered,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := stubF1Server(sess, tt.rcMsgs)
			defer srv.Close()

			origBase := baseURL
			origTimeNow := timeNow
			baseURL = srv.URL + "/v1"
			timeNow = func() time.Time { return tt.now }
			defer func() {
				baseURL = origBase
				timeNow = origTimeNow
			}()

			s := NewSeries()
			state, err := s.FetchLiveState()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if state != nil {
					t.Errorf("expected nil state, got %+v", state)
				}
				return
			}

			if state == nil {
				t.Fatal("expected non-nil state, got nil")
			}
			if state.Finished != tt.wantFinished {
				t.Errorf("Finished = %v, want %v", state.Finished, tt.wantFinished)
			}
		})
	}
}

func TestFetchLiveState_NotRaceSession(t *testing.T) {
	sess := Session{
		SessionKey:  9999,
		SessionName: "Qualifying",
		DateStart:   "2026-03-15T14:00:00Z",
	}

	srv := stubF1Server(sess, nil)
	defer srv.Close()

	origBase := baseURL
	baseURL = srv.URL + "/v1"
	defer func() { baseURL = origBase }()

	s := NewSeries()
	state, err := s.FetchLiveState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != nil {
		t.Errorf("expected nil for non-race session, got %+v", state)
	}
}
