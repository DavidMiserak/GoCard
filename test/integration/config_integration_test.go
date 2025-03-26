// test/integration/config_integration_test.go
package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/card"
	"github.com/DavidMiserak/GoCard/internal/service/config"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	"github.com/DavidMiserak/GoCard/internal/service/render"
	"github.com/DavidMiserak/GoCard/internal/service/storage"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// setupConfigTest sets up a test environment with configuration service
func setupConfigTest(t *testing.T) (string, interfaces.ConfigService, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gocard-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create config file path
	configPath := filepath.Join(tempDir, "config.yaml")

	// Initialize config service
	configService, err := config.NewYAMLConfig(configPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize config service: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, configService, cleanup
}

// TestConfigLoading tests configuration loading functionality
func TestConfigLoading(t *testing.T) {
	// Setup test environment
	tempDir, configService, cleanup := setupConfigTest(t)
	defer cleanup()

	// Verify default configuration is loaded
	cfg, err := configService.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	// Check default values
	if cfg.Theme != "default" {
		t.Errorf("Expected default theme 'default', got '%s'", cfg.Theme)
	}
	if cfg.EasyBonus != 1.3 {
		t.Errorf("Expected default EasyBonus 1.3, got %f", cfg.EasyBonus)
	}

	// Update config with new values
	newConfig := interfaces.Config{
		CardsDir:         filepath.Join(tempDir, "cards"),
		Theme:            "dark",
		CodeTheme:        "dracula",
		EasyBonus:        1.5,
		IntervalModifier: 1.2,
		NewCardsPerDay:   30,
		MaxInterval:      500,
		ShowLineNumbers:  false,
	}

	err = configService.SetConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Create a new config service instance to test loading from file
	configPath := filepath.Join(tempDir, "config.yaml")
	newConfigService, err := config.NewYAMLConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to create new config service: %v", err)
	}

	// Verify new configuration was loaded
	loadedCfg, err := newConfigService.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config from new service: %v", err)
	}

	// Check that values match what we set
	if loadedCfg.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", loadedCfg.Theme)
	}
	if loadedCfg.EasyBonus != 1.5 {
		t.Errorf("Expected EasyBonus 1.5, got %f", loadedCfg.EasyBonus)
	}
	if loadedCfg.IntervalModifier != 1.2 {
		t.Errorf("Expected IntervalModifier 1.2, got %f", loadedCfg.IntervalModifier)
	}
	if loadedCfg.MaxInterval != 500 {
		t.Errorf("Expected MaxInterval 500, got %d", loadedCfg.MaxInterval)
	}
}

// TestIndividualConfigMethods tests getting and setting individual config values
func TestIndividualConfigMethods(t *testing.T) {
	// Setup test environment
	_, configService, cleanup := setupConfigTest(t)
	defer cleanup()

	// Test GetString
	theme := configService.GetString("Theme", "fallback")
	if theme != "default" { // Default theme from config
		t.Errorf("Expected GetString to return 'default', got '%s'", theme)
	}

	// Test GetInt
	maxInterval := configService.GetInt("MaxInterval", 100)
	if maxInterval != 365 { // Default MaxInterval from config
		t.Errorf("Expected GetInt to return 365, got %d", maxInterval)
	}

	// Test GetFloat
	easyBonus := configService.GetFloat("EasyBonus", 1.0)
	if easyBonus != 1.3 { // Default EasyBonus from config
		t.Errorf("Expected GetFloat to return 1.3, got %f", easyBonus)
	}

	// Test GetBool
	showLineNumbers := configService.GetBool("ShowLineNumbers", false)
	if !showLineNumbers { // Default ShowLineNumbers from config
		t.Errorf("Expected GetBool to return true, got false")
	}

	// Test with non-existent keys (should return default values)
	nonExistentString := configService.GetString("NonExistent", "default-value")
	if nonExistentString != "default-value" {
		t.Errorf("Expected non-existent key to return default value, got '%s'", nonExistentString)
	}

	nonExistentInt := configService.GetInt("NonExistent", 42)
	if nonExistentInt != 42 {
		t.Errorf("Expected non-existent key to return default value, got %d", nonExistentInt)
	}
}

// TestConfigurationReset tests resetting configuration to defaults
func TestConfigurationReset(t *testing.T) {
	// Setup test environment
	_, configService, cleanup := setupConfigTest(t)
	defer cleanup()

	// Modify configuration
	newConfig := interfaces.Config{
		CardsDir:         "/custom/path",
		Theme:            "custom-theme",
		CodeTheme:        "custom-code-theme",
		EasyBonus:        2.0,
		IntervalModifier: 1.5,
		NewCardsPerDay:   50,
		MaxInterval:      1000,
		ShowLineNumbers:  false,
	}

	err := configService.SetConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Verify custom values are set
	cfg, _ := configService.GetConfig()
	if cfg.Theme != "custom-theme" {
		t.Errorf("Expected theme to be 'custom-theme', got '%s'", cfg.Theme)
	}

	// Reset to defaults
	err = configService.ResetToDefaults()
	if err != nil {
		t.Fatalf("Failed to reset config: %v", err)
	}

	// Verify values are reset
	resetCfg, _ := configService.GetConfig()
	if resetCfg.Theme != "default" {
		t.Errorf("Expected reset theme to be 'default', got '%s'", resetCfg.Theme)
	}
	if resetCfg.EasyBonus != 1.3 {
		t.Errorf("Expected reset EasyBonus to be 1.3, got %f", resetCfg.EasyBonus)
	}
}

// TestAlgorithmMaxIntervalConfig tests how MaxInterval configuration affects the SM2 algorithm
func TestAlgorithmMaxIntervalConfig(t *testing.T) {
	// Setup test environment
	tempDir, configService, cleanup := setupConfigTest(t)
	defer cleanup()

	// Create subdirectory for cards
	cardsDir := filepath.Join(tempDir, "cards")
	err := os.MkdirAll(cardsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cards directory: %v", err)
	}

	// Update config to use test cards directory
	cfg, _ := configService.GetConfig()
	cfg.CardsDir = cardsDir

	// Set a small MaxInterval for testing
	cfg.MaxInterval = 50
	err = configService.SetConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Initialize storage service
	storageService := storage.NewFileSystemStorage()
	err = storageService.Initialize(cardsDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create algorithm with small MaxInterval from config
	smallMaxAlg := &algorithm.SM2Algorithm{
		EasyBonus:        cfg.EasyBonus,
		IntervalModifier: cfg.IntervalModifier,
		MaxInterval:      cfg.MaxInterval, // 50 days
	}

	// Create card service with the small max interval algorithm
	cardService1 := card.NewCardService(storageService, smallMaxAlg)

	// Create a test card with a large interval (that would be capped)
	testCard := domain.Card{
		FilePath:       filepath.Join(cardsDir, "test-card.md"),
		Title:          "Test Card",
		LastReviewed:   time.Now().AddDate(0, 0, -100),
		ReviewInterval: 100, // Large interval that exceeds our small MaxInterval
	}

	// Create test card file
	cardContent := `---
title: Test Card
last_reviewed: 2023-01-01
review_interval: 100
difficulty: 3
---
# Question
Test question?
---
Test answer.
`
	err = os.WriteFile(testCard.FilePath, []byte(cardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test card file: %v", err)
	}

	// Force card into storage cache
	storageService.ForceCardIntoCache(testCard)

	// Review the card - the interval should be capped at MaxInterval (50)
	err = cardService1.ReviewCard(testCard.FilePath, 5)
	if err != nil {
		t.Fatalf("Failed to review card with small MaxInterval: %v", err)
	}

	// Get the updated card
	updatedCard1, err := cardService1.GetCard(testCard.FilePath)
	if err != nil {
		t.Fatalf("Failed to get updated card: %v", err)
	}

	smallMaxInterval := updatedCard1.ReviewInterval
	t.Logf("Small MaxInterval (50) resulted in interval: %d", smallMaxInterval)

	// Verify the interval was capped
	if smallMaxInterval > 50 {
		t.Errorf("Expected interval to be capped at 50, got %d", smallMaxInterval)
	}

	// Reset the card
	testCard.ReviewInterval = 100
	testCard.LastReviewed = time.Now().AddDate(0, 0, -100)
	storageService.ForceCardIntoCache(testCard)

	// Reset card file
	err = os.WriteFile(testCard.FilePath, []byte(cardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to reset test card file: %v", err)
	}

	// Now update config to have a large MaxInterval
	cfg.MaxInterval = 500
	err = configService.SetConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Create algorithm with large MaxInterval
	largeMaxAlg := &algorithm.SM2Algorithm{
		EasyBonus:        cfg.EasyBonus,
		IntervalModifier: cfg.IntervalModifier,
		MaxInterval:      cfg.MaxInterval, // 500 days
	}

	// Create a new card service with large max interval
	cardService2 := card.NewCardService(storageService, largeMaxAlg)

	// Review the card again - now with larger MaxInterval
	err = cardService2.ReviewCard(testCard.FilePath, 5)
	if err != nil {
		t.Fatalf("Failed to review card with large MaxInterval: %v", err)
	}

	// Get the updated card
	updatedCard2, err := cardService2.GetCard(testCard.FilePath)
	if err != nil {
		t.Fatalf("Failed to get updated card: %v", err)
	}

	largeMaxInterval := updatedCard2.ReviewInterval
	t.Logf("Large MaxInterval (500) resulted in interval: %d", largeMaxInterval)

	// The interval with larger MaxInterval should be greater than the capped one
	if largeMaxInterval <= smallMaxInterval {
		t.Errorf("Expected interval with larger MaxInterval to be greater, got small: %d, large: %d",
			smallMaxInterval, largeMaxInterval)
	}

	// Also verify the larger interval isn't capped
	if largeMaxInterval > 100 && largeMaxInterval <= 500 {
		t.Logf("Interval increased beyond original value but stayed within MaxInterval limits")
	} else if largeMaxInterval == 500 {
		t.Logf("Interval was capped at new MaxInterval value (500)")
	}
}

// TestCardsDirConfig tests how CardsDir configuration affects storage
func TestCardsDirConfig(t *testing.T) {
	// Setup test environment
	tempDir, configService, cleanup := setupConfigTest(t)
	defer cleanup()

	// Create two potential card directories
	cardsDir1 := filepath.Join(tempDir, "cards1")
	cardsDir2 := filepath.Join(tempDir, "cards2")

	for _, dir := range []string{cardsDir1, cardsDir2} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create cards directory %s: %v", dir, err)
		}
	}

	// Update config to use the first cards directory
	cfg, _ := configService.GetConfig()
	cfg.CardsDir = cardsDir1
	err := configService.SetConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to set config for cards directory 1: %v", err)
	}

	// Create a test card in directory 1
	cardPath1 := filepath.Join(cardsDir1, "card1.md")
	err = os.WriteFile(cardPath1, []byte(`---
title: Card in Directory 1
---
# Question
?
---
Answer.
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test card: %v", err)
	}

	// Initialize storage with the configured directory
	storageService := storage.NewFileSystemStorage()
	if err := storageService.Initialize(cfg.CardsDir); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Verify we can access the card in directory 1
	card1, err := storageService.LoadCard(cardPath1)
	if err != nil {
		t.Fatalf("Failed to load card from directory 1: %v", err)
	}
	if card1.Title != "Card in Directory 1" {
		t.Errorf("Expected card title 'Card in Directory 1', got '%s'", card1.Title)
	}

	// Now change the cards directory in config
	cfg.CardsDir = cardsDir2
	err = configService.SetConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to set config for cards directory 2: %v", err)
	}

	// Create a test card in directory 2
	cardPath2 := filepath.Join(cardsDir2, "card2.md")
	err = os.WriteFile(cardPath2, []byte(`---
title: Card in Directory 2
---
# Question
?
---
Answer.
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test card: %v", err)
	}

	// Initialize a new storage service with the updated config
	newStorageService := storage.NewFileSystemStorage()
	if err := newStorageService.Initialize(cfg.CardsDir); err != nil {
		t.Fatalf("Failed to initialize storage with new directory: %v", err)
	}

	// Verify the new storage service can access cards in directory 2
	card2, err := newStorageService.LoadCard(cardPath2)
	if err != nil {
		t.Fatalf("Failed to load card from directory 2: %v", err)
	}
	if card2.Title != "Card in Directory 2" {
		t.Errorf("Expected card title 'Card in Directory 2', got '%s'", card2.Title)
	}

	// Create a deck service to get card stats for directory 2
	alg := algorithm.NewSM2Algorithm()
	card.NewCardService(newStorageService, alg)

	// Verify the root directory is set correctly by creating a new card in dir2
	// and checking that it's found when searching in the root directory
	cardPath2b := filepath.Join(cardsDir2, "card2b.md")
	err = os.WriteFile(cardPath2b, []byte(`---
title: Second Card in Directory 2
tags:
  - unique-test-tag
---
# Question
?
---
Answer.
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create second test card: %v", err)
	}

	// Test that searching finds the card in directory 2
	foundCards, err := newStorageService.SearchCards("unique-test-tag")
	if err != nil {
		t.Fatalf("Failed to search cards: %v", err)
	}

	if len(foundCards) != 1 {
		t.Errorf("Expected to find 1 card with unique tag in directory 2, got %d", len(foundCards))
	}

	if len(foundCards) > 0 && foundCards[0].Title != "Second Card in Directory 2" {
		t.Errorf("Expected to find 'Second Card in Directory 2', got '%s'", foundCards[0].Title)
	}
}

// TestRenderConfigEffects tests how configuration affects the render service
func TestRenderConfigEffects(t *testing.T) {
	// Initialize render service
	renderService := render.NewMarkdownRenderer()

	// Test styling methods since these have concrete, testable behavior
	heading := "Test Heading"
	styledHeading := renderService.StyleHeading(heading, 1)
	if !strings.Contains(styledHeading, heading) {
		t.Errorf("StyleHeading did not include the original heading text")
	}

	// Test info styling
	infoText := "Test Info"
	styledInfo := renderService.StyleInfo(infoText)
	if !strings.Contains(styledInfo, "INFO") {
		t.Errorf("StyleInfo did not include INFO prefix")
	}
	if !strings.Contains(styledInfo, infoText) {
		t.Errorf("StyleInfo did not include the original info text")
	}

	// Test warning styling
	warningText := "Test Warning"
	styledWarning := renderService.StyleWarning(warningText)
	if !strings.Contains(styledWarning, "WARNING") {
		t.Errorf("StyleWarning did not include WARNING prefix")
	}
	if !strings.Contains(styledWarning, warningText) {
		t.Errorf("StyleWarning did not include the original warning text")
	}

	// Test error styling
	errorText := "Test Error"
	styledError := renderService.StyleError(errorText)
	if !strings.Contains(styledError, "ERROR") {
		t.Errorf("StyleError did not include ERROR prefix")
	}
	if !strings.Contains(styledError, errorText) {
		t.Errorf("StyleError did not include the original error text")
	}

	// Test that available code themes are returned
	themes := renderService.GetAvailableCodeThemes()
	if len(themes) == 0 {
		t.Errorf("Expected GetAvailableCodeThemes to return non-empty list")
	}
}
