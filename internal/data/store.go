// File: internal/data/store.go

package data

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
	"github.com/DavidMiserak/GoCard/internal/srs"
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

// UpdateCard updates a card in the store and returns whether it was found
func (s *Store) UpdateCard(updatedCard model.Card) bool {
	// Find and update the card in its deck
	for i, deck := range s.Decks {
		if deck.ID == updatedCard.DeckID {
			for j, card := range deck.Cards {
				if card.ID == updatedCard.ID {
					// Update the card
					s.Decks[i].Cards[j] = updatedCard
					return true
				}
			}
		}
	}
	return false
}

// UpdateDeckLastStudied updates the LastStudied timestamp for a deck
func (s *Store) UpdateDeckLastStudied(deckID string) bool {
	for i, deck := range s.Decks {
		if deck.ID == deckID {
			s.Decks[i].LastStudied = time.Now()
			return true
		}
	}
	return false
}

// SaveCardReview updates a card with its new review data and updates
// the parent deck's LastStudied timestamp
func (s *Store) SaveCardReview(card model.Card, rating int) bool {
	// Use the SRS algorithm to schedule the card
	updatedCard := srs.ScheduleCard(card, rating)

	// Update the card in the store
	cardUpdated := s.UpdateCard(updatedCard)

	// Update the deck's last studied timestamp
	deckUpdated := s.UpdateDeckLastStudied(card.DeckID)

	return cardUpdated && deckUpdated
}
