// File: internal/deck/deck_concurrent_test.go

package deck

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// TestConcurrentCardAddition tests adding cards to a deck concurrently
func TestConcurrentCardAddition(t *testing.T) {
	// Create a deck for testing
	testDeck := NewDeck("/test/deck", nil)

	// Number of goroutines to spawn
	numGoroutines := 4
	// Number of cards for each goroutine to add
	cardsPerGoroutine := 25

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch goroutines to add cards concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create and add cards
			for j := 0; j < cardsPerGoroutine; j++ {
				cardTitle := fmt.Sprintf("Card %d-%d", id, j)
				filePath := fmt.Sprintf("/test/deck/card_%d_%d.md", id, j)

				cardObj := &card.Card{
					Title:    cardTitle,
					FilePath: filePath,
				}

				testDeck.AddCard(cardObj)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify that all cards were added
	expectedCount := numGoroutines * cardsPerGoroutine
	if len(testDeck.Cards) != expectedCount {
		t.Errorf("Expected %d cards, got %d", expectedCount, len(testDeck.Cards))
	}
}

// TestConcurrentCardRemoval tests removing cards from a deck concurrently
func TestConcurrentCardRemoval(t *testing.T) {
	// Create a deck for testing
	testDeck := NewDeck("/test/deck", nil)

	// Add a number of cards first
	numCards := 100
	cards := make([]*card.Card, numCards)

	for i := 0; i < numCards; i++ {
		cards[i] = &card.Card{
			Title:    fmt.Sprintf("Card %d", i),
			FilePath: fmt.Sprintf("/test/deck/card_%d.md", i),
		}
		testDeck.AddCard(cards[i])
	}

	// Verify cards were added
	if len(testDeck.Cards) != numCards {
		t.Fatalf("Failed to add cards for test setup, got %d, expected %d", len(testDeck.Cards), numCards)
	}

	// Now remove half of them concurrently
	var wg sync.WaitGroup
	numGoroutines := 4
	cardsPerGoroutine := numCards / (2 * numGoroutines)
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			startIdx := id * cardsPerGoroutine
			endIdx := startIdx + cardsPerGoroutine

			for j := startIdx; j < endIdx; j++ {
				testDeck.RemoveCard(cards[j])
			}
		}(i)
	}

	wg.Wait()

	// Verify half the cards were removed
	expectedRemaining := numCards - (numGoroutines * cardsPerGoroutine)
	if len(testDeck.Cards) != expectedRemaining {
		t.Errorf("Expected %d cards remaining, got %d", expectedRemaining, len(testDeck.Cards))
	}
}

// TestConcurrentDeckHierarchy tests manipulating the deck hierarchy concurrently
func TestConcurrentDeckHierarchy(t *testing.T) {
	// Create a root deck
	rootDeck := NewDeck("/root", nil)

	// Create multiple subdecks concurrently
	numGoroutines := 5
	decksPerGoroutine := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < decksPerGoroutine; j++ {
				deckPath := fmt.Sprintf("/root/deck_%d_%d", id, j)
				subDeck := NewDeck(deckPath, nil)
				rootDeck.AddSubDeck(subDeck)

				// Add some cards to this subdeck
				for k := 0; k < 3; k++ {
					cardObj := &card.Card{
						Title:    fmt.Sprintf("Card %d_%d_%d", id, j, k),
						FilePath: fmt.Sprintf("%s/card_%d.md", deckPath, k),
					}
					subDeck.AddCard(cardObj)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify deck hierarchy
	expectedSubdecks := numGoroutines * decksPerGoroutine
	if len(rootDeck.SubDecks) != expectedSubdecks {
		t.Errorf("Expected %d subdecks, got %d", expectedSubdecks, len(rootDeck.SubDecks))
	}

	// Verify cards were added to the hierarchy
	allCards := rootDeck.GetAllCards()
	expectedCards := numGoroutines * decksPerGoroutine * 3
	if len(allCards) != expectedCards {
		t.Errorf("Expected %d total cards in hierarchy, got %d", expectedCards, len(allCards))
	}
}

// TestConcurrentStatisticsUpdate tests updating statistics concurrently
func TestConcurrentStatisticsUpdate(t *testing.T) {
	// Create a deck hierarchy for testing
	rootDeck := NewDeck("/root", nil)

	// Add several subdecks
	for i := 0; i < 5; i++ {
		subDeck := NewDeck(fmt.Sprintf("/root/subdeck_%d", i), nil)
		rootDeck.AddSubDeck(subDeck)

		// Add some cards to each subdeck
		for j := 0; j < 10; j++ {
			cardObj := &card.Card{
				Title:    fmt.Sprintf("Card %d_%d", i, j),
				FilePath: fmt.Sprintf("/root/subdeck_%d/card_%d.md", i, j),
			}
			subDeck.AddCard(cardObj)
		}
	}

	// Get all subdecks
	subDecks := make([]*Deck, 0, len(rootDeck.SubDecks))
	for _, subDeck := range rootDeck.SubDecks {
		subDecks = append(subDecks, subDeck)
	}

	// Concurrently update statistics by adding and removing cards
	var wg sync.WaitGroup
	wg.Add(len(subDecks))

	for i, subDeck := range subDecks {
		go func(idx int, d *Deck) {
			defer wg.Done()

			// Add some new cards
			for j := 0; j < 5; j++ {
				cardObj := &card.Card{
					Title:    fmt.Sprintf("NewCard %d_%d", idx, j),
					FilePath: fmt.Sprintf("%s/newcard_%d.md", d.Path, j),
				}
				d.AddCard(cardObj)
			}

			// Remove some existing cards
			existingCards := d.Cards
			for j := 0; j < 2 && j < len(existingCards); j++ {
				d.RemoveCard(existingCards[j])
			}
		}(i, subDeck)
	}

	wg.Wait()

	// Verify root deck statistics were updated
	rootTotalCards := rootDeck.CountAllCards()

	// Fix: Calculate the expected number of cards properly
	initialCards := 5 * 10 // Initial cards
	addedCards := 5 * 5    // Added cards
	removedCards := 5 * 2  // Removed cards
	expectedCards := initialCards + addedCards - removedCards

	if rootTotalCards != expectedCards {
		t.Errorf("Expected root deck to have %d cards after concurrent operations, got %d",
			expectedCards, rootTotalCards)
	}
}
