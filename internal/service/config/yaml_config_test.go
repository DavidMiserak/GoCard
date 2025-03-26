// internal/service/config/yaml_config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
)

// Helper to create a temporary config file
func createTempConfigFile(t *testing.T, content string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "gocard-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.yaml")
	if content != "" {
		err := os.WriteFile(configPath, []byte(content), 0644)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("failed to write temp config file: %v", err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return configPath, cleanup
}

// Create a test config struct
func createTestConfig() interfaces.Config {
	return interfaces.Config{
		CardsDir:         "/test/cards",
		Theme:            "test-theme",
		CodeTheme:        "test-code-theme",
		EasyBonus:        1.5,
		IntervalModifier: 1.2,
		NewCardsPerDay:   15,
		MaxInterval:      300,
		ShowLineNumbers:  false,
	}
}

func TestNewYAMLConfig(t *testing.T) {
	// Test with non-existent file (should create default config)
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()

	// Create new config service
	_, err := NewYAMLConfig(configPath)
	if err != nil {
		t.Fatalf("NewYAMLConfig() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", configPath)
	}

	// Test with existing file
	existingContent := `
cardsdir: /existing/path
theme: existing-theme
codetheme: existing-code-theme
easybonus: 1.4
intervalmodifier: 1.1
newcardsperday: 25
maxinterval: 500
showlinenumbers: true
`
	configPath2, cleanup2 := createTempConfigFile(t, existingContent)
	defer cleanup2()

	// Create new config service with existing file
	configSvc2, err := NewYAMLConfig(configPath2)
	if err != nil {
		t.Fatalf("NewYAMLConfig() with existing file error = %v", err)
	}

	// Get the config to verify it loaded from file
	config, _ := configSvc2.GetConfig()
	if config.Theme != "existing-theme" {
		t.Errorf("expected theme 'existing-theme', got '%s'", config.Theme)
	}
}

func TestGetConfig(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Get config
	config, err := configSvc.GetConfig()
	if err != nil {
		t.Errorf("GetConfig() error = %v", err)
	}

	// Verify default values
	if config.Theme != DefaultTheme {
		t.Errorf("expected default theme '%s', got '%s'", DefaultTheme, config.Theme)
	}
	if config.CardsDir == "" {
		t.Errorf("expected cards directory to be set")
	}
}

func TestSetConfig(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create test config
	testConfig := createTestConfig()

	// Set config
	err := configSvc.SetConfig(testConfig)
	if err != nil {
		t.Errorf("SetConfig() error = %v", err)
	}

	// Get config to verify it was set
	config, _ := configSvc.GetConfig()
	if config.Theme != testConfig.Theme {
		t.Errorf("expected theme '%s', got '%s'", testConfig.Theme, config.Theme)
	}
	if config.CardsDir != testConfig.CardsDir {
		t.Errorf("expected cards dir '%s', got '%s'", testConfig.CardsDir, config.CardsDir)
	}
	if config.EasyBonus != testConfig.EasyBonus {
		t.Errorf("expected easy bonus %f, got %f", testConfig.EasyBonus, config.EasyBonus)
	}
}

func TestGetString(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set test config
	testConfig := createTestConfig()
	configSvc.SetConfig(testConfig)

	// Test cases
	testCases := []struct {
		key      string
		expected string
		default_ string
	}{
		{"CardsDir", testConfig.CardsDir, "default-dir"},
		{"Theme", testConfig.Theme, "default-theme"},
		{"CodeTheme", testConfig.CodeTheme, "default-code-theme"},
		{"NonExistent", "non-existent-default", "non-existent-default"},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := configSvc.GetString(tc.key, tc.default_)
			if result != tc.expected {
				t.Errorf("GetString(%s) expected '%s', got '%s'", tc.key, tc.expected, result)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set test config
	testConfig := createTestConfig()
	configSvc.SetConfig(testConfig)

	// Test cases
	testCases := []struct {
		key      string
		expected int
		default_ int
	}{
		{"NewCardsPerDay", testConfig.NewCardsPerDay, 10},
		{"MaxInterval", testConfig.MaxInterval, 200},
		{"NonExistent", 42, 42},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := configSvc.GetInt(tc.key, tc.default_)
			if result != tc.expected {
				t.Errorf("GetInt(%s) expected %d, got %d", tc.key, tc.expected, result)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set test config
	testConfig := createTestConfig()
	configSvc.SetConfig(testConfig)

	// Test ShowLineNumbers (should be false in test config)
	result := configSvc.GetBool("ShowLineNumbers", true)
	if result != testConfig.ShowLineNumbers {
		t.Errorf("GetBool(ShowLineNumbers) expected %v, got %v", testConfig.ShowLineNumbers, result)
	}

	// Test non-existent key (should return default)
	result = configSvc.GetBool("NonExistent", true)
	if !result {
		t.Errorf("GetBool(NonExistent) expected default value true, got false")
	}
}

func TestGetFloat(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set test config
	testConfig := createTestConfig()
	configSvc.SetConfig(testConfig)

	// Test cases
	testCases := []struct {
		key      string
		expected float64
		default_ float64
	}{
		{"EasyBonus", testConfig.EasyBonus, 1.0},
		{"IntervalModifier", testConfig.IntervalModifier, 1.0},
		{"NonExistent", 3.14, 3.14},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := configSvc.GetFloat(tc.key, tc.default_)
			if result != tc.expected {
				t.Errorf("GetFloat(%s) expected %f, got %f", tc.key, tc.expected, result)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set test config
	testConfig := createTestConfig()
	err := configSvc.SetConfig(testConfig)
	if err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to exist after SaveConfig")
	}

	// Create a new config service with the same file to verify data was saved
	configSvc2, _ := NewYAMLConfig(configPath)
	config2, _ := configSvc2.GetConfig()

	if config2.Theme != testConfig.Theme {
		t.Errorf("expected theme '%s' to be saved, got '%s'", testConfig.Theme, config2.Theme)
	}
	if config2.EasyBonus != testConfig.EasyBonus {
		t.Errorf("expected easy bonus %f to be saved, got %f", testConfig.EasyBonus, config2.EasyBonus)
	}
}

func TestResetToDefaults(t *testing.T) {
	// Setup
	configPath, cleanup := createTempConfigFile(t, "")
	defer cleanup()
	configSvc, _ := NewYAMLConfig(configPath)

	// Create and set non-default config
	testConfig := createTestConfig()
	configSvc.SetConfig(testConfig)

	// Reset to defaults
	err := configSvc.ResetToDefaults()
	if err != nil {
		t.Errorf("ResetToDefaults() error = %v", err)
	}

	// Verify config was reset
	config, _ := configSvc.GetConfig()
	if config.Theme != DefaultTheme {
		t.Errorf("expected default theme '%s', got '%s'", DefaultTheme, config.Theme)
	}
	if config.EasyBonus != DefaultEasyBonus {
		t.Errorf("expected default easy bonus %f, got %f", DefaultEasyBonus, config.EasyBonus)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a config file with known content
	content := `
cardsdir: /custom/path
theme: custom-theme
codetheme: custom-code-theme
easybonus: 1.6
intervalmodifier: 1.3
newcardsperday: 30
maxinterval: 400
showlinenumbers: true
`
	configPath, cleanup := createTempConfigFile(t, content)
	defer cleanup()

	// Create yaml config and access its internal implementation
	configSvc, _ := NewYAMLConfig(configPath)
	yamlConfig, ok := configSvc.(*YAMLConfig)
	if !ok {
		t.Fatal("Expected *YAMLConfig type")
	}

	// Call loadConfig (testing private method via public interface)
	// We already tested it indirectly via NewYAMLConfig, but let's be thorough
	yamlConfig.config = interfaces.Config{} // Clear config
	err := yamlConfig.loadConfig()
	if err != nil {
		t.Errorf("loadConfig() error = %v", err)
	}

	// Verify values were loaded from file
	if yamlConfig.config.Theme != "custom-theme" {
		t.Errorf("expected theme 'custom-theme', got '%s'", yamlConfig.config.Theme)
	}
	if yamlConfig.config.CardsDir != "/custom/path" {
		t.Errorf("expected cards dir '/custom/path', got '%s'", yamlConfig.config.CardsDir)
	}
	if yamlConfig.config.EasyBonus != 1.6 {
		t.Errorf("expected easy bonus 1.6, got %f", yamlConfig.config.EasyBonus)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	// This function is private, so we'll test it indirectly
	config := getDefaultConfig()

	// Check some key default values
	if config.Theme != DefaultTheme {
		t.Errorf("expected default theme '%s', got '%s'", DefaultTheme, config.Theme)
	}
	if config.CodeTheme != DefaultCodeTheme {
		t.Errorf("expected default code theme '%s', got '%s'", DefaultCodeTheme, config.CodeTheme)
	}
	if config.EasyBonus != DefaultEasyBonus {
		t.Errorf("expected default easy bonus %f, got %f", DefaultEasyBonus, config.EasyBonus)
	}
	if config.MaxInterval != DefaultMaxInterval {
		t.Errorf("expected default max interval %d, got %d", DefaultMaxInterval, config.MaxInterval)
	}
	if config.ShowLineNumbers != DefaultShowLineNumbers {
		t.Errorf("expected default show line numbers %v, got %v", DefaultShowLineNumbers, config.ShowLineNumbers)
	}
}
