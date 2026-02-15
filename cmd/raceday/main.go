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
)

func main() {
	status := flag.Bool("status", false, "Print one-line status for tmux status bar")
	driver := flag.Int("driver", 0, "Favorite driver car number (overrides config)")
	initCfg := flag.Bool("init-config", false, "Create default config file")
	flag.Parse()

	if *initCfg {
		config.EnsureDefault()
		fmt.Println("Config created at ~/.config/raceday/config.yaml")
		return
	}

	cfg := config.Load()

	// CLI --driver overrides config
	drivers := cfg.Drivers
	if *driver > 0 {
		drivers = []int{*driver}
	}

	if *status {
		runStatus(drivers)
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

func runStatus(drivers []int) {
	// Try live feed first
	live, err := nascar.FetchLiveFeed()
	if err == nil && live.IsLiveCupRace() {
		fmt.Print(formatLiveStatus(live, drivers))
		return
	}

	// Fall back to schedule
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
	fmt.Print(formatScheduleStatus(race, primary))
}

func formatLiveStatus(feed *nascar.LiveFeed, drivers []int) string {
	flagSym := nascar.FlagSymbol(feed.FlagState)
	s := fmt.Sprintf("%s %s | Lap %d/%d", flagSym, feed.RunName, feed.LapNumber, feed.LapsInRace)

	if leader := feed.Leader(); leader != nil {
		s += fmt.Sprintf(" | P1 #%s %s", leader.VehicleNumber, leader.Driver.LastName)
	}

	var driverParts []string
	for _, d := range drivers {
		carNum := strconv.Itoa(d)
		if v := feed.FindDriver(carNum); v != nil {
			diff := v.RunningPosition - v.StartingPosition
			diffStr := ""
			if diff > 0 {
				diffStr = fmt.Sprintf(" [+%d]", diff)
			} else if diff < 0 {
				diffStr = fmt.Sprintf(" [%d]", diff)
			}
			driverParts = append(driverParts, fmt.Sprintf("#%s %s P%d%s", v.VehicleNumber, v.Driver.LastName, v.RunningPosition, diffStr))
		}
	}
	if len(driverParts) > 0 {
		s += " | " + strings.Join(driverParts, " ")
	}

	return s
}

func formatScheduleStatus(race *nascar.Race, driverNum int) string {
	start, err := race.RaceStartUTC()
	if err != nil {
		return race.RaceName
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

	s := fmt.Sprintf("🏁 %s | %s", race.RaceName, timeStr)
	if race.TelevisionBroadcaster != "" {
		s += fmt.Sprintf(" | %s", race.TelevisionBroadcaster)
	}
	if driverNum > 0 {
		s += fmt.Sprintf(" | #%d", driverNum)
	}
	return s
}
