package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SeriesList holds configured series names (e.g. "nascar", "f1").
// Backward-compatible: unmarshals from int (old format) or string or []string.
type SeriesList []string

func (s *SeriesList) UnmarshalYAML(value *yaml.Node) error {
	var list []string
	if err := value.Decode(&list); err == nil {
		*s = list
		return nil
	}
	var n int
	if err := value.Decode(&n); err == nil {
		*s = []string{"nascar"}
		return nil
	}
	var str string
	if err := value.Decode(&str); err == nil {
		*s = []string{str}
		return nil
	}
	return fmt.Errorf("invalid series format")
}

// DriverMap maps series name to driver numbers.
// Backward-compatible: unmarshals from []int (old format, treated as NASCAR).
type DriverMap map[string][]int

func (d *DriverMap) UnmarshalYAML(value *yaml.Node) error {
	var m map[string][]int
	if err := value.Decode(&m); err == nil {
		*d = m
		return nil
	}
	var list []int
	if err := value.Decode(&list); err == nil {
		*d = map[string][]int{"nascar": list}
		return nil
	}
	return fmt.Errorf("invalid drivers format")
}

type Config struct {
	Drivers          DriverMap  `yaml:"drivers"`
	Series           SeriesList `yaml:"series"`
	Theme            string     `yaml:"theme"`
	Weather          bool   `yaml:"weather"`
	Notify           Notify `yaml:"notify"`
	StatusWidth      int    `yaml:"status_width"`
	Marquee          bool   `yaml:"marquee"`
	MarqueeSpeed     int    `yaml:"marquee_speed"`
	MarqueeSeparator string `yaml:"marquee_separator"`
}

type Notify struct {
	Cautions    bool `yaml:"cautions"`
	LeadChanges bool `yaml:"lead_changes"`
	Desktop     bool `yaml:"desktop"`
}

func DefaultConfig() Config {
	return Config{
		Series:           SeriesList{"nascar"},
		Drivers:          DriverMap{},
		Theme:            "default",
		Weather:          true,
		MarqueeSpeed:     2,
		MarqueeSeparator: " • ",
		Notify: Notify{
			Cautions:    true,
			LeadChanges: false,
			Desktop:     false,
		},
	}
}

func configPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return filepath.Join(home, ".config", "raceday", "config.yaml")
}

// Load reads config from ~/.config/raceday/config.yaml.
// Returns default config if file doesn't exist.
func Load() Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

// Save writes the config to disk, creating directories as needed.
func Save(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// EnsureDefault creates a default config file if none exists.
func EnsureDefault() {
	if _, err := os.Stat(configPath()); os.IsNotExist(err) {
		_ = Save(DefaultConfig())
	}
}
