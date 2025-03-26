// internal/service/card/card_service_test.go
package card

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/storage"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// Basic test setup for the card service
func setupCardServiceTest(t *testing.T) (string, *DefaultCardService, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-card-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Set up the storage service
	storageService := storage.NewFileSystemStorage()
	if err := storageService.Initialize(tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to initialize storage: %v", err)
	}

	// Set up the algorithm
	alg := algorithm.NewSM2Algorithm()

	// Create the card service
	cardService := NewCardService(storageService, alg).(*DefaultCardService)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cardService, cleanup
}

// Create a sample card file for testing
func createSampleCardFile(tempDir string, filename string, content string) (string, error) {
	// Fix any newline issues and ensure proper formatting
	content = strings.ReplaceAll(content, "\r\n", "\n") // Normalize newlines
	if !strings.HasSuffix(content, "\n") {
		content += "\n" // Ensure file ends with newline
	}

	cardPath := filepath.Join(tempDir, filename)
	return cardPath, os.WriteFile(cardPath, []byte(content), 0644)
}

func TestGetCard(t *testing.T) {
	tempDir, cardService, cleanup := setupCardServiceTest(t)
	defer cleanup()

	// Create a sample card file
	cardContent := `---
title: Test Card
tags:
  - test
difficulty: 3
---
# Question

What is this test for?

---

To test the card service.
`

	cardPath, err := createSampleCardFile(tempDir, "test-card.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Test retrieving the card
	card, err := cardService.GetCard(cardPath)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	// Verify card properties
	if card.Title != "Test Card" {
		t.Errorf("expected title 'Test Card', got '%s'", card.Title)
	}

	if len(card.Tags) != 1 || card.Tags[0] != "test" {
		t.Errorf("expected tags [test], got %v", card.Tags)
	}

	if card.Difficulty != 3 {
		t.Errorf("expected difficulty 3, got %d", card.Difficulty)
	}
}

func TestReviewCard(t *testing.T) {
	tempDir, cardService, cleanup := setupCardServiceTest(t)
	defer cleanup()

	// Create a sample card file
	cardContent := `---
title: Test Card
tags:
  - test
difficulty: 3
---
# Question

What is this test for?

---

To test the card service.
`

	cardPath, err := createSampleCardFile(tempDir, "test-card.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Test reviewing the card
	err = cardService.ReviewCard(cardPath, 4)
	if err != nil {
		t.Fatalf("failed to review card: %v", err)
	}

	// Verify that the card was updated
	card, err := cardService.GetCard(cardPath)
	if err != nil {
		t.Fatalf("failed to get card after review: %v", err)
	}

	// Verify review data was updated
	if card.LastReviewed.IsZero() {
		t.Error("expected LastReviewed to be set")
	}

	if card.ReviewInterval != 2 {
		t.Errorf("expected ReviewInterval to be 2 for a new card with rating 4, got %d", card.ReviewInterval)
	}

	// The difficulty should have been updated based on the rating (5-4=1)
	if card.Difficulty != 1 {
		t.Errorf("expected Difficulty to be updated to 1, got %d", card.Difficulty)
	}
}

func TestIsDue(t *testing.T) {
	tempDir, cardService, cleanup := setupCardServiceTest(t)
	defer cleanup()

	// Create a sample card file with review data
	cardContent := `---
title: Test Card
tags:
  - test
difficulty: 3
last_reviewed: 2023-01-01
review_interval: 365
---
# Question

What is this test for?

---

To test the card service.
`

	cardPath, err := createSampleCardFile(tempDir, "due-card.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Card should be due since last_reviewed is in the past
	isDue := cardService.IsDue(cardPath)
	if !isDue {
		t.Error("expected card to be due")
	}

	// Update the card to be reviewed today
	err = cardService.ReviewCard(cardPath, 5)
	if err != nil {
		t.Fatalf("failed to review card: %v", err)
	}

	// Card should no longer be due
	isDue = cardService.IsDue(cardPath)
	if isDue {
		t.Error("expected card to not be due after review")
	}
}

// Test for GetDueDate - using direct cache manipulation to avoid YAML parsing issues
func TestGetDueDate(t *testing.T) {
	tempDir, cardService, cleanup := setupCardServiceTest(t)
	defer cleanup()

	// Create a simple test card for the file system
	cardContent := `---
title: Due Date Test Card
---
Question?
---
Answer
`

	cardPath, err := createSampleCardFile(tempDir, "due-date-card.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Create a test card with known dates and put it directly in the cache
	testCard := domain.Card{
		FilePath:       cardPath,
		Title:          "Due Date Test Card",
		LastReviewed:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ReviewInterval: 30,
	}

	// Force the card into the storage cache
	storage := cardService.storage.(*storage.FileSystemStorage)
	storage.ForceCardIntoCache(testCard)

	// Get the due date
	dueDate := cardService.GetDueDate(cardPath)

	// Expected: 30 days after 2025-01-01 = 2025-01-31
	expectedDate := "2025-01-31"
	actualDate := dueDate.Format("2006-01-02")

	if actualDate != expectedDate {
		t.Errorf("expected due date %s, got %s", expectedDate, actualDate)
	}

	// Test with non-existent card
	nonExistentDueDate := cardService.GetDueDate(filepath.Join(tempDir, "non-existent.md"))
	if !nonExistentDueDate.IsZero() {
		t.Errorf("expected zero time for non-existent card, got %v", nonExistentDueDate)
	}
}
