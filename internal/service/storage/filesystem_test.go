// internal/service/storage/filesystem_test.go
package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/domain"
)

// Setup helper to create a test environment with a temporary directory
func setupFileSystemTest(t *testing.T) (*FileSystemStorage, string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-fs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a new storage service
	storage := NewFileSystemStorage()
	if err := storage.Initialize(tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to initialize storage: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		storage.Close()
		os.RemoveAll(tempDir)
	}

	return storage, tempDir, cleanup
}

// Helper to create a sample card file
func createSampleCardFile(dir, filename, content string) (string, error) {
	filePath := filepath.Join(dir, filename)
	return filePath, os.WriteFile(filePath, []byte(content), 0644)
}

// Helper to create a sample deck structure
func createSampleDeckStructure(baseDir string) error {
	// Create deck directories
	deckPaths := []string{
		filepath.Join(baseDir, "Programming"),
		filepath.Join(baseDir, "Programming", "Go"),
		filepath.Join(baseDir, "Languages"),
		filepath.Join(baseDir, "Languages", "Spanish"),
	}

	for _, path := range deckPaths {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// Create sample cards
	cardContents := map[string]string{
		filepath.Join(baseDir, "Programming", "Go", "concurrency.md"): `---
title: Go Concurrency
tags:
  - go
  - programming
  - concurrency
difficulty: 2
last_reviewed: 2023-01-01
review_interval: 7
---
# What is the difference between a goroutine and a thread?

---

Goroutines are lighter weight than threads.
`,
		filepath.Join(baseDir, "Languages", "Spanish", "verbs.md"): `---
title: Spanish Verbs
tags:
  - spanish
  - language
difficulty: 3
---
# What is the conjugation of "hablar" in present tense?

---

yo hablo
tú hablas
él/ella/usted habla
`,
	}

	for path, content := range cardContents {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func TestClose(t *testing.T) {
	// Setup
	fs, _, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Populate caches
	fs.cardCache["test"] = domain.Card{Title: "Test Card"}
	fs.deckCache["test"] = domain.Deck{Name: "Test Deck"}

	// Call Close
	err := fs.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify caches are cleared
	if len(fs.cardCache) != 0 {
		t.Errorf("expected cardCache to be empty after Close()")
	}
	if len(fs.deckCache) != 0 {
		t.Errorf("expected deckCache to be empty after Close()")
	}
}

func TestLoadCard(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample card file
	cardContent := `---
title: Test Card
tags:
  - test
  - sample
difficulty: 3
last_reviewed: 2023-01-15
review_interval: 14
---
# Test Question

This is a test question.

---

This is the answer.
`
	cardPath, err := createSampleCardFile(tempDir, "test-card.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Test loading the card
	card, err := fs.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("LoadCard() error = %v", err)
	}

	// Verify card properties
	if card.Title != "Test Card" {
		t.Errorf("expected Title to be 'Test Card', got '%s'", card.Title)
	}

	if len(card.Tags) != 2 || card.Tags[0] != "test" || card.Tags[1] != "sample" {
		t.Errorf("expected Tags to be [test sample], got %v", card.Tags)
	}

	if card.Difficulty != 3 {
		t.Errorf("expected Difficulty to be 3, got %d", card.Difficulty)
	}

	if card.ReviewInterval != 14 {
		t.Errorf("expected ReviewInterval to be 14, got %d", card.ReviewInterval)
	}

	// Check question and answer extraction
	if !strings.Contains(card.Question, "Test Question") {
		t.Errorf("expected Question to contain 'Test Question', got '%s'", card.Question)
	}

	if !strings.Contains(card.Answer, "This is the answer") {
		t.Errorf("expected Answer to contain 'This is the answer', got '%s'", card.Answer)
	}

	// Test caching
	// Load the card again - should use cache
	cachedCard, err := fs.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("LoadCard() from cache error = %v", err)
	}

	if cachedCard.Title != card.Title {
		t.Errorf("cache returned different card: expected '%s', got '%s'", card.Title, cachedCard.Title)
	}

	// Test non-existent card
	_, err = fs.LoadCard(filepath.Join(tempDir, "non-existent.md"))
	if err == nil {
		t.Error("expected error loading non-existent card, got nil")
	}
}

func TestUpdateCardMetadata(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
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

To test UpdateCardMetadata.
`
	cardPath, err := createSampleCardFile(tempDir, "update-test.md", cardContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Load the card
	card, err := fs.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("failed to load card: %v", err)
	}

	// Modify card metadata
	card.LastReviewed = card.LastReviewed.AddDate(0, 0, 1) // Add a day
	card.ReviewInterval = 21
	card.Difficulty = 2

	// Update the card
	err = fs.UpdateCardMetadata(card)
	if err != nil {
		t.Fatalf("UpdateCardMetadata() error = %v", err)
	}

	// Load the card again to verify changes
	updatedCard, err := fs.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("failed to load updated card: %v", err)
	}

	// Verify changes
	if updatedCard.ReviewInterval != 21 {
		t.Errorf("expected updated ReviewInterval to be 21, got %d", updatedCard.ReviewInterval)
	}

	if updatedCard.Difficulty != 2 {
		t.Errorf("expected updated Difficulty to be 2, got %d", updatedCard.Difficulty)
	}

	// Check that the last_reviewed date was updated in the file
	content, err := os.ReadFile(cardPath)
	if err != nil {
		t.Fatalf("failed to read updated card file: %v", err)
	}

	if !strings.Contains(string(content), "last_reviewed:") {
		t.Error("expected updated file to contain 'last_reviewed' field")
	}

	if !strings.Contains(string(content), "review_interval: 21") {
		t.Error("expected updated file to contain 'review_interval: 21'")
	}
}

func TestListCardPaths(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample structure
	if err := createSampleDeckStructure(tempDir); err != nil {
		t.Fatalf("failed to create sample deck structure: %v", err)
	}

	// Test listing card paths in a directory
	goDeckPath := filepath.Join(tempDir, "Programming", "Go")
	cardPaths, err := fs.ListCardPaths(goDeckPath)
	if err != nil {
		t.Fatalf("ListCardPaths() error = %v", err)
	}

	// Verify results
	if len(cardPaths) != 1 {
		t.Errorf("expected 1 card path, got %d", len(cardPaths))
	}

	expectedPath := filepath.Join(goDeckPath, "concurrency.md")
	if len(cardPaths) > 0 && cardPaths[0] != expectedPath {
		t.Errorf("expected card path %s, got %s", expectedPath, cardPaths[0])
	}

	// Test listing from directory with no cards
	emptyDirPath := filepath.Join(tempDir, "Programming", "Empty")
	if err := os.MkdirAll(emptyDirPath, 0755); err != nil {
		t.Fatalf("failed to create empty directory: %v", err)
	}

	emptyPaths, err := fs.ListCardPaths(emptyDirPath)
	if err != nil {
		t.Fatalf("ListCardPaths() on empty dir error = %v", err)
	}

	if len(emptyPaths) != 0 {
		t.Errorf("expected 0 card paths in empty dir, got %d", len(emptyPaths))
	}

	// Test with non-existent directory
	_, err = fs.ListCardPaths(filepath.Join(tempDir, "NonExistent"))
	if err == nil {
		t.Error("expected error listing from non-existent dir, got nil")
	}
}

func TestLoadDeck(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample structure
	if err := createSampleDeckStructure(tempDir); err != nil {
		t.Fatalf("failed to create sample deck structure: %v", err)
	}

	// Test loading a deck
	goDeckPath := filepath.Join(tempDir, "Programming", "Go")
	deck, err := fs.LoadDeck(goDeckPath)
	if err != nil {
		t.Fatalf("LoadDeck() error = %v", err)
	}

	// Verify deck properties
	if deck.Path != goDeckPath {
		t.Errorf("expected Path to be %s, got %s", goDeckPath, deck.Path)
	}

	if deck.Name != "Go" {
		t.Errorf("expected Name to be 'Go', got '%s'", deck.Name)
	}

	parentPath := filepath.Join(tempDir, "Programming")
	if deck.ParentPath != parentPath {
		t.Errorf("expected ParentPath to be %s, got %s", parentPath, deck.ParentPath)
	}

	// Test caching
	// Load the deck again - should use cache
	cachedDeck, err := fs.LoadDeck(goDeckPath)
	if err != nil {
		t.Fatalf("LoadDeck() from cache error = %v", err)
	}

	if cachedDeck.Name != deck.Name {
		t.Errorf("cache returned different deck: expected '%s', got '%s'", deck.Name, cachedDeck.Name)
	}

	// Test with non-existent directory
	_, err = fs.LoadDeck(filepath.Join(tempDir, "NonExistent"))
	if err == nil {
		t.Error("expected error loading non-existent deck, got nil")
	}

	// Test with a file instead of a directory
	filePath := filepath.Join(tempDir, "test-file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err = fs.LoadDeck(filePath)
	if err == nil {
		t.Error("expected error loading file as deck, got nil")
	}
}

func TestListDeckPaths(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample structure
	if err := createSampleDeckStructure(tempDir); err != nil {
		t.Fatalf("failed to create sample deck structure: %v", err)
	}

	// Test listing subdecks
	programmingPath := filepath.Join(tempDir, "Programming")
	deckPaths, err := fs.ListDeckPaths(programmingPath)
	if err != nil {
		t.Fatalf("ListDeckPaths() error = %v", err)
	}

	// Verify results
	if len(deckPaths) != 1 {
		t.Errorf("expected 1 subdeck in Programming, got %d", len(deckPaths))
	}

	expectedPath := filepath.Join(programmingPath, "Go")
	if len(deckPaths) > 0 && deckPaths[0] != expectedPath {
		t.Errorf("expected deck path %s, got %s", expectedPath, deckPaths[0])
	}

	// Test with non-existent directory
	_, err = fs.ListDeckPaths(filepath.Join(tempDir, "NonExistent"))
	if err == nil {
		t.Error("expected error listing from non-existent dir, got nil")
	}
}

func TestFindCardsByTag(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample structure
	if err := createSampleDeckStructure(tempDir); err != nil {
		t.Fatalf("failed to create sample deck structure: %v", err)
	}

	// Set root directory for search
	fs.rootDir = tempDir

	// Test finding cards by tag
	cards, err := fs.FindCardsByTag("go")
	if err != nil {
		t.Fatalf("FindCardsByTag() error = %v", err)
	}

	// Verify results
	if len(cards) != 1 {
		t.Errorf("expected 1 card with tag 'go', got %d", len(cards))
	}

	if len(cards) > 0 && cards[0].Title != "Go Concurrency" {
		t.Errorf("expected card with title 'Go Concurrency', got '%s'", cards[0].Title)
	}

	// Test with non-existent tag
	nonExistentCards, err := fs.FindCardsByTag("nonexistent")
	if err != nil {
		t.Fatalf("FindCardsByTag() with non-existent tag error = %v", err)
	}

	if len(nonExistentCards) != 0 {
		t.Errorf("expected 0 cards with non-existent tag, got %d", len(nonExistentCards))
	}
}

func TestSearchCards(t *testing.T) {
	// Setup
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create a sample structure
	if err := createSampleDeckStructure(tempDir); err != nil {
		t.Fatalf("failed to create sample deck structure: %v", err)
	}

	// Set root directory for search
	fs.rootDir = tempDir

	// Test searching for cards
	cards, err := fs.SearchCards("goroutine")
	if err != nil {
		t.Fatalf("SearchCards() error = %v", err)
	}

	// Verify results
	if len(cards) != 1 {
		t.Errorf("expected 1 card matching 'goroutine', got %d", len(cards))
	}

	if len(cards) > 0 && cards[0].Title != "Go Concurrency" {
		t.Errorf("expected card with title 'Go Concurrency', got '%s'", cards[0].Title)
	}

	// Test with empty query (should error)
	_, err = fs.SearchCards("")
	if err == nil {
		t.Error("expected error with empty search query, got nil")
	}

	// Test with non-matching query
	nonMatchingCards, err := fs.SearchCards("nonmatching")
	if err != nil {
		t.Fatalf("SearchCards() with non-matching query error = %v", err)
	}

	if len(nonMatchingCards) != 0 {
		t.Errorf("expected 0 cards matching non-matching query, got %d", len(nonMatchingCards))
	}
}
