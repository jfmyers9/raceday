package series

import "time"

// PostRaceGracePeriod is how long to keep displaying results after a race
// finishes. Both NASCAR and F1 use this to show final standings before the
// live state goes away.
const PostRaceGracePeriod = 90 * time.Minute

// Race represents a scheduled race from any series.
type Race struct {
	SeriesName  string
	ShortName   string
	RaceName    string
	TrackName   string
	StartTime   time.Time
	Broadcaster string
	Complete    bool
	Lat, Lon    float64
}

// Driver represents a driver's current position in a live session.
type Driver struct {
	Number   string // "24" for NASCAR, "1" for F1
	Name     string
	FullName string
	Team     string
	Position int
	Gap      string  // "+1.234" or "+1 LAP"
	Delta    float64 // delta from starting position
	Compound string  // tire compound: SOFT, MEDIUM, HARD, INTERMEDIATE, WET
}

// LiveState represents real-time session data from any series.
type LiveState struct {
	SeriesName string
	ShortName  string
	RaceName   string
	TrackName  string
	CurrentLap int
	TotalLaps  int // 0 if unknown (e.g. F1 timed sessions)
	FlagSymbol string
	FlagName   string
	Finished   bool
	Leader     Driver
	Positions  []Driver
	Lat, Lon   float64
}

// Series is the interface each racing series must implement.
type Series interface {
	Name() string
	ShortName() string
	FetchSchedule(year int) ([]Race, error)
	FetchLiveState() (*LiveState, error)
}

// NextRaceAcrossAll returns the soonest upcoming (incomplete) race
// across all provided series. Returns nil if none found.
func NextRaceAcrossAll(allSeries []Series, now time.Time) *Race {
	var best *Race
	for _, s := range allSeries {
		races, err := s.FetchSchedule(now.Year())
		if err != nil {
			continue
		}
		for i := range races {
			r := &races[i]
			if r.Complete {
				continue
			}
			if r.StartTime.Before(now) {
				continue
			}
			if best == nil || r.StartTime.Before(best.StartTime) {
				best = r
			}
		}
	}
	return best
}
