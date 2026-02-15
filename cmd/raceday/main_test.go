package main

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestAssembleSegments(t *testing.T) {
	segs := []segment{
		{"🟢 DAYTONA 500 | Lap 142/200", 0, true},
		{" | #24 Byron P6 [-2]", 1, true},
		{" | P1 #8 Busch", 2, false},
		{" | 72°F ☀️ 12mph", 3, false},
	}

	all := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2] | P1 #8 Busch | 72°F ☀️ 12mph"

	t.Run("no width shows all", func(t *testing.T) {
		result := assembleSegments(segs, 0)
		if result != all {
			t.Errorf("got %q", result)
		}
	})

	t.Run("marquee mode passes width 0", func(t *testing.T) {
		result := assembleSegments(segs, 0)
		if result != all {
			t.Errorf("marquee mode should include all segments\ngot:  %q\nwant: %q", result, all)
		}
		narrow := assembleSegments(segs, 40)
		if narrow == all {
			t.Error("narrow width should have dropped segments but didn't")
		}
	})

	t.Run("drops weather first", func(t *testing.T) {
		result := assembleSegments(segs, 65)
		expected := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2] | P1 #8 Busch"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("drops leader before driver", func(t *testing.T) {
		result := assembleSegments(segs, 50)
		expected := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2]"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("core always included", func(t *testing.T) {
		result := assembleSegments(segs, 30)
		expected := "🟢 DAYTONA 500 | Lap 142/200"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})
}

func TestAssembleHybrid(t *testing.T) {
	segs := []segment{
		{"🟢 DAYTONA 500 | Lap 142/200", 0, true},
		{" | #24 Byron P6", 1, true},
		{" | P1 #8 Busch", 2, false},
		{" | 72°F Sun 12mph", 3, false},
	}

	t.Run("static segments with spacer before marquee", func(t *testing.T) {
		result := assembleHybrid(segs, 80, 2, " • ")
		staticWithSpacer := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 | "
		if len(result) < len(staticWithSpacer) || result[:len(staticWithSpacer)] != staticWithSpacer {
			t.Errorf("static prefix + spacer missing\ngot:  %q\nwant prefix: %q", result, staticWithSpacer)
		}
	})

	t.Run("output width matches target", func(t *testing.T) {
		for _, w := range []int{40, 60, 80, 100} {
			result := assembleHybrid(segs, w, 2, " • ")
			got := runewidth.StringWidth(result)
			if got != w {
				t.Errorf("width %d: display width = %d", w, got)
			}
		}
	})

	t.Run("narrow width drops low-priority statics", func(t *testing.T) {
		// Width so small only priority-0 static fits
		result := assembleHybrid(segs, 20, 2, " • ")
		got := runewidth.StringWidth(result)
		if got != 20 {
			t.Errorf("display width = %d, want 20", got)
		}
	})

	t.Run("marquee pads when content fits", func(t *testing.T) {
		short := []segment{
			{"Flag", 0, true},
			{" | extra", 0, false},
		}
		result := assembleHybrid(short, 40, 2, " • ")
		if runewidth.StringWidth(result) != 40 {
			t.Errorf("display width = %d, want 40", runewidth.StringWidth(result))
		}
		// Static "Flag" + spacer " | " + dynamic " | extra" padded
		expected := "Flag |  | extra"
		if len(result) < len(expected) || result[:len(expected)] != expected {
			t.Errorf("short content should be padded, got %q", result)
		}
	})

	t.Run("no dynamic segments pads static", func(t *testing.T) {
		staticOnly := []segment{
			{"🟢 DAYTONA 500", 0, true},
			{" | #24 Byron P6", 1, true},
		}
		result := assembleHybrid(staticOnly, 60, 2, " • ")
		got := runewidth.StringWidth(result)
		if got != 60 {
			t.Errorf("display width = %d, want 60", got)
		}
	})
}
