// File: internal/storage/store_test.go
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/card"
)

func TestCardStore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Initialize a card store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}

	// Test creating a card
	title := "Test Card"
	question := "What is the test question?"
	answer := "This is the test answer."
	tags := []string{"test", "example"}

	testCard, err := store.CreateCard(title, question, answer, tags)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(testCard.FilePath); os.IsNotExist(err) {
		t.Errorf("Card file was not created on disk")
	}

	// Test loading the card
	loadedCard, err := store.LoadCard(testCard.FilePath)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	// Verify card contents
	if loadedCard.Title != title {
		t.Errorf("Expected title %q, got %q", title, loadedCard.Title)
	}
	if loadedCard.Question != question {
		t.Errorf("Expected question %q, got %q", question, loadedCard.Question)
	}
	if loadedCard.Answer != answer {
		t.Errorf("Expected answer %q, got %q", answer, loadedCard.Answer)
	}

	// Test updating a card
	loadedCard.Difficulty = 3
	loadedCard.LastReviewed = time.Now()
	loadedCard.ReviewInterval = 2

	// Debug print
	fmt.Printf("Before saving: ReviewInterval=%d\n", loadedCard.ReviewInterval)

	if err := store.SaveCard(loadedCard); err != nil {
		t.Fatalf("Failed to save updated card: %v", err)
	}

	// Debug: Check what we're writing to the file
	content, err := os.ReadFile(testCard.FilePath)
	if err == nil {
		fmt.Printf("File content after save:\n%s\n", string(content))
	}

	// Reload the card and check if updates persisted
	updatedCard, err := store.LoadCard(testCard.FilePath)
	if err != nil {
		t.Fatalf("Failed to reload card: %v", err)
	}

	// Debug print
	fmt.Printf("After loading: ReviewInterval=%d\n", updatedCard.ReviewInterval)

	if updatedCard.Difficulty != 3 {
		t.Errorf("Expected difficulty 3, got %d", updatedCard.Difficulty)
	}
	if updatedCard.ReviewInterval != 2 {
		t.Errorf("Expected review interval 2, got %d", updatedCard.ReviewInterval)
	}

	// Test deleting a card
	if err := store.DeleteCard(updatedCard); err != nil {
		t.Fatalf("Failed to delete card: %v", err)
	}

	// Check if file was removed
	if _, err := os.Stat(testCard.FilePath); !os.IsNotExist(err) {
		t.Errorf("Card file was not deleted from disk")
	}

	// Test loading all cards in a directory
	// Create a few test cards
	for i := 0; i < 3; i++ {
		_, err := store.CreateCard(
			"Test Card"+strconv.Itoa(i),
			"Question "+strconv.Itoa(i),
			"Answer "+strconv.Itoa(i),
			[]string{"test"},
		)
		if err != nil {
			t.Fatalf("Failed to create test card %d: %v", i, err)
		}
	}

	// Create a subdirectory with cards
	subDir := filepath.Join(tempDir, "subcategory")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Add a card in the subdirectory
	subCard := &card.Card{
		Title:          "Subdirectory Card",
		Question:       "Question in subdirectory",
		Answer:         "Answer in subdirectory",
		Tags:           []string{"sub", "test"},
		Created:        time.Now(),
		LastReviewed:   time.Time{},
		ReviewInterval: 0,
		Difficulty:     0,
		FilePath:       filepath.Join(subDir, "subcard.md"),
	}
	if err := store.SaveCard(subCard); err != nil {
		t.Fatalf("Failed to save card in subdirectory: %v", err)
	}

	// Reload all cards
	newStore, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new card store: %v", err)
	}

	// Check if all cards were loaded
	expectedCount := 4 // 3 in root dir + 1 in subdir
	if len(newStore.Cards) != expectedCount {
		t.Errorf("Expected %d cards, got %d", expectedCount, len(newStore.Cards))
	}

	// Check if subdirectory card was loaded
	found := false
	for _, c := range newStore.Cards {
		if c.Title == "Subdirectory Card" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Failed to find card in subdirectory")
	}

	// Test due cards
	// Make sure all cards are due initially (since LastReviewed is zero time)
	dueCards := newStore.GetDueCards()
	if len(dueCards) != expectedCount {
		t.Errorf("Expected %d due cards, got %d", expectedCount, len(dueCards))
	}

	// Update one card to be not due
	var updatedCardPath string
	for path, c := range newStore.Cards {
		// Set a review date in the past
		c.LastReviewed = time.Now()
		// Use a large interval to ensure it's not due
		c.ReviewInterval = 30 // due in 30 days

		// Debug print
		fmt.Printf("Setting card to not due: LastReviewed=%v, ReviewInterval=%d\n",
			c.LastReviewed, c.ReviewInterval)

		if err := newStore.SaveCard(c); err != nil {
			t.Fatalf("Failed to update card review date: %v", err)
		}
		updatedCardPath = path
		break // only update one card
	}

	// Reload and check due cards
	newerStore, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create newer card store: %v", err)
	}

	// Debug print all cards from the new store
	for path, c := range newerStore.Cards {
		fmt.Printf("Card in new store: %s, LastReviewed=%v, ReviewInterval=%d, IsDue=%v\n",
			path, c.LastReviewed, c.ReviewInterval,
			algorithm.SM2.IsDue(c))
	}

	// Verify the specific card we updated is not due
	updatedCard = newerStore.Cards[updatedCardPath]
	if updatedCard == nil {
		t.Fatalf("Failed to find updated card at path: %s", updatedCardPath)
	}

	// Debug
	if algorithm.SM2.IsDue(updatedCard) {
		t.Errorf("Card should not be due: LastReviewed=%v, ReviewInterval=%d, Current time=%v, Due date=%v",
			updatedCard.LastReviewed, updatedCard.ReviewInterval, time.Now(),
			updatedCard.LastReviewed.AddDate(0, 0, updatedCard.ReviewInterval))
	}

	dueCards = newerStore.GetDueCards()
	if len(dueCards) != expectedCount-1 {
		t.Errorf("Expected %d due cards after update, got %d", expectedCount-1, len(dueCards))

		// Additional debug - list which cards are due
		for _, c := range dueCards {
			fmt.Printf("Due card: %s, LastReviewed=%v, ReviewInterval=%d\n",
				c.Title, c.LastReviewed, c.ReviewInterval)
		}
	}
}
