package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jfmyers/tmux-raceday/internal/config"
	"github.com/jfmyers/tmux-raceday/internal/f1"
	"github.com/jfmyers/tmux-raceday/internal/nascar"
	"github.com/jfmyers/tmux-raceday/internal/series"
	"github.com/jfmyers/tmux-raceday/internal/ui"
	"github.com/jfmyers/tmux-raceday/internal/weather"
	"github.com/mattn/go-runewidth"
)

func main() {
	status := flag.Bool("status", false, "Print one-line status for tmux status bar")
	driver := flag.Int("driver", 0, "Favorite driver car number (overrides config)")
	noWeather := flag.Bool("no-weather", false, "Disable weather data")
	width := flag.Int("width", 0, "Fixed output width for status mode (0=use config)")
	marquee := flag.Bool("marquee", false, "Enable marquee scrolling for long status text")
	initCfg := flag.Bool("init-config", false, "Create default config file")
	flag.Parse()

	if *initCfg {
		config.EnsureDefault()
		fmt.Println("Config created at ~/.config/raceday/config.yaml")
		return
	}

	cfg := config.Load()
	if *noWeather {
		cfg.Weather = false
	}

	// Merge all driver numbers from config into flat list
	var drivers []int
	for _, name := range cfg.Series {
		drivers = append(drivers, cfg.Drivers[name]...)
	}
	// CLI --driver overrides config
	if *driver > 0 {
		drivers = []int{*driver}
	}

	if *status {
		w := *width
		if w == 0 {
			w = cfg.StatusWidth
		}
		m := *marquee || cfg.Marquee
		runStatus(cfg, drivers, w, m)
		return
	}

	// Full TUI mode — use first driver as primary favorite
	primary := 0
	if len(drivers) > 0 {
		primary = drivers[0]
	}
	m := ui.NewModel(primary)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "raceday: %v\n", err)
		os.Exit(1)
	}
}

// segment is a piece of the status bar with a priority.
// Lower priority number = higher importance (dropped last).
// Static segments are always visible; non-static ones rotate in the marquee.
type segment struct {
	text     string
	priority int  // 0=core, 1=your drivers, 2=leader, 3=weather
	static   bool // true = pinned visible, false = part of marquee rotation
}

func runStatus(cfg config.Config, drivers []int, width int, marquee bool) {
	var allSeries []series.Series
	for _, name := range cfg.Series {
		switch name {
		case "nascar":
			allSeries = append(allSeries, nascar.NewSeries())
		case "f1":
			allSeries = append(allSeries, f1.NewSeries())
		}
	}
	multiSeries := len(allSeries) > 1

	var segments []segment

	// Check each series for a live session.
	var liveState *series.LiveState
	for _, s := range allSeries {
		st, err := s.FetchLiveState()
		if err == nil && st != nil {
			liveState = st
			break
		}
	}

	if liveState != nil {
		segments = liveSegmentsFromState(liveState, drivers, multiSeries)
		if cfg.Weather {
			if ws := weatherSuffixFromCoords(liveState.Lat, liveState.Lon); ws != "" {
				segments = append(segments, segment{ws, 3, false})
			}
		}
	} else {
		race := series.NextRaceAcrossAll(allSeries, time.Now())
		if race == nil {
			fmt.Print("No upcoming races")
			return
		}

		primary := 0
		if len(drivers) > 0 {
			primary = drivers[0]
		}
		segments = scheduleSegmentsFromRace(race, primary, multiSeries)
		if cfg.Weather {
			if ws := weatherSuffixFromCoords(race.Lat, race.Lon); ws != "" {
				segments = append(segments, segment{ws, 3, false})
			}
		}
	}

	var s string
	if marquee && width > 0 {
		s = assembleHybrid(segments, width, cfg.MarqueeSpeed, cfg.MarqueeSeparator)
	} else {
		s = assembleSegments(segments, width)
		if width > 0 {
			s = ui.PadToWidth(s, width)
		}
	}
	fmt.Print(s)
}

// assembleSegments joins segments, dropping lowest-priority ones first
// if the result exceeds width. When width is 0, all segments are included.
func assembleSegments(segs []segment, width int) string {
	if width <= 0 {
		var parts []string
		for _, seg := range segs {
			parts = append(parts, seg.text)
		}
		return strings.Join(parts, "")
	}

	// Try including all segments; drop highest-priority-number first
	for maxPri := 3; maxPri >= 0; maxPri-- {
		var parts []string
		for _, seg := range segs {
			if seg.priority <= maxPri {
				parts = append(parts, seg.text)
			}
		}
		result := strings.Join(parts, "")
		if runewidth.StringWidth(result) <= width {
			return result
		}
		// If we still exceed width even at this priority level,
		// return what we have and let marquee/truncation handle it
		if maxPri == 0 {
			return result
		}
	}

	return ""
}

// assembleHybrid renders static segments in place and scrolls the rest via marquee.
// If static segments alone exceed width, lower-priority static segments are dropped.
func assembleHybrid(segs []segment, width, speed int, sep string) string {
	var statics, dynamics []segment
	for _, s := range segs {
		if s.static {
			statics = append(statics, s)
		} else {
			dynamics = append(dynamics, s)
		}
	}

	// Build static text, dropping lowest-priority statics if over budget
	staticText := joinSegments(statics)
	for maxPri := 3; runewidth.StringWidth(staticText) > width && maxPri >= 0; maxPri-- {
		var kept []segment
		for _, s := range statics {
			if s.priority <= maxPri {
				kept = append(kept, s)
			}
		}
		statics = kept
		staticText = joinSegments(statics)
	}

	spacer := " | "
	spacerWidth := runewidth.StringWidth(spacer)
	remaining := width - runewidth.StringWidth(staticText) - spacerWidth
	if remaining <= 0 || len(dynamics) == 0 {
		return ui.PadToWidth(staticText, width)
	}

	dynText := joinSegments(dynamics)
	marqueed := ui.MarqueeText(dynText, remaining, speed, sep)
	return staticText + spacer + marqueed
}

func joinSegments(segs []segment) string {
	var parts []string
	for _, s := range segs {
		parts = append(parts, s.text)
	}
	return strings.Join(parts, "")
}

func weatherSuffixFromCoords(lat, lon float64) string {
	if lat == 0 && lon == 0 {
		return ""
	}
	c, err := weather.FetchCurrent(lat, lon)
	if err != nil {
		return ""
	}
	return fmt.Sprintf(" | %.0f°F %s %.0fmph %s", c.Temp, weather.Symbol(c.WeatherCode), c.WindSpeed, weather.WindDirectionArrow(c.WindDirection))
}

func liveSegmentsFromState(state *series.LiveState, drivers []int, multiSeries bool) []segment {
	prefix := ""
	if multiSeries {
		prefix = state.ShortName + ": "
	}

	lapStr := fmt.Sprintf("Lap %d", state.CurrentLap)
	if state.TotalLaps > 0 {
		lapStr = fmt.Sprintf("Lap %d/%d", state.CurrentLap, state.TotalLaps)
	}

	flagPart := ""
	if state.FlagSymbol != "" {
		flagPart = state.FlagSymbol + " "
	}

	segs := []segment{
		{fmt.Sprintf("%s%s%s | %s", prefix, flagPart, state.RaceName, lapStr), 0, true},
	}

	if state.Leader.Number != "" {
		segs = append(segs, segment{
			fmt.Sprintf(" | P1 #%s %s", state.Leader.Number, state.Leader.Name), 2, false,
		})
	}

	for i, d := range drivers {
		carNum := strconv.Itoa(d)
		for _, p := range state.Positions {
			if p.Number == carNum {
				diffStr := ""
				if p.Delta > 0 {
					diffStr = fmt.Sprintf(" [+%d]", int(p.Delta))
				} else if p.Delta < 0 {
					diffStr = fmt.Sprintf(" [%d]", int(p.Delta))
				}
				segs = append(segs, segment{
					fmt.Sprintf(" | #%s %s P%d%s", p.Number, p.Name, p.Position, diffStr), 1, i == 0,
				})
				break
			}
		}
	}

	return segs
}

func scheduleSegmentsFromRace(race *series.Race, driverNum int, multiSeries bool) []segment {
	prefix := ""
	if multiSeries {
		prefix = race.ShortName + ": "
	}

	if race.StartTime.IsZero() {
		return []segment{{fmt.Sprintf("🏁 %s%s", prefix, race.RaceName), 0, true}}
	}

	local := race.StartTime.Local()
	now := time.Now()

	var timeStr string
	switch {
	case local.Year() == now.Year() && local.YearDay() == now.YearDay():
		timeStr = "Today " + local.Format("3:04 PM")
	case local.Year() == now.Year() && local.YearDay() == now.YearDay()+1:
		timeStr = "Tomorrow " + local.Format("3:04 PM")
	default:
		timeStr = local.Format("Jan 2 3:04 PM")
	}

	segs := []segment{
		{fmt.Sprintf("🏁 %s%s", prefix, race.RaceName), 0, true},
		{fmt.Sprintf(" | %s", timeStr), 0, false},
	}
	if race.Broadcaster != "" {
		segs = append(segs, segment{fmt.Sprintf(" | %s", race.Broadcaster), 2, false})
	}
	if driverNum > 0 {
		segs = append(segs, segment{fmt.Sprintf(" | #%d", driverNum), 1, true})
	}
	return segs
}
