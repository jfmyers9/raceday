package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers/tmux-raceday/internal/nascar"
)

func renderScheduleView(race *nascar.Race, width int) string {
	if race == nil {
		return "No race schedule available."
	}

	var b strings.Builder

	title := titleStyle.Render(fmt.Sprintf("📅 %s — %s", race.RaceName, race.TrackName))
	b.WriteString(title)
	b.WriteString("\n\n")

	now := time.Now().UTC()

	// Column header
	hdr := fmt.Sprintf("%-6s  %-40s  %-20s  %s", "TIME", "EVENT", "NOTES", "COUNTDOWN")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteString("\n")

	for _, ev := range race.Schedule {
		evTime, err := time.Parse("2006-01-02T15:04:05", ev.StartTimeUTC)
		if err != nil {
			continue
		}

		local := evTime.Local()
		timeStr := local.Format("Mon 3:04 PM")

		countdown := formatCountdown(evTime, now)

		style := rowStyle
		runTypeLabel := ""
		switch ev.RunType {
		case 1:
			runTypeLabel = " 🟢"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		case 2:
			runTypeLabel = " 🏎"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		case 3:
			runTypeLabel = " 🏁"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)
		}

		// Dim past events
		if evTime.Before(now) {
			style = dimStyle
			countdown = "done"
		}

		name := truncate(ev.EventName, 38) + runTypeLabel
		notes := truncate(ev.Notes, 20)

		row := fmt.Sprintf("%-6s  %-40s  %-20s  %s", timeStr, name, notes, countdown)
		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}

func formatCountdown(target, now time.Time) string {
	diff := target.Sub(now)
	if diff < 0 {
		return ""
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	mins := int(diff.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}
