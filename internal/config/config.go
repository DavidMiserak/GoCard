// internal/config/config.go - Configuration management
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	// Basic settings
	CardsDir string `yaml:"cards_dir"`
	FirstRun bool   `yaml:"first_run"`

	// Logging settings
	Logging struct {
		Level       string `yaml:"level"`
		FileEnabled bool   `yaml:"file_enabled"`
		FilePath    string `yaml:"file_path"`
	} `yaml:"logging"`

	// UI settings
	UI struct {
		Theme           string `yaml:"theme"`
		HighlightTheme  string `yaml:"highlight_theme"`
		ShowLineNumbers bool   `yaml:"show_line_numbers"`
	} `yaml:"ui"`

	// Spaced repetition settings
	SpacedRepetition struct {
		EasyBonus        float64 `yaml:"easy_bonus"`
		IntervalModifier float64 `yaml:"interval_modifier"`
		MaxInterval      int     `yaml:"max_interval"`
		NewCardsPerDay   int     `yaml:"new_cards_per_day"`
	} `yaml:"spaced_repetition"`
}

// Default returns the default configuration
func Default() *Config {
	cfg := &Config{
		CardsDir: "~/GoCard",
		FirstRun: true,
	}

	// Default logging settings
	cfg.Logging.Level = "info"
	cfg.Logging.FileEnabled = false
	cfg.Logging.FilePath = "~/.gocard.log"

	// Default UI settings
	cfg.UI.Theme = "auto"
	cfg.UI.HighlightTheme = "monokai"
	cfg.UI.ShowLineNumbers = true

	// Default spaced repetition settings
	cfg.SpacedRepetition.EasyBonus = 1.3
	cfg.SpacedRepetition.IntervalModifier = 1.0
	cfg.SpacedRepetition.MaxInterval = 365
	cfg.SpacedRepetition.NewCardsPerDay = 20

	return cfg
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

// getConfigPath returns the path to the config file
func getConfigPath(customPath string) string {
	if customPath != "" {
		path, _ := expandPath(customPath)
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".gocard.yaml"
	}
	return filepath.Join(homeDir, ".gocard.yaml")
}

// Load loads the configuration from a file
func Load(customPath string) (*Config, error) {
	configPath := getConfigPath(customPath)

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, return default config
		return Default(), nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths
	cfg.CardsDir, err = expandPath(cfg.CardsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	cfg.Logging.FilePath, _ = expandPath(cfg.Logging.FilePath)

	return cfg, nil
}

// Save saves the configuration to a file
func Save(cfg *Config) error {
	configPath := getConfigPath("")

	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
