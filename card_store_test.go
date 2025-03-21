// Filename: card_store_test.go
// Version: 0.0.0
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

	card, err := store.CreateCard(title, question, answer, tags)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(card.FilePath); os.IsNotExist(err) {
		t.Errorf("Card file was not created on disk")
	}

	// Test loading the card
	loadedCard, err := store.LoadCard(card.FilePath)
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

	if err := store.SaveCard(loadedCard); err != nil {
		t.Fatalf("Failed to save updated card: %v", err)
	}

	// Reload the card and check if updates persisted
	updatedCard, err := store.LoadCard(card.FilePath)
	if err != nil {
		t.Fatalf("Failed to reload card: %v", err)
	}

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
	if _, err := os.Stat(card.FilePath); !os.IsNotExist(err) {
		t.Errorf("Card file was not deleted from disk")
	}

	// Test loading all cards in a directory
	// Create a few test cards
	for i := 0; i < 3; i++ {
		_, err := store.CreateCard(
			fmt.Sprintf("Test Card %d", i),
			fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i),
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
	subCard := &Card{
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
	for _, c := range newStore.Cards {
		c.LastReviewed = time.Now()
		c.ReviewInterval = 30 // due in 30 days
		if err := newStore.SaveCard(c); err != nil {
			t.Fatalf("Failed to update card review date: %v", err)
		}
		break // only update one card
	}

	// Reload and check due cards
	newerStore, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create newer card store: %v", err)
	}

	dueCards = newerStore.GetDueCards()
	if len(dueCards) != expectedCount-1 {
		t.Errorf("Expected %d due cards after update, got %d", expectedCount-1, len(dueCards))
	}
}
