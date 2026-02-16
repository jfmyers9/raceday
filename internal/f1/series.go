package f1

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jfmyers/tmux-raceday/internal/series"
)

// F1Series implements series.Series for Formula 1.
type F1Series struct{}

func NewSeries() *F1Series { return &F1Series{} }

func (s *F1Series) Name() string      { return "Formula 1" }
func (s *F1Series) ShortName() string  { return "F1" }

func (s *F1Series) FetchSchedule(year int) ([]series.Race, error) {
	meetings, err := FetchMeetings(year)
	if err != nil {
		return nil, fmt.Errorf("f1 meetings: %w", err)
	}
	sessions, err := FetchRaceSessions(year)
	if err != nil {
		return nil, fmt.Errorf("f1 race sessions: %w", err)
	}

	meetingByKey := make(map[int]Meeting, len(meetings))
	for _, m := range meetings {
		meetingByKey[m.MeetingKey] = m
	}

	now := time.Now()
	var races []series.Race
	for _, sess := range sessions {
		m, ok := meetingByKey[sess.MeetingKey]
		if !ok {
			continue
		}

		startTime, err := time.Parse(time.RFC3339, sess.DateStart)
		if err != nil {
			continue
		}

		endTime, _ := time.Parse(time.RFC3339, sess.DateEnd)
		complete := !endTime.IsZero() && endTime.Before(now)

		raceName := m.MeetingName
		if !strings.Contains(strings.ToLower(raceName), "grand prix") {
			raceName += " Grand Prix"
		}

		lat, lon, _ := CircuitCoords(m.Location)

		races = append(races, series.Race{
			SeriesName: s.Name(),
			ShortName:  s.ShortName(),
			RaceName:   raceName,
			TrackName:  m.CircuitShortName,
			StartTime:  startTime,
			Complete:   complete,
			Lat:        lat,
			Lon:        lon,
		})
	}
	return races, nil
}

func (s *F1Series) FetchLiveState() (*series.LiveState, error) {
	sess, err := FetchLatestSession()
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.SessionName != "Race" {
		return nil, nil
	}

	endTime, _ := time.Parse(time.RFC3339, sess.DateEnd)
	if !endTime.IsZero() && endTime.Before(time.Now()) {
		return nil, nil
	}

	positions, err := FetchPositions(sess.SessionKey)
	if err != nil {
		return nil, err
	}
	drivers, err := FetchDrivers(sess.SessionKey)
	if err != nil {
		return nil, err
	}
	rcMsgs, err := FetchRaceControl(sess.SessionKey)
	if err != nil {
		return nil, err
	}
	stints, _ := FetchStints(sess.SessionKey)

	// Build map of current tire compound per driver.
	// Stints arrive in order; last entry per driver is the current stint.
	compoundByDriver := make(map[int]string)
	for _, st := range stints {
		compoundByDriver[st.DriverNumber] = st.Compound
	}

	driverMap := make(map[int]DriverInfo, len(drivers))
	for _, d := range drivers {
		driverMap[d.DriverNumber] = d
	}

	// Deduplicate positions: keep only the latest entry per driver.
	latestPos := make(map[int]Position)
	for _, p := range positions {
		if existing, ok := latestPos[p.DriverNumber]; !ok || p.Date > existing.Date {
			latestPos[p.DriverNumber] = p
		}
	}

	// Build sorted driver list.
	var sorted []Position
	for _, p := range latestPos {
		sorted = append(sorted, p)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	var driverList []series.Driver
	for _, p := range sorted {
		d := driverMap[p.DriverNumber]
		driverList = append(driverList, series.Driver{
			Number:   fmt.Sprintf("%d", p.DriverNumber),
			Name:     d.NameAcronym,
			FullName: d.FullName,
			Team:     d.TeamName,
			Position: p.Position,
			Compound: compoundByDriver[p.DriverNumber],
		})
	}

	// Determine flag state from last race control flag message.
	flagSymbol, flagName := "", ""
	currentLap := 0
	for i := len(rcMsgs) - 1; i >= 0; i-- {
		msg := rcMsgs[i]
		if msg.LapNumber > currentLap {
			currentLap = msg.LapNumber
		}
		if flagSymbol == "" && msg.Category == "Flag" {
			flagSymbol, flagName = mapFlag(msg)
		}
	}

	// Fallback: highest lap from any race control message.
	for _, msg := range rcMsgs {
		if msg.LapNumber > currentLap {
			currentLap = msg.LapNumber
		}
	}

	lat, lon, _ := CircuitCoords(sess.Location)

	var leader series.Driver
	if len(driverList) > 0 {
		leader = driverList[0]
	}

	return &series.LiveState{
		SeriesName: s.Name(),
		ShortName:  s.ShortName(),
		RaceName:   sess.CircuitShortName,
		TrackName:  sess.CircuitShortName,
		CurrentLap: currentLap,
		TotalLaps:  0,
		FlagSymbol: flagSymbol,
		FlagName:   flagName,
		Leader:     leader,
		Positions:  driverList,
		Lat:        lat,
		Lon:        lon,
	}, nil
}

func mapFlag(msg RaceControlMessage) (symbol, name string) {
	switch msg.Flag {
	case "GREEN":
		return "🟢", "GREEN"
	case "YELLOW":
		if strings.Contains(msg.Message, "VIRTUAL SAFETY CAR") {
			return "🟡", "VSC"
		}
		if strings.Contains(msg.Message, "SAFETY CAR") {
			return "🟡", "SAFETY CAR"
		}
		return "🟡", "YELLOW"
	case "RED":
		return "🔴", "RED"
	case "CHEQUERED":
		return "🏁", "CHEQUERED"
	default:
		return "", ""
	}
}
