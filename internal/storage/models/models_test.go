// File: internal/storage/models/models_test.go

package models

import (
	"fmt"
	"testing"
	"time"
)

// TestCardCreation tests the creation of cards with validation
func TestCardCreation(t *testing.T) {
	// Test valid card creation
	card, err := NewCard("Test Card", "Test Question", "Test Answer", []string{"tag1", "tag2"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if card == nil {
		t.Fatalf("Expected card to be created, got nil")
	}

	// Verify fields
	if card.GetTitle() != "Test Card" {
		t.Errorf("Expected title 'Test Card', got %s", card.GetTitle())
	}
	if card.GetQuestion() != "Test Question" {
		t.Errorf("Expected question 'Test Question', got %s", card.GetQuestion())
	}
	if card.GetAnswer() != "Test Answer" {
		t.Errorf("Expected answer 'Test Answer', got %s", card.GetAnswer())
	}

	tags := card.GetTags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	} else {
		if tags[0] != "tag1" || tags[1] != "tag2" {
			t.Errorf("Tags don't match expected values: %v", tags)
		}
	}

	// Test card creation with invalid data
	_, err = NewCard("", "Question", "Answer", []string{})
	if err == nil {
		t.Errorf("Expected error for empty title, got none")
	}

	_, err = NewCard("Title", "", "Answer", []string{})
	if err == nil {
		t.Errorf("Expected error for empty question, got none")
	}

	// Test that empty answer is allowed
	card, err = NewCard("Title", "Question", "", []string{})
	if err != nil {
		t.Errorf("Expected no error for empty answer, got %v", err)
	}
	if card == nil {
		t.Errorf("Expected card with empty answer to be created")
	}
}

// TestCardThreadSafety tests the thread-safety of card operations
func TestCardThreadSafety(t *testing.T) {
	card, _ := NewCard("Test Card", "Test Question", "Test Answer", []string{"tag1"})

	// Test setting and getting title
	err := card.SetTitle("New Title")
	if err != nil {
		t.Errorf("Failed to set title: %v", err)
	}
	if card.GetTitle() != "New Title" {
		t.Errorf("Expected title 'New Title', got %s", card.GetTitle())
	}

	// Test setting and getting question
	err = card.SetQuestion("New Question")
	if err != nil {
		t.Errorf("Failed to set question: %v", err)
	}
	if card.GetQuestion() != "New Question" {
		t.Errorf("Expected question 'New Question', got %s", card.GetQuestion())
	}

	// Test setting and getting answer
	card.SetAnswer("New Answer")
	if card.GetAnswer() != "New Answer" {
		t.Errorf("Expected answer 'New Answer', got %s", card.GetAnswer())
	}

	// Test setting and getting tags
	newTags := []string{"new", "tags"}
	card.SetTags(newTags)
	tags := card.GetTags()
	if len(tags) != 2 || tags[0] != "new" || tags[1] != "tags" {
		t.Errorf("Expected tags %v, got %v", newTags, tags)
	}

	// Test that modifying returned tags doesn't affect original
	tags[0] = "modified"
	originalTags := card.GetTags()
	if originalTags[0] != "new" {
		t.Errorf("Tags were modified externally: %v", originalTags)
	}

	// Test setting and getting difficulty
	err = card.SetDifficulty(3)
	if err != nil {
		t.Errorf("Failed to set difficulty: %v", err)
	}
	if card.GetDifficulty() != 3 {
		t.Errorf("Expected difficulty 3, got %d", card.GetDifficulty())
	}

	// Test invalid difficulty
	err = card.SetDifficulty(6)
	if err == nil {
		t.Errorf("Expected error for invalid difficulty, got none")
	}

	// Test review interval
	card.SetReviewInterval(5)
	if card.GetReviewInterval() != 5 {
		t.Errorf("Expected review interval 5, got %d", card.GetReviewInterval())
	}

	// Test last reviewed time
	now := time.Now()
	card.SetLastReviewedTime(now)
	if !card.GetLastReviewedTime().Equal(now) {
		t.Errorf("Expected last reviewed time %v, got %v", now, card.GetLastReviewedTime())
	}
}

// TestDeckCreation tests the creation of decks with validation
func TestDeckCreation(t *testing.T) {
	// Test valid deck creation
	deck, err := NewDeck("/test/deck", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if deck == nil {
		t.Fatalf("Expected deck to be created, got nil")
	}

	// Verify fields
	if deck.GetName() != "deck" {
		t.Errorf("Expected name 'deck', got %s", deck.GetName())
	}
	if deck.GetPath() != "/test/deck" {
		t.Errorf("Expected path '/test/deck', got %s", deck.GetPath())
	}

	// Test deck creation with invalid data
	_, err = NewDeck("", nil)
	if err == nil {
		t.Errorf("Expected error for empty path, got none")
	}

	// Test with parent deck
	parentDeck, _ := NewDeck("/test", nil)
	childDeck, err := NewDeck("/test/child", parentDeck)
	if err != nil {
		t.Errorf("Expected no error for child deck, got %v", err)
	}
	if childDeck.GetParentDeck() != parentDeck {
		t.Errorf("Expected parent deck to be set correctly")
	}
}

// TestDeckCardOperations tests adding and removing cards from a deck
func TestDeckCardOperations(t *testing.T) {
	deck, _ := NewDeck("/test/deck", nil)

	// Create test cards
	card1, _ := NewCard("Card 1", "Question 1", "Answer 1", []string{"tag1"})
	card1.SetFilePath("/test/deck/card1.md")

	card2, _ := NewCard("Card 2", "Question 2", "Answer 2", []string{"tag2"})
	card2.SetFilePath("/test/deck/card2.md")

	// Add cards to deck
	deck.AddCard(card1)
	deck.AddCard(card2)

	// Verify cards were added
	cards := deck.GetCards()
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}

	// Test removing a card by title
	result := deck.RemoveCard(card1)
	if !result {
		t.Errorf("Failed to remove card")
	}

	// Verify card was removed
	cards = deck.GetCards()
	if len(cards) != 1 {
		t.Errorf("Expected 1 card after removal, got %d", len(cards))
	}

	// Test removing a card by filepath
	card3, _ := NewCard("Card 3", "Question 3", "Answer 3", []string{"tag3"})
	card3.SetFilePath("/test/deck/card3.md")
	deck.AddCard(card3)

	result = deck.RemoveCard(card3)
	if !result {
		t.Errorf("Failed to remove card by filepath")
	}

	// Verify card was removed
	cards = deck.GetCards()
	if len(cards) != 1 {
		t.Errorf("Expected 1 card after second removal, got %d", len(cards))
	}
}

// TestDeckHierarchy tests deck parent-child relationships
func TestDeckHierarchy(t *testing.T) {
	// Create deck hierarchy
	rootDeck, _ := NewDeck("/root", nil)
	subDeck1, _ := NewDeck("/root/sub1", rootDeck)
	subDeck2, _ := NewDeck("/root/sub1/sub2", subDeck1)

	// Add subdeck relationships
	rootDeck.AddSubDeck(subDeck1)
	subDeck1.AddSubDeck(subDeck2)

	// Verify parent-child relationships
	subDecks := rootDeck.GetSubDecks()
	if len(subDecks) != 1 {
		t.Errorf("Expected 1 subdeck, got %d", len(subDecks))
	}

	// Test PathFromRoot
	if subDeck1.PathFromRoot() != "sub1" {
		t.Errorf("Expected path 'sub1', got %s", subDeck1.PathFromRoot())
	}
	if subDeck2.PathFromRoot() != "sub1/sub2" {
		t.Errorf("Expected path 'sub1/sub2', got %s", subDeck2.PathFromRoot())
	}

	// Test GetDeckByPath
	foundDeck := rootDeck.GetDeckByPath("sub1/sub2")
	if foundDeck != subDeck2 {
		t.Errorf("Failed to find subdeck by path")
	}

	// Test AllDecks
	allDecks := rootDeck.AllDecks()
	if len(allDecks) != 3 {
		t.Errorf("Expected 3 decks in total, got %d", len(allDecks))
	}

	// Test removing a subdeck
	rootDeck.RemoveSubDeck("sub1")
	subDecks = rootDeck.GetSubDecks()
	if len(subDecks) != 0 {
		t.Errorf("Expected 0 subdecks after removal, got %d", len(subDecks))
	}
}

// TestTaggedCards tests finding cards by tag
func TestTaggedCards(t *testing.T) {
	deck, _ := NewDeck("/test/deck", nil)

	// Create cards with different tags
	card1, _ := NewCard("Card 1", "Q1", "A1", []string{"important", "review"})
	card1.SetFilePath("/test/deck/card1.md")

	card2, _ := NewCard("Card 2", "Q2", "A2", []string{"review"})
	card2.SetFilePath("/test/deck/card2.md")

	card3, _ := NewCard("Card 3", "Q3", "A3", []string{"common"})
	card3.SetFilePath("/test/deck/card3.md")

	// Add cards to deck
	deck.AddCard(card1)
	deck.AddCard(card2)
	deck.AddCard(card3)

	// Find cards by tag
	importantCards := deck.GetCardsByTag("important")
	if len(importantCards) != 1 {
		t.Errorf("Expected 1 important card, got %d", len(importantCards))
	}

	reviewCards := deck.GetCardsByTag("review")
	if len(reviewCards) != 2 {
		t.Errorf("Expected 2 review cards, got %d", len(reviewCards))
	}

	// Test with non-existent tag
	noCards := deck.GetCardsByTag("nonexistent")
	if len(noCards) != 0 {
		t.Errorf("Expected 0 cards for nonexistent tag, got %d", len(noCards))
	}
}

// TestStatisticsUpdates tests that statistics are updated correctly
func TestStatisticsUpdates(t *testing.T) {
	rootDeck, _ := NewDeck("/root", nil)
	subDeck, _ := NewDeck("/root/sub", rootDeck)

	rootDeck.AddSubDeck(subDeck)

	// Add some cards to the subdeck
	for i := 0; i < 3; i++ {
		card, _ := NewCard(
			fmt.Sprintf("Card %d", i),
			fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i),
			[]string{"test"},
		)
		card.SetFilePath(fmt.Sprintf("/root/sub/card%d.md", i))
		subDeck.AddCard(card)
	}

	// Force statistics update
	rootDeck.UpdateStatistics()

	// Wait a short time for async updates to complete
	time.Sleep(50 * time.Millisecond)

	// Verify statistics
	rootStats := rootDeck.GetStatistics()
	if rootStats["total_cards"] != 3 {
		t.Errorf("Expected 3 total cards in root stats, got %d", rootStats["total_cards"])
	}

	subStats := subDeck.GetStatistics()
	if subStats["total_cards"] != 3 {
		t.Errorf("Expected 3 total cards in subdeck stats, got %d", subStats["total_cards"])
	}
}
