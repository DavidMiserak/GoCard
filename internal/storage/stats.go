// Package storage implements the file-based storage system for GoCard.
// This file contains operations for calculating and retrieving statistics.
package storage

import (
	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
)

// GetReviewStats returns statistics about the review process
func (s *CardStore) GetReviewStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalCards := len(s.Cards)
	dueCards := len(s.GetDueCards())

	// Count cards by interval ranges
	newCards := 0
	young := 0  // 1-7 days
	mature := 0 // > 7 days

	for _, cardObj := range s.Cards {
		if cardObj.ReviewInterval == 0 {
			newCards++
		} else if cardObj.ReviewInterval <= 7 {
			young++
		} else {
			mature++
		}
	}

	stats["total_cards"] = totalCards
	stats["due_cards"] = dueCards
	stats["new_cards"] = newCards
	stats["young_cards"] = young
	stats["mature_cards"] = mature

	return stats
}

// GetDeckStats returns statistics about a specific deck
func (s *CardStore) GetDeckStats(deckObj *deck.Deck) map[string]interface{} {
	stats := make(map[string]interface{})

	allCards := deckObj.GetAllCards()
	totalCards := len(allCards)

	// Get due cards
	var dueCards []*card.Card
	for _, cardObj := range allCards {
		if algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
		}
	}

	// Count cards by interval ranges
	newCards := 0
	young := 0  // 1-7 days
	mature := 0 // > 7 days

	for _, cardObj := range allCards {
		if cardObj.ReviewInterval == 0 {
			newCards++
		} else if cardObj.ReviewInterval <= 7 {
			young++
		} else {
			mature++
		}
	}

	stats["total_cards"] = totalCards
	stats["due_cards"] = len(dueCards)
	stats["new_cards"] = newCards
	stats["young_cards"] = young
	stats["mature_cards"] = mature
	stats["sub_decks"] = len(deckObj.SubDecks)
	stats["direct_cards"] = len(deckObj.Cards)

	return stats
}
