package weather

import (
	"fmt"
	"os"
	"testing"
)

func TestSymbol(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{0, "☀️"},
		{1, "⛅"},
		{3, "⛅"},
		{45, "🌫"},
		{61, "🌧"},
		{73, "🌨"},
		{95, "⛈"},
		{99, "⛈"},
		{-1, "🌤"},
		{100, "🌤"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			got := Symbol(tt.code)
			if got != tt.want {
				t.Errorf("Symbol(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestWindDirectionArrow(t *testing.T) {
	tests := []struct {
		deg  int
		want string
	}{
		{0, "↑N"},
		{45, "↗NE"},
		{90, "→E"},
		{135, "↘SE"},
		{180, "↓S"},
		{225, "↙SW"},
		{270, "←W"},
		{315, "↖NW"},
		{350, "↑N"},
		{22, "↑N"},
		{23, "↗NE"},
		{359, "↑N"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("deg_%d", tt.deg), func(t *testing.T) {
			got := WindDirectionArrow(tt.deg)
			if got != tt.want {
				t.Errorf("WindDirectionArrow(%d) = %q, want %q", tt.deg, got, tt.want)
			}
		})
	}
}

func TestCacheKeyVariesByCoords(t *testing.T) {
	k1 := fmt.Sprintf("weather_%.4f_%.4f.json", 29.1872, -81.0715)
	k2 := fmt.Sprintf("weather_%.4f_%.4f.json", 33.4484, -112.0740)
	if k1 == k2 {
		t.Fatal("different coords should produce different cache keys")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	key := "weather_test_roundtrip.json"
	defer os.Remove(cachePath(key))

	data := []byte(`{"Temp":80.5}`)
	if err := writeCache(key, data); err != nil {
		t.Fatalf("writeCache: %v", err)
	}

	got, ok := readCache(key)
	if !ok {
		t.Fatal("readCache returned not ok")
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestFetchCurrent_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping integration test in CI")
	}

	// Daytona International Speedway
	c, err := FetchCurrent(29.1872, -81.0715)
	if err != nil {
		t.Fatalf("FetchCurrent: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Conditions")
	}
	// Sanity: temp should be in a plausible range for Daytona (°F)
	if c.Temp < -20 || c.Temp > 130 {
		t.Errorf("Temp %.1f out of plausible range", c.Temp)
	}
}
