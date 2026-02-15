package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Drivers          []int  `yaml:"drivers"`
	Series           int    `yaml:"series"`
	Theme            string `yaml:"theme"`
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
		Series:           1, // Cup Series
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
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "raceday", "config.yaml")
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
