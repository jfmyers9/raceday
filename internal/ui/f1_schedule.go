package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers/tmux-raceday/internal/series"
)

func renderF1ScheduleView(races []series.Race, width int) string {
	if len(races) == 0 {
		return "No F1 schedule available."
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("🏎 F1 Calendar"))
	b.WriteString("\n\n")

	hdr := fmt.Sprintf("%-14s  %-35s  %-18s  %s",
		"DATE", "RACE", "CIRCUIT", "COUNTDOWN")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteString("\n")

	now := time.Now()
	nextHighlighted := false

	for _, r := range races {
		dateStr := r.StartTime.Local().Format("Mon Jan 2")
		countdown := formatCountdown(r.StartTime, now.UTC())

		style := rowStyle
		if r.Complete || r.StartTime.Before(now) {
			style = dimStyle
			countdown = "done"
		} else if !nextHighlighted {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)
			nextHighlighted = true
		}

		row := fmt.Sprintf("%-14s  %-35s  %-18s  %s",
			dateStr,
			truncate(r.RaceName, 35),
			truncate(r.TrackName, 18),
			countdown,
		)
		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}
