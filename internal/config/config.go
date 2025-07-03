package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	MaxHistoryItems   int  `json:"max_history_items"`
	MaxHistoryDays    int  `json:"max_history_days"`
	StartWithSystem   bool `json:"start_with_system"`
	ShowNotifications bool `json:"show_notifications"`
	DarkMode          bool `json:"dark_mode"`

	MonitorInterval int `json:"monitor_interval_ms"`
	MaxItemSize     int `json:"max_item_size_bytes"`

	// Update settings
	CheckUpdatesOnStartup bool `json:"check_updates_on_startup"`
	AutoDownloadUpdates   bool `json:"auto_download_updates"`
}

func Default() *Config {
	return &Config{
		MaxHistoryItems:   1000,
		MaxHistoryDays:    30,
		StartWithSystem:   true,
		ShowNotifications: true,
		DarkMode:          false,

		MonitorInterval: 500,
		MaxItemSize:     10 * 1024 * 1024, // 10MB

		CheckUpdatesOnStartup: true,
		AutoDownloadUpdates:   false,
	}
}

func Load(path string) (*Config, error) {
	config := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return default config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	config.validate()

	return config, nil
}

func (c *Config) Save(path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) validate() {
	if c.MaxHistoryItems <= 0 {
		c.MaxHistoryItems = 1000
	}
	if c.MaxHistoryDays <= 0 {
		c.MaxHistoryDays = 30
	}
	if c.MonitorInterval <= 0 {
		c.MonitorInterval = 500
	}
	if c.MaxItemSize <= 0 {
		c.MaxItemSize = 10 * 1024 * 1024
	}
}
