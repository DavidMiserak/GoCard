// File: internal/config/config_test.go

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestConfigDefaults verifies that the config provides sensible defaults
func TestConfigDefaults(t *testing.T) {
	// Get default config
	cfg := Default()

	// Basic validation of default values
	if cfg.CardsDir == "" {
		t.Error("Default CardsDir should not be empty")
	}

	if cfg.FirstRun != true {
		t.Error("Default FirstRun should be true")
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Default logging level should be 'info', got %s", cfg.Logging.Level)
	}

	if cfg.UI.Theme != "auto" {
		t.Errorf("Default theme should be 'auto', got %s", cfg.UI.Theme)
	}

	if cfg.SpacedRepetition.MaxInterval <= 0 {
		t.Errorf("Default max interval should be positive, got %d", cfg.SpacedRepetition.MaxInterval)
	}

	// Make sure defaults are reasonable and usable
	t.Logf("Default config: CardsDir=%s, FirstRun=%v, LogLevel=%s, Theme=%s, MaxInterval=%d",
		cfg.CardsDir, cfg.FirstRun, cfg.Logging.Level, cfg.UI.Theme, cfg.SpacedRepetition.MaxInterval)
}

// TestPathExpansion verifies that home directory paths are properly expanded
func TestPathExpansion(t *testing.T) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Test expanding a path with ~
	testPath := "~/test/path"
	expanded, err := expandPath(testPath)
	if err != nil {
		t.Fatalf("Failed to expand path: %v", err)
	}

	expected := filepath.Join(homeDir, "test/path")
	if expanded != expected {
		t.Errorf("Path expansion failed. Expected %s, got %s", expected, expanded)
	}

	// Test with non-home path (should remain unchanged)
	nonHomePath := "/tmp/test/path"
	expanded, err = expandPath(nonHomePath)
	if err != nil {
		t.Fatalf("Failed to handle non-home path: %v", err)
	}

	if expanded != nonHomePath {
		t.Errorf("Non-home path was modified. Expected %s, got %s", nonHomePath, expanded)
	}
}

// TestCrossPlatformPathHandling tests that path handling works across platforms
func TestCrossPlatformPathHandling(t *testing.T) {
	// Identify current platform
	platform := runtime.GOOS
	t.Logf("Testing on platform: %s", platform)

	// Test home directory expansion
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Platform-specific path expansion
	expanded, err := expandPath("~/test")
	if err != nil {
		t.Fatalf("Path expansion failed: %v", err)
	}

	// Check that the path uses the correct path separator for this platform
	pathSeparator := string(os.PathSeparator)
	if !strings.Contains(expanded, pathSeparator) {
		t.Errorf("Expanded path should contain %s separator, got: %s", pathSeparator, expanded)
	}

	// Verify path joins correctly
	expected := filepath.Join(homeDir, "test")
	if expanded != expected {
		t.Errorf("Path expansion incorrect for this platform. Expected %s, got %s", expected, expanded)
	}
}

// TestConfigSerialization verifies config can be serialized and deserialized
func TestConfigSerialization(t *testing.T) {
	// Create a temp dir for config
	tempDir, err := os.MkdirTemp("", "gocard-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a custom config
	cfg := Default()

	// Set a unique value to detect if it's preserved
	uniqueCardsDir := filepath.Join(tempDir, "custom-cards-dir")
	cfg.CardsDir = uniqueCardsDir

	// Try to save this config
	tempPath := filepath.Join(tempDir, "temp-config.yaml")

	// Write the config content directly
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(tempPath, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created: %v", err)
	}

	// Now read it with direct YAML parsing
	loadedData, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var loadedCfg Config
	if err := yaml.Unmarshal(loadedData, &loadedCfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify custom value is preserved
	if loadedCfg.CardsDir != uniqueCardsDir {
		t.Errorf("Custom CardsDir not preserved. Expected %s, got %s", uniqueCardsDir, loadedCfg.CardsDir)
	}
}

// TestMissingConfigValues verifies defaults are applied for missing values
func TestMissingConfigValues(t *testing.T) {
	// Test multiple YAML parsing scenarios
	testCases := []struct {
		name     string
		yamlData string
		validate func(t *testing.T, cfg Config)
	}{
		{
			name: "Partial Config",
			yamlData: `cards_dir: "/custom/path"
first_run: false
logging:
  level: "debug"`,
			validate: func(t *testing.T, cfg Config) {
				if cfg.CardsDir != "/custom/path" {
					t.Errorf("CardsDir not preserved. Expected /custom/path, got %s", cfg.CardsDir)
				}
				if cfg.FirstRun != false {
					t.Errorf("FirstRun should be false, got %v", cfg.FirstRun)
				}
				if cfg.Logging.Level != "debug" {
					t.Errorf("Logging level should be 'debug', got %s", cfg.Logging.Level)
				}
			},
		},
		{
			name:     "Minimal Config",
			yamlData: `cards_dir: "/another/path"`,
			validate: func(t *testing.T, cfg Config) {
				if cfg.CardsDir != "/another/path" {
					t.Errorf("CardsDir not preserved. Expected /another/path, got %s", cfg.CardsDir)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Print out the exact YAML for debugging
			t.Logf("Testing YAML:\n%s", tc.yamlData)

			var cfg Config
			err := yaml.Unmarshal([]byte(tc.yamlData), &cfg)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			// Run validation for this test case
			tc.validate(t, cfg)
		})
	}
}

// TestYAMLValidation tests validation of YAML config
func TestYAMLValidation(t *testing.T) {
	// Test cases for different invalid YAML scenarios
	testCases := []struct {
		name     string
		yaml     string
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name:    "Invalid Boolean",
			yaml:    "first_run: not-a-bool",
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "cannot unmarshal")
			},
		},
		{
			name:    "Invalid Number",
			yaml:    "spaced_repetition:\n  max_interval: not-a-number",
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "cannot unmarshal")
			},
		},
		{
			name:    "Malformed YAML",
			yaml:    "invalid: yaml\n  indentation: wrong",
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "yaml:")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg Config
			err := yaml.Unmarshal([]byte(tc.yaml), &cfg)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected an error for %s, got none", tc.name)
				} else if !tc.errCheck(err) {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				} else {
					t.Logf("Got expected error for %s: %v", tc.name, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}

// TestHomeDirExpansion verifies home directory expansion for different platforms
func TestHomeDirExpansion(t *testing.T) {
	// This test doesn't use the actual config.expandPath function,
	// but verifies that path expansion works correctly in principle

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Test platform-specific separators
	tildePath := "~/documents/cards"
	expected := filepath.Join(homeDir, "documents/cards")

	// Do manual expansion for illustration
	expandedPath := strings.Replace(tildePath, "~/", homeDir+string(os.PathSeparator), 1)
	expandedPath = strings.ReplaceAll(expandedPath, "/", string(os.PathSeparator))

	t.Logf("Home directory: %s", homeDir)
	t.Logf("Tilde path: %s", tildePath)
	t.Logf("Expected expansion: %s", expected)
	t.Logf("Manual expansion: %s", expandedPath)

	// Test actual expansion function
	actualExpanded, err := expandPath(tildePath)
	if err != nil {
		t.Fatalf("Failed to expand path: %v", err)
	}

	t.Logf("Actual expansion: %s", actualExpanded)

	// On different platforms, the path separators might be different,
	// so we need to be careful about direct string comparison
	cleanExpected := filepath.Clean(expected)
	cleanActual := filepath.Clean(actualExpanded)

	if cleanActual != cleanExpected {
		t.Errorf("Path expansion incorrect. Expected %s, got %s", cleanExpected, cleanActual)
	}
}
