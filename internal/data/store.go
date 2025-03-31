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
	store.Decks = GetDummyDecks()

	return store
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

// GetDueCardsForDeck returns cards due for review in a specific deck
func (s *Store) GetDueCardsForDeck(deckID string) []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		if deck.ID == deckID {
			for _, card := range deck.Cards {
				if card.NextReview.Before(now) {
					dueCards = append(dueCards, card)
				}
			}
			break
		}
	}

	return dueCards
}
