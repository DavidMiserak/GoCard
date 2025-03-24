// File: internal/storage/store_test.go

// Package storage contains tests for the file-based storage system.
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
}

// Test deck operations in a separate test function to isolate them
func TestDecks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-deck-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Initialize a new store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store for deck tests: %v", err)
	}

	// Verify that the root deck exists
	if store.RootDeck == nil {
		t.Fatalf("Root deck not initialized")
	}

	// Test creating a deck
	algorithmsDeck, err := store.CreateDeck("Algorithms", nil)
	if err != nil {
		t.Fatalf("Failed to create algorithms deck: %v", err)
	}

	// Verify the deck directory was created
	algorithmsDirPath := filepath.Join(tempDir, "algorithms")
	if _, err := os.Stat(algorithmsDirPath); os.IsNotExist(err) {
		t.Errorf("Deck directory was not created on disk")
	}

	// Verify the deck was added to the store
	if store.Decks[algorithmsDirPath] != algorithmsDeck {
		t.Errorf("Deck not correctly added to store.Decks map")
	}

	// Verify the deck was added as a subdeck to the root deck
	if store.RootDeck.SubDecks["algorithms"] != algorithmsDeck {
		t.Errorf("Deck not correctly added as subdeck to root deck")
	}

	// Test creating a subdeck
	sortingDeck, err := store.CreateDeck("Sorting", algorithmsDeck)
	if err != nil {
		t.Fatalf("Failed to create sorting deck: %v", err)
	}

	// Verify the subdeck directory was created
	sortingDirPath := filepath.Join(algorithmsDirPath, "sorting")
	if _, err := os.Stat(sortingDirPath); os.IsNotExist(err) {
		t.Errorf("Subdeck directory was not created on disk")
	}

	// Verify the subdeck was added to the parent deck
	if algorithmsDeck.SubDecks["sorting"] != sortingDeck {
		t.Errorf("Subdeck not correctly added to parent deck")
	}

	// Test creating a card in a specific deck
	cardTitle := "Quick Sort"
	cardQuestion := "How does quicksort work?"
	cardAnswer := "Quicksort is a divide and conquer algorithm that picks a pivot..."
	cardTags := []string{"sorting", "algorithms"}

	quickSortCard, err := store.CreateCardInDeck(cardTitle, cardQuestion, cardAnswer, cardTags, sortingDeck)
	if err != nil {
		t.Fatalf("Failed to create card in deck: %v", err)
	}

	// Verify the card file was created in the correct directory
	if filepath.Dir(quickSortCard.FilePath) != sortingDirPath {
		t.Errorf("Card file not created in correct directory. Expected: %s, Got: %s",
			sortingDirPath, filepath.Dir(quickSortCard.FilePath))
	}

	// Verify the card was added to the deck (more direct check)
	if len(sortingDeck.Cards) == 0 {
		t.Errorf("Card not added to deck's Cards slice")
	} else if sortingDeck.Cards[0].Title != cardTitle {
		t.Errorf("Wrong card in deck. Expected: %s, Got: %s", cardTitle, sortingDeck.Cards[0].Title)
	}

	// Test moving a card between decks
	// First create another deck to move the card to
	searchingDeck, err := store.CreateDeck("Searching", algorithmsDeck)
	if err != nil {
		t.Fatalf("Failed to create searching deck: %v", err)
	}

	// Move the card to the new deck
	err = store.MoveCard(quickSortCard, searchingDeck)
	if err != nil {
		t.Fatalf("Failed to move card between decks: %v", err)
	}

	// Calculate the expected new path
	searchingDirPath := filepath.Join(algorithmsDirPath, "searching")
	expectedNewPath := filepath.Join(searchingDirPath, filepath.Base(quickSortCard.FilePath))

	// Get the moved card using the new path
	movedCard, exists := store.GetCardByPath(expectedNewPath)
	if !exists {
		t.Errorf("Card not found at new location: %s", expectedNewPath)
	} else {
		// The card was successfully moved
		// For backward compatibility, update the test's reference to point to the new card
		quickSortCard = movedCard
	}

	// Verify the card file was moved to the new directory
	if filepath.Dir(quickSortCard.FilePath) != searchingDirPath {
		t.Errorf("Card file not moved to correct directory. Expected: %s, Got: %s",
			searchingDirPath, filepath.Dir(quickSortCard.FilePath))
	}

	// Verify the card file was moved to the new directory
	searchingDirPath = filepath.Join(algorithmsDirPath, "searching")
	if filepath.Dir(quickSortCard.FilePath) != searchingDirPath {
		t.Errorf("Card file not moved to correct directory. Expected: %s, Got: %s",
			searchingDirPath, filepath.Dir(quickSortCard.FilePath))
	}

	// Verify the card was removed from the old deck
	if len(sortingDeck.Cards) != 0 {
		t.Errorf("Card not removed from original deck after move, still has %d cards", len(sortingDeck.Cards))
	}

	// Verify the card was added to the new deck
	if len(searchingDeck.Cards) == 0 {
		t.Errorf("Card not added to new deck after move")
	} else if searchingDeck.Cards[0].Title != cardTitle {
		t.Errorf("Wrong card in new deck. Expected: %s, Got: %s", cardTitle, searchingDeck.Cards[0].Title)
	}

	// Test renaming a deck
	err = store.RenameDeck(sortingDeck, "SortAlgorithms")
	if err != nil {
		t.Fatalf("Failed to rename deck: %v", err)
	}

	// Verify the directory was renamed
	newSortingDirPath := filepath.Join(algorithmsDirPath, "sortalgorithms")
	if _, err := os.Stat(newSortingDirPath); os.IsNotExist(err) {
		t.Errorf("Deck directory was not renamed on disk")
	}

	// Verify the old directory no longer exists
	if _, err := os.Stat(sortingDirPath); !os.IsNotExist(err) {
		t.Errorf("Old deck directory still exists after rename")
	}

	// Verify the deck's name was updated
	if sortingDeck.Name != "sortalgorithms" {
		t.Errorf("Deck name not updated. Expected: sortalgorithms, Got: %s", sortingDeck.Name)
	}

	// Test deleting a deck
	err = store.DeleteDeck(searchingDeck)
	if err != nil {
		t.Fatalf("Failed to delete deck: %v", err)
	}

	// Verify the directory was deleted
	if _, err := os.Stat(searchingDirPath); !os.IsNotExist(err) {
		t.Errorf("Deck directory not deleted from disk")
	}

	// Verify the deck was removed from the parent's subdecks
	if _, exists := algorithmsDeck.SubDecks["searching"]; exists {
		t.Errorf("Deck not removed from parent's subdecks after deletion")
	}

	// Verify the card was removed from the store
	if _, exists := store.Cards[quickSortCard.FilePath]; exists {
		t.Errorf("Card not removed from store after deck deletion")
	}

	// Test getting deck stats
	// First add some cards to test with
	for i := 0; i < 3; i++ {
		_, err := store.CreateCardInDeck(
			fmt.Sprintf("Algorithm %d", i),
			fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i),
			[]string{"test"},
			algorithmsDeck,
		)
		if err != nil {
			t.Fatalf("Failed to create test card %d: %v", i, err)
		}
	}

	// Get deck stats
	stats := store.GetDeckStats(algorithmsDeck)

	// Verify stats - we should have exactly 3 cards now
	if stats["total_cards"].(int) != 3 {
		t.Errorf("Expected 3 total cards in deck stats, got %d", stats["total_cards"])
	}

	if stats["sub_decks"].(int) != 1 {
		t.Errorf("Expected 1 subdeck in deck stats, got %d", stats["sub_decks"])
	}
}

// Test deck hierarchy in a separate test function
func TestDeckHierarchy(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-hierarchy-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}

	// Create a hierarchy of decks
	mathDeck, _ := store.CreateDeck("Math", nil)
	calculusDeck, _ := store.CreateDeck("Calculus", mathDeck)
	limitsDeck, _ := store.CreateDeck("Limits", calculusDeck)

	// Verify the hierarchy
	allDecks := store.RootDeck.AllDecks()
	if len(allDecks) != 4 { // Root + 3 created decks
		t.Errorf("Expected 4 decks in total, got %d", len(allDecks))
	}

	// Test deck path methods
	if limitsDeck.PathFromRoot() != "math/calculus/limits" {
		t.Errorf("Incorrect path from root. Expected: math/calculus/limits, Got: %s",
			limitsDeck.PathFromRoot())
	}

	// Test GetDeckByPath
	foundDeck, err := store.GetDeckByRelativePath("math/calculus")
	if err != nil {
		t.Fatalf("Failed to get deck by path: %v", err)
	}
	if foundDeck != calculusDeck {
		t.Errorf("GetDeckByRelativePath returned incorrect deck")
	}
}

// Test card loading in a separate test function
func TestCardLoading(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-loading-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a new store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store for card loading tests: %v", err)
	}

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

	// Add a card in the subdirectory - but create this directly through the file system
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

	// Write the card file directly using our formatter
	content, err := store.FormatCardAsMarkdown(subCard)
	if err != nil {
		t.Fatalf("Failed to format card as markdown: %v", err)
	}

	// Write the card file directly
	if err := os.WriteFile(subCard.FilePath, content, 0644); err != nil {
		t.Fatalf("Failed to write card file: %v", err)
	}

	// Reload all cards with a new store to test discovery
	newStore, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create new card store: %v", err)
	}

	// Check if all cards were loaded - should be exactly 4 cards
	expectedCount := 4 // 3 in root dir + 1 in subdir
	if len(newStore.Cards) != expectedCount {
		t.Errorf("Expected %d cards, got %d", expectedCount, len(newStore.Cards))
	}

	// Check if subdirectory was discovered as a deck
	subDirDeck, err := newStore.GetDeckByRelativePath("subcategory")
	if err != nil {
		t.Fatalf("Failed to get subdirectory deck: %v", err)
	}

	if subDirDeck == nil {
		t.Fatalf("Subdirectory deck not discovered")
	}

	// Check if the card in the subdirectory was assigned to the correct deck
	if len(subDirDeck.Cards) != 1 {
		t.Errorf("Expected 1 card in subdirectory deck, got %d", len(subDirDeck.Cards))
	}
}

func TestDueCards(t *testing.T) {
	// Create a temporary directory in a completely different location
	tempDir, err := os.MkdirTemp("", "gocard-due-test-isolated")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tempDir)

	// Initialize a new store
	_, err = NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subcategory")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Force reload to recognize the subdirectory as a deck
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to reload card store: %v", err)
	}

	// Get the subdirectory deck
	subDirDeck, err := store.GetDeckByRelativePath("subcategory")
	if err != nil {
		t.Fatalf("Failed to get subdirectory deck: %v", err)
	}

	// Create a due card in the root deck
	dueCard, err := store.CreateCard(
		"Due Card",
		"Due Question",
		"Due Answer",
		[]string{"due"},
	)
	if err != nil {
		t.Fatalf("Failed to create due card: %v", err)
	}

	// Create a due card in the subdirectory
	subDueCard, err := store.CreateCardInDeck(
		"Sub Due Card",
		"Sub Due Question",
		"Sub Due Answer",
		[]string{"due"},
		subDirDeck,
	)
	if err != nil {
		t.Fatalf("Failed to create sub due card: %v", err)
	}

	// Verify the card was created with correct properties
	if subDueCard == nil {
		t.Fatalf("Failed to create sub due card")
	}
	if filepath.Base(subDueCard.FilePath) != "sub-due-card.md" {
		t.Errorf("Incorrect filename for subdirectory due card: %s", filepath.Base(subDueCard.FilePath))
	}

	// Add this after creating the cards but before checking due cards
	fmt.Printf("Debugging all cards in store:\n")
	for path, c := range store.Cards {
		fmt.Printf("Card at %s: %s\n", path, c.Title)
	}

	// Both cards should be due initially
	dueCardsRoot := store.GetDueCardsInDeck(store.RootDeck)

	// Add this when checking due cards
	fmt.Printf("Due cards in root deck and subdecks:\n")
	for i, c := range dueCardsRoot {
		fmt.Printf("Due card %d: %s at %s\n", i, c.Title, c.FilePath)
	}

	if len(dueCardsRoot) != 2 { // 2 total cards (1 in root, 1 in sub)
		t.Errorf("Expected 2 due cards in root deck and all subdecks, got %d", len(dueCardsRoot))
	}

	dueCardsSub := store.GetDueCardsInDeck(subDirDeck)
	if len(dueCardsSub) != 1 { // 1 card in subdirectory
		t.Errorf("Expected 1 due card in subdirectory deck, got %d", len(dueCardsSub))
	}

	// Update one card to be not due
	dueCard.LastReviewed = time.Now()
	dueCard.ReviewInterval = 30 // due in 30 days
	if err := store.SaveCard(dueCard); err != nil {
		t.Fatalf("Failed to update card: %v", err)
	}

	// Reload to ensure changes are persisted
	newerStore, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create newer card store: %v", err)
	}

	// Debug print all cards
	for path, c := range newerStore.Cards {
		fmt.Printf("Card in new store: %s, LastReviewed=%v, ReviewInterval=%d, IsDue=%v\n",
			path, c.LastReviewed, c.ReviewInterval,
			algorithm.SM2.IsDue(c))
	}

	// Verify due cards - we should only see the subdirectory card as due
	dueCardsRoot = newerStore.GetDueCardsInDeck(newerStore.RootDeck)
	if len(dueCardsRoot) != 1 { // Only the subdir card should be due
		t.Errorf("Expected 1 due card after update, got %d", len(dueCardsRoot))
	}
}
