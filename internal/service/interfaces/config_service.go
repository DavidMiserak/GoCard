// internal/service/interfaces/config_service.go
package interfaces

// ConfigService manages application configuration
type ConfigService interface {
	// Configuration management
	GetConfig() (Config, error)
	SetConfig(config Config) error

	// Individual settings
	GetString(key string, defaultValue string) string
	GetInt(key string, defaultValue int) int
	GetBool(key string, defaultValue bool) bool
	GetFloat(key string, defaultValue float64) float64

	// Storage
	SaveConfig() error
	ResetToDefaults() error
}

// Config holds application configuration
type Config struct {
	CardsDir         string  // Root directory for cards
	Theme            string  // UI theme
	CodeTheme        string  // Code syntax highlighting theme
	EasyBonus        float64 // SM-2 algorithm easy bonus multiplier
	IntervalModifier float64 // SM-2 algorithm interval modifier
	NewCardsPerDay   int     // Limit on new cards per day
	MaxInterval      int     // Maximum review interval in days
	ShowLineNumbers  bool    // Whether to show line numbers in code blocks
}
