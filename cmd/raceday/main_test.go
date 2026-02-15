package main

import (
	"testing"
)

func TestAssembleSegments(t *testing.T) {
	segs := []segment{
		{"🟢 DAYTONA 500 | Lap 142/200", 0},
		{" | #24 Byron P6 [-2]", 1},
		{" | P1 #8 Busch", 2},
		{" | 72°F ☀️ 12mph", 3},
	}

	// Marquee mode: width=0 preserves all segments so MarqueeText
	// receives the full string and actually scrolls.
	t.Run("marquee mode passes width 0", func(t *testing.T) {
		result := assembleSegments(segs, 0)
		all := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2] | P1 #8 Busch | 72°F ☀️ 12mph"
		if result != all {
			t.Errorf("marquee mode should include all segments\ngot:  %q\nwant: %q", result, all)
		}
		// With a narrow width, segments would be dropped — confirm that.
		narrow := assembleSegments(segs, 40)
		if narrow == all {
			t.Error("narrow width should have dropped segments but didn't")
		}
	})

	t.Run("no width shows all", func(t *testing.T) {
		result := assembleSegments(segs, 0)
		if result != "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2] | P1 #8 Busch | 72°F ☀️ 12mph" {
			t.Errorf("got %q", result)
		}
	})

	t.Run("drops weather first", func(t *testing.T) {
		result := assembleSegments(segs, 65)
		// Weather (pri 3) dropped, rest fits
		expected := "🟢 DAYTONA 500 | Lap 142/200 | #24 Byron P6 [-2] | P1 #8 Busch"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("drops leader before driver", func(t *testing.T) {
		result := assembleSegments(segs, 50)
		// Weather + leader (pri 3,2) dropped, driver stays
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
