// File: internal/data/store.go

package data

import (
	"github.com/DavidMiserak/GoCard/internal/model"
	"time"
)

// Store manages all data for the application
type Store struct {
	Decks []model.Deck
}

// NewStore creates a new data store with dummy data
func NewStore() *Store {
	store := &Store{
		Decks: []model.Deck{},
	}

	// Add dummy data
	store.addDummyData()

	return store
}

// addDummyData adds sample decks and cards
func (s *Store) addDummyData() {
	// Go Programming Deck
	goCards := []model.Card{
		{
			ID:           "go-1",
			Question:     "What is the purpose of the \"defer\" keyword in Go?",
			Answer:       "The \"defer\" keyword in Go schedules a function call to be executed just before the function returns. This is often used for cleanup actions, ensuring they will be executed even if the function panics.",
			DeckID:       "go-programming",
			LastReviewed: time.Now(),
			NextReview:   time.Now().Add(24 * time.Hour),
			Ease:         2.5,
			Interval:     1,
			Rating:       4,
		},
		// Add more cards here as needed
	}

	goDeck := model.Deck{
		ID:          "go-programming",
		Name:        "Go Programming",
		Description: "Basic Go programming concepts",
		Cards:       goCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now(),
	}

	s.Decks = append(s.Decks, goDeck)

	// Computer Science Deck
	csCards := []model.Card{
		{
			ID:           "cs-1",
			Question:     "What is a compiler?",
			Answer:       "A compiler is a program that translates source code written in a high-level programming language into machine code or another lower-level form.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.3,
			Interval:     2,
			Rating:       3,
		},
		// Add more cards as needed
	}

	csDeck := model.Deck{
		ID:          "computer-science",
		Name:        "Computer Science",
		Description: "General computer science concepts",
		Cards:       csCards,
		CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		LastStudied: time.Now().Add(-24 * time.Hour),
	}

	s.Decks = append(s.Decks, csDeck)

	// Add more dummy decks matching the screenshots
	// Data Structures, Algorithms, Bubble Tea UI
}

// GetDecks returns all decks
func (s *Store) GetDecks() []model.Deck {
	return s.Decks
}

// GetDeck returns a deck by ID
func (s *Store) GetDeck(id string) (model.Deck, bool) {
	for _, deck := range s.Decks {
		if deck.ID == id {
			return deck, true
		}
	}
	return model.Deck{}, false
}

// GetDueCards returns cards due for review
func (s *Store) GetDueCards() []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		for _, card := range deck.Cards {
			if card.NextReview.Before(now) {
				dueCards = append(dueCards, card)
			}
		}
	}

	return dueCards
}
