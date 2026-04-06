package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

const testCardContent = `---
tags: [go,test]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Question

What is a goroutine?

## Answer

A goroutine is a lightweight thread managed by the Go runtime.
`

func TestSaveCardToMarkdown(t *testing.T) {
	// Create temp file with initial content
	tempDir := t.TempDir()
	cardPath := filepath.Join(tempDir, "test-card.md")
	if err := os.WriteFile(cardPath, []byte(testCardContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	store := &Store{}
	card := model.Card{
		ID:           cardPath,
		DeckID:       tempDir,
		Interval:     3,
		Ease:         2.5,
		LastReviewed: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
	}

	if err := store.SaveCardToMarkdown(card); err != nil {
		t.Fatalf("SaveCardToMarkdown error: %v", err)
	}

	// Re-read and verify
	content, err := os.ReadFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "review_interval: 3") {
		t.Errorf("Expected review_interval: 3, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "difficulty: 2.5") {
		t.Errorf("Expected difficulty: 2.5, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "last_reviewed: 2026-04-06") {
		t.Errorf("Expected last_reviewed: 2026-04-06, got:\n%s", contentStr)
	}
	// Verify question/answer preserved
	if !strings.Contains(contentStr, "What is a goroutine?") {
		t.Error("Question content was lost")
	}
	if !strings.Contains(contentStr, "lightweight thread") {
		t.Error("Answer content was lost")
	}
	// Verify tags preserved
	if !strings.Contains(contentStr, "tags: [go,test]") {
		t.Error("Tags were lost")
	}
}

func TestSaveCardToMarkdownSkipsDummyCards(t *testing.T) {
	store := &Store{}

	// Card with non-path ID should be silently skipped
	card := model.Card{
		ID:       "go-basics-1",
		DeckID:   "go-basics",
		Interval: 5,
	}

	if err := store.SaveCardToMarkdown(card); err != nil {
		t.Errorf("Expected nil error for dummy card, got: %v", err)
	}
}

func TestSaveCardToMarkdownSkipsNonExistentFiles(t *testing.T) {
	store := &Store{}

	card := model.Card{
		ID:       "/tmp/nonexistent/card.md",
		DeckID:   "/tmp/nonexistent",
		Interval: 5,
	}

	if err := store.SaveCardToMarkdown(card); err != nil {
		t.Errorf("Expected nil error for non-existent file, got: %v", err)
	}
}

func TestSaveCardReviewPersistsToFile(t *testing.T) {
	// Create a temp deck directory with a card file
	tempDir := t.TempDir()
	cardPath := filepath.Join(tempDir, "test-card.md")
	if err := os.WriteFile(cardPath, []byte(testCardContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create store from the temp directory
	store, err := NewStoreFromDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Find the card
	deck, found := store.GetDeck(tempDir)
	if !found {
		t.Fatal("Deck not found in store")
	}
	if len(deck.Cards) == 0 {
		t.Fatal("No cards in deck")
	}

	card := deck.Cards[0]
	if card.Interval != 0 {
		t.Fatalf("Expected initial interval 0, got %d", card.Interval)
	}

	// Review the card with rating 4 (Good)
	success := store.SaveCardReview(card, 4)
	if !success {
		t.Fatal("SaveCardReview returned false")
	}

	// Re-read the file and verify the interval was updated
	content, err := os.ReadFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	contentStr := string(content)

	// Rating 4 with interval 0 → interval becomes 1 (defaultInterval)
	if !strings.Contains(contentStr, "review_interval: 1") {
		t.Errorf("Expected review_interval: 1 after rating 4, got:\n%s", contentStr)
	}
}

func TestSaveDeckToMarkdown(t *testing.T) {
	// Create a temp deck with two card files
	tempDir := t.TempDir()

	card1Content := `---
tags: [go]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Question

Question one?

## Answer

Answer one.
`
	card2Content := `---
tags: [go]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# Question

Question two?

## Answer

Answer two.
`
	card1Path := filepath.Join(tempDir, "card1.md")
	card2Path := filepath.Join(tempDir, "card2.md")
	if err := os.WriteFile(card1Path, []byte(card1Content), 0644); err != nil {
		t.Fatalf("Failed to write card1: %v", err)
	}
	if err := os.WriteFile(card2Path, []byte(card2Content), 0644); err != nil {
		t.Fatalf("Failed to write card2: %v", err)
	}

	// Create store and update cards in memory
	store, err := NewStoreFromDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	deck, found := store.GetDeck(tempDir)
	if !found {
		t.Fatal("Deck not found")
	}

	// Update both cards' intervals in memory
	for i := range deck.Cards {
		updated := deck.Cards[i]
		updated.Interval = 5
		updated.Ease = 2.8
		updated.LastReviewed = time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
		store.UpdateCard(updated)
	}

	// Save via SaveDeckToMarkdown
	if err := store.SaveDeckToMarkdown(tempDir); err != nil {
		t.Fatalf("SaveDeckToMarkdown error: %v", err)
	}

	// Verify both files updated
	for _, path := range []string{card1Path, card2Path} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", path, err)
		}
		contentStr := string(content)
		if !strings.Contains(contentStr, "review_interval: 5") {
			t.Errorf("Expected review_interval: 5 in %s, got:\n%s", filepath.Base(path), contentStr)
		}
		if !strings.Contains(contentStr, "difficulty: 2.8") {
			t.Errorf("Expected difficulty: 2.8 in %s, got:\n%s", filepath.Base(path), contentStr)
		}
	}
}
