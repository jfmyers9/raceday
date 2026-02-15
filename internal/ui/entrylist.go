package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jfmyers/tmux-raceday/internal/nascar"
)

func renderEntryListView(feed *nascar.LiveFeed, favDriver string, width int) string {
	if feed == nil || len(feed.Vehicles) == 0 {
		return "No entry list available (race not live)."
	}

	var b strings.Builder

	title := titleStyle.Render(fmt.Sprintf("📋 Entry List — %s", feed.RunName))
	b.WriteString(title)
	b.WriteString("\n\n")

	hdr := fmt.Sprintf("%-4s  %-22s  %-4s  %-20s  %-5s",
		"#", "DRIVER", "MFR", "SPONSOR", "START")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteString("\n")

	// Sort by car number
	vehicles := make([]nascar.Vehicle, len(feed.Vehicles))
	copy(vehicles, feed.Vehicles)
	sort.SliceStable(vehicles, func(i, j int) bool {
		return vehicles[i].VehicleNumber < vehicles[j].VehicleNumber
	})

	for _, v := range vehicles {
		mfr := expandManufacturer(v.VehicleManufacturer)
		row := fmt.Sprintf("%-4s  %-22s  %-4s  %-20s  P%-4d",
			v.VehicleNumber,
			truncate(v.Driver.FullName, 22),
			mfr,
			truncate(v.SponsorName, 20),
			v.StartingPosition,
		)

		style := rowStyle
		if v.VehicleNumber == favDriver {
			style = favStyle
		}
		if v.Status != 1 || v.IsOnDVP {
			style = dimStyle
		}

		b.WriteString(style.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}

func expandManufacturer(code string) string {
	switch code {
	case "Tyt":
		return "TYT"
	case "Chv":
		return "CHV"
	case "Frd":
		return "FRD"
	default:
		return code
	}
}
