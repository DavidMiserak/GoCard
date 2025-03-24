// File: internal/storage/review.go

// Package storage implements the file-based storage system for GoCard.
// This file contains operations related to reviewing cards and spaced repetition.
package storage

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
)

// ReviewCard reviews a card with the given difficulty rating (0-5)
// and saves the updated card to disk
func (s *CardStore) ReviewCard(cardObj *card.Card, rating int) error {
	// Apply the SM-2 algorithm to calculate the next review date
	algorithm.SM2.CalculateNextReview(cardObj, rating)

	// Save the updated card to disk
	return s.SaveCard(cardObj)
}

// GetDueCards returns all cards that are due for review
func (s *CardStore) GetDueCards() []*card.Card {
	var dueCards []*card.Card

	for _, cardObj := range s.Cards {
		if algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
		}
	}

	return dueCards
}

// GetDueCardsInDeck returns due cards in a specific deck and its subdecks
func (s *CardStore) GetDueCardsInDeck(deckObj *deck.Deck) []*card.Card {
	var dueCards []*card.Card
	seen := make(map[string]bool) // Track filepaths we've already seen

	// Get all cards in this deck and its subdecks
	allCards := deckObj.GetAllCards()

	// Filter for due cards
	for _, cardObj := range allCards {
		if !seen[cardObj.FilePath] && algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
			seen[cardObj.FilePath] = true
		}
	}

	return dueCards
}

// GetNextDueDate returns the date when the next card will be due
func (s *CardStore) GetNextDueDate() time.Time {
	var nextDue time.Time

	// Set nextDue to far future initially
	nextDue = time.Now().AddDate(10, 0, 0)

	for _, cardObj := range s.Cards {
		// Skip cards that are already due
		if algorithm.SM2.IsDue(cardObj) {
			return time.Now()
		}

		cardDueDate := algorithm.SM2.CalculateDueDate(cardObj)
		if cardDueDate.Before(nextDue) {
			nextDue = cardDueDate
		}
	}

	return nextDue
}
