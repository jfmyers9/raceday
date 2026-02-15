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
	"github.com/jfmyers/tmux-raceday/internal/nascar"
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

	// CLI --driver overrides config
	drivers := cfg.Drivers
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
	var segments []segment

	live, err := nascar.FetchLiveFeed()
	if err == nil && live.IsLiveCupRace() {
		segments = liveSegments(live, drivers)
		if cfg.Weather {
			if ws := weatherSuffix(live.TrackID); ws != "" {
				segments = append(segments, segment{ws, 3, false})
			}
		}
	} else {
		year := time.Now().Year()
		races, err := nascar.FetchCupSchedule(year)
		if err != nil {
			fmt.Fprintf(os.Stderr, "raceday: %v\n", err)
			os.Exit(1)
		}

		race := nascar.NextRace(races)
		if race == nil {
			fmt.Print("No upcoming races")
			return
		}

		primary := 0
		if len(drivers) > 0 {
			primary = drivers[0]
		}
		segments = scheduleSegments(race, primary)
		if cfg.Weather {
			if ws := weatherSuffix(race.TrackID); ws != "" {
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

func weatherSuffix(trackID int) string {
	if trackID == 0 {
		return ""
	}
	lat, lon, ok := nascar.TrackCoords(trackID)
	if !ok {
		return ""
	}
	c, err := weather.FetchCurrent(lat, lon)
	if err != nil {
		return ""
	}
	return fmt.Sprintf(" | %.0f°F %s %.0fmph %s", c.Temp, weather.Symbol(c.WeatherCode), c.WindSpeed, weather.WindDirectionArrow(c.WindDirection))
}

func liveSegments(feed *nascar.LiveFeed, drivers []int) []segment {
	flagSym := nascar.FlagSymbol(feed.FlagState)
	segs := []segment{
		{fmt.Sprintf("%s %s", flagSym, feed.RunName), 0, true},
		{fmt.Sprintf(" | Lap %d/%d", feed.LapNumber, feed.LapsInRace), 0, false},
	}

	if leader := feed.Leader(); leader != nil {
		segs = append(segs, segment{
			fmt.Sprintf(" | P1 #%s %s", leader.VehicleNumber, leader.Driver.LastName), 2, false,
		})
	}

	for i, d := range drivers {
		carNum := strconv.Itoa(d)
		if v := feed.FindDriver(carNum); v != nil {
			diff := v.RunningPosition - v.StartingPosition
			diffStr := ""
			if diff > 0 {
				diffStr = fmt.Sprintf(" [+%d]", diff)
			} else if diff < 0 {
				diffStr = fmt.Sprintf(" [%d]", diff)
			}
			segs = append(segs, segment{
				fmt.Sprintf(" | #%s %s P%d%s", v.VehicleNumber, v.Driver.LastName, v.RunningPosition, diffStr), 1, i == 0,
			})
		}
	}

	return segs
}

func scheduleSegments(race *nascar.Race, driverNum int) []segment {
	start, err := race.RaceStartUTC()
	if err != nil {
		return []segment{{race.RaceName, 0, true}}
	}

	local := start.Local()
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
		{fmt.Sprintf("🏁 %s", race.RaceName), 0, true},
		{fmt.Sprintf(" | %s", timeStr), 0, false},
	}
	if race.TelevisionBroadcaster != "" {
		segs = append(segs, segment{fmt.Sprintf(" | %s", race.TelevisionBroadcaster), 2, false})
	}
	if driverNum > 0 {
		segs = append(segs, segment{fmt.Sprintf(" | #%d", driverNum), 1, true})
	}
	return segs
}
