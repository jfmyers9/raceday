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
