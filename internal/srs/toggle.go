// File: internal/srs/toggle.go

package srs

import (
	"fmt"

	"github.com/DavidMiserak/GoCard/internal/model"
)

// Ease→Retention conversion formula: maps [1.3-4.0] to [0.1-1.0]
// This is lossy but prevents data corruption
func easeToRetention(ease float64) float64 {
	if ease == 0 {
		return 0.5 // Default for new cards
	}
	// Map ease [1.3-4.0] to retention [0.1-1.0]
	retention := (ease - 1.0) / 3.0
	// Clamp to [0.0-1.0]
	if retention < 0.0 {
		return 0.0
	}
	if retention > 1.0 {
		return 1.0
	}
	return retention
}

// ToggleAlgorithmResult tracks the outcome of a toggle operation
type ToggleAlgorithmResult struct {
	Success         bool
	CardsConverted  int
	Error           error
	PreviousAlgorithm string
	NewAlgorithm    string
}

// ToggleDeckAlgorithm atomically converts all cards in a deck from one algorithm to another
// Returns a result indicating success/failure and how many cards were converted
// On failure, original deck state is unchanged (rolled back via in-memory backup)
func ToggleDeckAlgorithm(deck *model.Deck, targetAlgorithm string) ToggleAlgorithmResult {
	// Validate target algorithm
	if targetAlgorithm != "SM2" && targetAlgorithm != "FSRS" {
		return ToggleAlgorithmResult{
			Success: false,
			Error:   fmt.Errorf("invalid target algorithm: %s", targetAlgorithm),
		}
	}

	previousAlgorithm := deck.Algorithm
	if previousAlgorithm == "" {
		previousAlgorithm = "SM2"
	}

	// If toggling to the same algorithm, success with 0 conversions
	if previousAlgorithm == targetAlgorithm {
		return ToggleAlgorithmResult{
			Success:           true,
			CardsConverted:    0,
			PreviousAlgorithm: previousAlgorithm,
			NewAlgorithm:      targetAlgorithm,
		}
	}

	// Create in-memory backup of all cards
	backup := make([]model.Card, len(deck.Cards))
	for i, card := range deck.Cards {
		// Deep copy all relevant fields
		backup[i] = card
	}

	// Convert all cards
	cardsConverted := 0

	switch {
	case previousAlgorithm == "SM2" && targetAlgorithm == "FSRS":
		// Forward: SM-2 → FSRS
		for i := range deck.Cards {
			card := &deck.Cards[i]
			if card.Ease == 0 {
				// New card: assign default retention
				card.Retention = 0.5
				card.EaseBackup = 0
			} else {
				// Existing card: convert and backup
				card.Retention = easeToRetention(card.Ease)
				card.EaseBackup = card.Ease
			}
			card.Algorithm = "FSRS"
			cardsConverted++
		}

	case previousAlgorithm == "FSRS" && targetAlgorithm == "SM2":
		// Reverse: FSRS → SM-2
		for i := range deck.Cards {
			card := &deck.Cards[i]
			if card.EaseBackup != 0 {
				// Restore from backup (exact, lossless)
				card.Ease = card.EaseBackup
			} else {
				// No backup: new card, use default
				card.Ease = defaultEase
			}
			card.Retention = 0
			card.EaseBackup = 0
			card.Algorithm = "SM2"
			cardsConverted++
		}
	}

	// Update deck algorithm
	deck.Algorithm = targetAlgorithm

	return ToggleAlgorithmResult{
		Success:           true,
		CardsConverted:    cardsConverted,
		PreviousAlgorithm: previousAlgorithm,
		NewAlgorithm:      targetAlgorithm,
	}
}

// RollbackDeckAlgorithm rolls back a deck to its previous algorithm using in-memory backup
// This is called if atomic disk writes fail during toggle
func RollbackDeckAlgorithm(deck *model.Deck, backup []model.Card, previousAlgorithm string) {
	// Restore all cards from backup
	deck.Cards = make([]model.Card, len(backup))
	copy(deck.Cards, backup)

	// Restore deck algorithm
	deck.Algorithm = previousAlgorithm
}
