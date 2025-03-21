// File: internal/deck/deck_test.go
package deck

import (
	"testing"

	"github.com/DavidMiserak/GoCard/internal/card"
)

func TestDeck(t *testing.T) {
	// Create a root deck
	rootDeck := NewDeck("/root", nil)

	// Test basic deck properties
	if rootDeck.Name != "root" {
		t.Errorf("Expected deck name 'root', got %s", rootDeck.Name)
	}

	if rootDeck.Path != "/root" {
		t.Errorf("Expected deck path '/root', got %s", rootDeck.Path)
	}

	if len(rootDeck.Cards) != 0 {
		t.Errorf("Expected 0 cards, got %d", len(rootDeck.Cards))
	}

	if len(rootDeck.SubDecks) != 0 {
		t.Errorf("Expected 0 subdecks, got %d", len(rootDeck.SubDecks))
	}

	// Test creating subdeck
	subDeck1 := NewDeck("/root/subdeck1", rootDeck)
	rootDeck.AddSubDeck(subDeck1)

	if len(rootDeck.SubDecks) != 1 {
		t.Errorf("Expected 1 subdeck, got %d", len(rootDeck.SubDecks))
	}

	if rootDeck.SubDecks["subdeck1"] != subDeck1 {
		t.Errorf("Subdeck not correctly added to parent")
	}

	if subDeck1.ParentDeck != rootDeck {
		t.Errorf("Parent deck not correctly set")
	}

	// Test nested subdeck
	subDeck2 := NewDeck("/root/subdeck1/subdeck2", subDeck1)
	subDeck1.AddSubDeck(subDeck2)

	if len(subDeck1.SubDecks) != 1 {
		t.Errorf("Expected 1 nested subdeck, got %d", len(subDeck1.SubDecks))
	}

	// Test adding cards
	card1 := &card.Card{
		Title:    "Card 1",
		FilePath: "/root/card1.md",
	}

	card2 := &card.Card{
		Title:    "Card 2",
		FilePath: "/root/subdeck1/card2.md",
	}

	card3 := &card.Card{
		Title:    "Card 3",
		FilePath: "/root/subdeck1/subdeck2/card3.md",
	}

	rootDeck.AddCard(card1)
	subDeck1.AddCard(card2)
	subDeck2.AddCard(card3)

	if len(rootDeck.Cards) != 1 {
		t.Errorf("Expected 1 card in root deck, got %d", len(rootDeck.Cards))
	}

	if len(subDeck1.Cards) != 1 {
		t.Errorf("Expected 1 card in subdeck1, got %d", len(subDeck1.Cards))
	}

	if len(subDeck2.Cards) != 1 {
		t.Errorf("Expected 1 card in subdeck2, got %d", len(subDeck2.Cards))
	}

	// Test GetAllCards
	allCards := rootDeck.GetAllCards()
	if len(allCards) != 3 {
		t.Errorf("Expected 3 cards in total, got %d", len(allCards))
	}

	// Test AllDecks
	allDecks := rootDeck.AllDecks()
	if len(allDecks) != 3 {
		t.Errorf("Expected 3 decks in total, got %d", len(allDecks))
	}

	// Test GetDeckByPath
	foundDeck := rootDeck.GetDeckByPath("subdeck1/subdeck2")
	if foundDeck != subDeck2 {
		t.Errorf("GetDeckByPath failed to find correct deck")
	}

	// Test PathFromRoot
	if rootDeck.PathFromRoot() != "" {
		t.Errorf("Expected root deck path to be empty, got %s", rootDeck.PathFromRoot())
	}

	if subDeck1.PathFromRoot() != "subdeck1" {
		t.Errorf("Expected subdeck1 path to be 'subdeck1', got %s", subDeck1.PathFromRoot())
	}

	if subDeck2.PathFromRoot() != "subdeck1/subdeck2" {
		t.Errorf("Expected subdeck2 path to be 'subdeck1/subdeck2', got %s", subDeck2.PathFromRoot())
	}

	// Test CountAllCards
	if rootDeck.CountAllCards() != 3 {
		t.Errorf("Expected 3 cards in total, got %d", rootDeck.CountAllCards())
	}

	if subDeck1.CountAllCards() != 2 {
		t.Errorf("Expected 2 cards in subdeck1 (including nested), got %d", subDeck1.CountAllCards())
	}

	// Test RemoveCard
	if !rootDeck.RemoveCard(card1) {
		t.Errorf("Failed to remove card from deck")
	}

	if len(rootDeck.Cards) != 0 {
		t.Errorf("Expected 0 cards after removal, got %d", len(rootDeck.Cards))
	}

	// Test GetCardsByTag
	card2.Tags = []string{"test", "important"}
	card3.Tags = []string{"test"}

	testTagCards := rootDeck.GetCardsByTag("test")
	if len(testTagCards) != 2 {
		t.Errorf("Expected 2 cards with tag 'test', got %d", len(testTagCards))
	}

	importantTagCards := rootDeck.GetCardsByTag("important")
	if len(importantTagCards) != 1 {
		t.Errorf("Expected 1 card with tag 'important', got %d", len(importantTagCards))
	}
}
