// internal/service/config/yaml_config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/service/interfaces"

	"gopkg.in/yaml.v3"
)

// Default configuration values
const (
	DefaultConfigFile       = "~/.gocard.yaml"
	DefaultCardsDir         = "~/GoCard"
	DefaultTheme            = "default"
	DefaultCodeTheme        = "monokai"
	DefaultEasyBonus        = 1.3
	DefaultIntervalModifier = 1.0
	DefaultNewCardsPerDay   = 20
	DefaultMaxInterval      = 365
	DefaultShowLineNumbers  = true
)

// YAMLConfig implements the ConfigService interface using YAML files
type YAMLConfig struct {
	configPath string
	config     interfaces.Config
}

// NewYAMLConfig creates a new YAML-based configuration service
func NewYAMLConfig(configPath string) (interfaces.ConfigService, error) {
	// Expand home directory in config path if needed
	if configPath == DefaultConfigFile || configPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Create a new config service
	configSvc := &YAMLConfig{
		configPath: configPath,
		config:     getDefaultConfig(),
	}

	// Try to load the config file
	if _, err := os.Stat(configPath); err == nil {
		if err := configSvc.loadConfig(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Config file doesn't exist, save the default config
		if err := configSvc.SaveConfig(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return configSvc, nil
}

// GetConfig returns the current configuration
func (yc *YAMLConfig) GetConfig() (interfaces.Config, error) {
	return yc.config, nil
}

// SetConfig updates the configuration
func (yc *YAMLConfig) SetConfig(config interfaces.Config) error {
	yc.config = config
	return yc.SaveConfig()
}

// GetString retrieves a string configuration value
func (yc *YAMLConfig) GetString(key string, defaultValue string) string {
	switch key {
	case "CardsDir":
		return yc.config.CardsDir
	case "Theme":
		return yc.config.Theme
	case "CodeTheme":
		return yc.config.CodeTheme
	default:
		return defaultValue
	}
}

// GetInt retrieves an integer configuration value
func (yc *YAMLConfig) GetInt(key string, defaultValue int) int {
	switch key {
	case "NewCardsPerDay":
		return yc.config.NewCardsPerDay
	case "MaxInterval":
		return yc.config.MaxInterval
	default:
		return defaultValue
	}
}

// GetBool retrieves a boolean configuration value
func (yc *YAMLConfig) GetBool(key string, defaultValue bool) bool {
	switch key {
	case "ShowLineNumbers":
		return yc.config.ShowLineNumbers
	default:
		return defaultValue
	}
}

// GetFloat retrieves a float configuration value
func (yc *YAMLConfig) GetFloat(key string, defaultValue float64) float64 {
	switch key {
	case "EasyBonus":
		return yc.config.EasyBonus
	case "IntervalModifier":
		return yc.config.IntervalModifier
	default:
		return defaultValue
	}
}

// SaveConfig saves the configuration to the YAML file
func (yc *YAMLConfig) SaveConfig() error {
	// Create parent directory if needed
	configDir := filepath.Dir(yc.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(yc.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(yc.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResetToDefaults resets the configuration to default values
func (yc *YAMLConfig) ResetToDefaults() error {
	yc.config = getDefaultConfig()
	return yc.SaveConfig()
}

// loadConfig loads the configuration from the YAML file
func (yc *YAMLConfig) loadConfig() error {
	// Read the config file
	data, err := os.ReadFile(yc.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal YAML
	var config interfaces.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Update the configuration
	yc.config = config

	return nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() interfaces.Config {
	// Expand home directory in cards directory
	cardsDir := DefaultCardsDir
	if home, err := os.UserHomeDir(); err == nil {
		cardsDir = filepath.Join(home, cardsDir[2:])
	}

	return interfaces.Config{
		CardsDir:         cardsDir,
		Theme:            DefaultTheme,
		CodeTheme:        DefaultCodeTheme,
		EasyBonus:        DefaultEasyBonus,
		IntervalModifier: DefaultIntervalModifier,
		NewCardsPerDay:   DefaultNewCardsPerDay,
		MaxInterval:      DefaultMaxInterval,
		ShowLineNumbers:  DefaultShowLineNumbers,
	}
}

// Ensure YAMLConfig implements ConfigService
var _ interfaces.ConfigService = (*YAMLConfig)(nil)
