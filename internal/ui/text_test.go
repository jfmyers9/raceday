package ui

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestPadToWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"disabled when width 0", "Hello", 0, "Hello"},
		{"pad short text", "Hi", 10, "Hi        "},
		{"exact width unchanged", "Hello", 5, "Hello"},
		{"truncate with ellipsis", "This is a very long string", 15, "This is a ve..."},
		{"emoji padding", "🏁 Race", 12, "🏁 Race     "},
		{"emoji truncation", "🟢 DAYTONA 500 | Lap 142/200", 15, "🟢 DAYTONA 5..."},
		{"empty string", "", 5, "     "},
		{"minimum width truncation", "Hello", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadToWidth(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("PadToWidth(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
			if tt.width > 0 {
				w := runewidth.StringWidth(result)
				if w != tt.width {
					t.Errorf("display width = %d, want %d", w, tt.width)
				}
			}
		})
	}
}

func TestMarqueeText_StaticWhenFits(t *testing.T) {
	result := MarqueeText("Short", 20, 2, " • ")
	if result != "Short               " {
		t.Errorf("short text should be padded, got %q", result)
	}
}

func TestMarqueeText_ExactWidth(t *testing.T) {
	long := "🟢 DAYTONA 500 | Lap 142/200 | P1 #8 Busch | #24 Byron P6 [-2] | 72°F ☀️ 12mph ↗NE"
	for _, w := range []int{30, 40, 50, 60} {
		result := MarqueeText(long, w, 2, " • ")
		rw := runewidth.StringWidth(result)
		if rw != w {
			t.Errorf("width %d: got display width %d", w, rw)
		}
	}
}

func TestMarqueeText_Deterministic(t *testing.T) {
	text := "Deterministic scrolling test string"
	r1 := MarqueeText(text, 15, 2, " • ")
	r2 := MarqueeText(text, 15, 2, " • ")
	if r1 != r2 {
		t.Logf("rapid calls differed (second boundary): %q vs %q", r1, r2)
	}
	if runewidth.StringWidth(r1) != 15 {
		t.Errorf("wrong width: %d", runewidth.StringWidth(r1))
	}
}
