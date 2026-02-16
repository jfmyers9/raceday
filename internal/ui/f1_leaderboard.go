package ui

import (
	"fmt"
	"strings"

	"github.com/jfmyers/tmux-raceday/internal/series"
)

func compoundAbbrev(compound string) string {
	switch compound {
	case "SOFT":
		return "S"
	case "MEDIUM":
		return "M"
	case "HARD":
		return "H"
	case "INTERMEDIATE":
		return "I"
	case "WET":
		return "W"
	default:
		return ""
	}
}

func renderF1LeaderboardView(state *series.LiveState, width int) string {
	if state == nil {
		return "No live F1 session."
	}

	var b strings.Builder

	flagPart := ""
	if state.FlagSymbol != "" {
		flagPart = state.FlagSymbol + " "
	}
	title := fmt.Sprintf("%s%s — %s", flagPart, state.RaceName, state.TrackName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	var info []string
	if state.TotalLaps > 0 {
		info = append(info, fmt.Sprintf("Lap %d/%d", state.CurrentLap, state.TotalLaps))
	} else if state.CurrentLap > 0 {
		info = append(info, fmt.Sprintf("Lap %d", state.CurrentLap))
	}
	if state.FlagName != "" {
		info = append(info, state.FlagName)
	}
	if len(info) > 0 {
		b.WriteString(dimStyle.Render(strings.Join(info, " | ")))
	}
	b.WriteString("\n\n")

	hdr := fmt.Sprintf("%-4s %-4s %-5s %-4s %-22s %-20s %s",
		"POS", "#", "DRV", "TYRE", "NAME", "TEAM", "GAP")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteString("\n")

	for _, d := range state.Positions {
		gap := d.Gap
		if d.Position == 1 {
			gap = "LEADER"
		}
		tyre := compoundAbbrev(d.Compound)
		row := fmt.Sprintf("%-4d %-4s %-5s %-4s %-22s %-20s %s",
			d.Position,
			d.Number,
			d.Name,
			tyre,
			truncate(d.FullName, 22),
			truncate(d.Team, 20),
			gap,
		)
		b.WriteString(rowStyle.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}
