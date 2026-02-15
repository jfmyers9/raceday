package ui

import (
	"fmt"
	"strings"

	"github.com/jfmyers/tmux-raceday/internal/nascar"
)

func renderStandingsView(standings []nascar.PointsEntry, favDriver string, width int) string {
	if len(standings) == 0 {
		return "Loading standings..."
	}

	var b strings.Builder

	title := titleStyle.Render("🏆 Cup Series Standings")
	b.WriteString(title)
	b.WriteString("\n\n")

	hdr := fmt.Sprintf("%-4s  %-4s  %-22s  %-6s  %-6s  %-5s  %-5s  %-5s  %-5s  %s",
		"POS", "#", "DRIVER", "PTS", "RACE", "WINS", "T5", "T10", "STAGE", "GAP")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteString("\n")

	for _, e := range standings {
		if !e.IsPointsEligible {
			continue
		}

		name := e.FirstName + " " + e.LastName
		stageTotal := e.Stage1Points + e.Stage2Points
		gapStr := ""
		if e.DeltaLeader > 0 {
			gapStr = fmt.Sprintf("-%d", e.DeltaLeader)
		}

		rookieTag := ""
		if e.IsRookie {
			rookieTag = " (R)"
		}

		row := fmt.Sprintf("%-4d  %-4s  %-22s  %-6d  %-6d  %-5d  %-5d  %-5d  %-5d  %s",
			e.PointsPosition,
			e.CarNumber,
			truncate(name+rookieTag, 22),
			e.Points,
			e.PointsThisRace,
			e.Wins,
			e.Top5,
			e.Top10,
			stageTotal,
			gapStr,
		)

		style := rowStyle
		if e.CarNumber == favDriver {
			style = favStyle
		}

		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}
