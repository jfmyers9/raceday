package ui

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

// PadToWidth pads or truncates text to a fixed display width.
// Uses display columns (not bytes/runes), so emoji and CJK are handled correctly.
// Text longer than width is truncated with "..." suffix.
// Text shorter than width is padded with spaces.
func PadToWidth(text string, width int) string {
	if width <= 0 {
		return text
	}

	currentWidth := runewidth.StringWidth(text)

	if currentWidth > width {
		ellipsis := "..."
		ellipsisWidth := runewidth.StringWidth(ellipsis)

		if width <= ellipsisWidth {
			return runewidth.Truncate(ellipsis, width, "")
		}

		truncated := runewidth.Truncate(text, width-ellipsisWidth, "")
		result := truncated + ellipsis

		resultWidth := runewidth.StringWidth(result)
		if resultWidth < width {
			return result + strings.Repeat(" ", width-resultWidth)
		} else if resultWidth > width {
			return runewidth.Truncate(result, width, "")
		}
		return result
	}

	if currentWidth < width {
		return text + strings.Repeat(" ", width-currentWidth)
	}

	return text
}

// MarqueeText creates a scrolling marquee effect for text exceeding the target width.
// Position is derived from wall-clock time (stateless), so each tmux refresh
// advances the scroll by speed * status-interval characters.
func MarqueeText(text string, width int, speed int, separator string) string {
	if width <= 0 {
		return text
	}

	textWidth := runewidth.StringWidth(text)

	if textWidth <= width {
		return PadToWidth(text, width)
	}

	extended := text + separator + text
	extendedRunes := []rune(extended)

	now := time.Now().Unix()
	totalChars := len(extendedRunes)
	position := int(now*int64(speed)) % totalChars

	var result []rune
	resultWidth := 0

	for i := 0; i < totalChars && resultWidth < width; i++ {
		idx := (position + i) % totalChars
		r := extendedRunes[idx]
		rw := runewidth.RuneWidth(r)

		if resultWidth+rw <= width {
			result = append(result, r)
			resultWidth += rw
		} else {
			break
		}
	}

	out := string(result)
	// Final width check — per-rune tracking can drift from StringWidth
	// with multi-codepoint emoji (e.g. ☀️ = ☀ + VS16)
	actual := runewidth.StringWidth(out)
	if actual < width {
		out += strings.Repeat(" ", width-actual)
	}
	return out
}
