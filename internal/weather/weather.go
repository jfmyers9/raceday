package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 10 * time.Minute

var httpClient = &http.Client{Timeout: 10 * time.Second}

type Conditions struct {
	Temp          float64
	WeatherCode   int
	WindSpeed     float64
	WindGusts     float64
	Precipitation float64
	WindDirection  int
	FeelsLike     float64
}

type apiResponse struct {
	Current struct {
		Temperature   float64 `json:"temperature_2m"`
		WeatherCode   int     `json:"weather_code"`
		WindSpeed     float64 `json:"wind_speed_10m"`
		WindGusts     float64 `json:"wind_gusts_10m"`
		Precipitation float64 `json:"precipitation"`
		WindDirection  int     `json:"wind_direction_10m"`
		FeelsLike     float64 `json:"apparent_temperature"`
	} `json:"current"`
}

// FetchCurrent retrieves current weather conditions from Open-Meteo.
func FetchCurrent(lat, lon float64) (*Conditions, error) {
	key := fmt.Sprintf("weather_%.4f_%.4f.json", lat, lon)

	if data, ok := readCache(key); ok {
		var c Conditions
		if err := json.Unmarshal(data, &c); err == nil {
			return &c, nil
		}
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f"+
			"&current=temperature_2m,weather_code,wind_speed_10m,wind_gusts_10m,precipitation,wind_direction_10m,apparent_temperature"+
			"&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch",
		lat, lon,
	)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("weather fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather fetch: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("weather read body: %w", err)
	}

	var api apiResponse
	if err := json.Unmarshal(body, &api); err != nil {
		return nil, fmt.Errorf("weather parse: %w", err)
	}

	c := &Conditions{
		Temp:          api.Current.Temperature,
		WeatherCode:   api.Current.WeatherCode,
		WindSpeed:     api.Current.WindSpeed,
		WindGusts:     api.Current.WindGusts,
		Precipitation: api.Current.Precipitation,
		WindDirection:  api.Current.WindDirection,
		FeelsLike:     api.Current.FeelsLike,
	}

	if data, err := json.Marshal(c); err == nil {
		_ = writeCache(key, data)
	}

	return c, nil
}

// Symbol maps a WMO weather code to a compact emoji.
func Symbol(code int) string {
	switch code {
	case 0:
		return "☀️"
	case 1, 2, 3:
		return "⛅"
	case 45, 48:
		return "🌫"
	case 51, 53, 55, 56, 57:
		return "🌧"
	case 61, 63, 65, 66, 67:
		return "🌧"
	case 71, 73, 75, 77:
		return "🌨"
	case 80, 81, 82:
		return "🌧"
	case 85, 86:
		return "🌨"
	case 95, 96, 99:
		return "⛈"
	default:
		return "🌤"
	}
}

// WindDirectionArrow returns a compass arrow for wind direction in degrees.
func WindDirectionArrow(deg int) string {
	arrows := []string{"↑N", "↗NE", "→E", "↘SE", "↓S", "↙SW", "←W", "↖NW"}
	idx := ((deg + 22) % 360) / 45
	return arrows[idx]
}

func cacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "raceday")
}

func cachePath(key string) string { return filepath.Join(cacheDir(), key) }

func readCache(key string) ([]byte, bool) {
	p := cachePath(key)
	info, err := os.Stat(p)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > cacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return data, true
}

func writeCache(key string, data []byte) error {
	if err := os.MkdirAll(cacheDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cachePath(key), data, 0o644)
}
