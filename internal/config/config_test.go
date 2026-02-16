package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDurationUnmarshalYAML(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{`"2h"`, 2 * time.Hour},
		{`"30m"`, 30 * time.Minute},
		{`"1h30m"`, 90 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var d Duration
			if err := yaml.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("unmarshal %s: %v", tt.input, err)
			}
			if time.Duration(d) != tt.want {
				t.Errorf("got %v, want %v", time.Duration(d), tt.want)
			}
		})
	}
}

func TestDurationUnmarshalYAMLInvalid(t *testing.T) {
	var d Duration
	if err := yaml.Unmarshal([]byte(`"not-a-duration"`), &d); err == nil {
		t.Error("expected error for invalid duration string")
	}
}

func TestWeatherWindowDefault(t *testing.T) {
	cfg := DefaultConfig()
	if time.Duration(cfg.WeatherWindow) != 2*time.Hour {
		t.Errorf("default WeatherWindow = %v, want 2h", time.Duration(cfg.WeatherWindow))
	}
}

func TestWeatherWindowFromYAML(t *testing.T) {
	input := `weather_window: "45m"`
	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if time.Duration(cfg.WeatherWindow) != 45*time.Minute {
		t.Errorf("WeatherWindow = %v, want 45m", time.Duration(cfg.WeatherWindow))
	}
}
