// File: internal/srs/toggle_test.go

package srs

import (
	"testing"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestEaseToRetentionConversion(t *testing.T) {
	testCases := []struct {
		ease     float64
		expected float64
		tolerance float64
	}{
		{1.3, 0.1, 0.01},
		{1.8, 0.27, 0.01},
		{2.5, 0.5, 0.01},
		{3.5, 0.83, 0.01},
		{4.0, 1.0, 0.01},
	}

	for _, tc := range testCases {
		result := easeToRetention(tc.ease)
		if result < tc.expected-tc.tolerance || result > tc.expected+tc.tolerance {
			t.Errorf("easeToRetention(%.1f) = %.2f, expected %.2f (±%.2f)", tc.ease, result, tc.expected, tc.tolerance)
		}
	}
}

func TestEaseToRetentionBoundaries(t *testing.T) {
	// Test boundaries
	result := easeToRetention(0)
	if result != 0.5 {
		t.Errorf("easeToRetention(0) = %.2f, expected 0.50 (default)", result)
	}

	result = easeToRetention(1.3)
	if result < 0.0 || result > 1.0 {
		t.Errorf("easeToRetention(1.3) = %.2f, expected in [0.0-1.0]", result)
	}

	result = easeToRetention(4.0)
	if result < 0.0 || result > 1.0 {
		t.Errorf("easeToRetention(4.0) = %.2f, expected in [0.0-1.0]", result)
	}
}

func TestToggleSM2ToFSRS(t *testing.T) {
	// Create a deck with SM-2 cards
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
		Cards: []model.Card{
			{
				ID:       "card1.md",
				Question: "Q1",
				Answer:   "A1",
				Ease:     2.5,
				Interval: 3,
			},
			{
				ID:       "card2.md",
				Question: "Q2",
				Answer:   "A2",
				Ease:     3.5,
				Interval: 5,
			},
		},
	}

	result := ToggleDeckAlgorithm(deck, "FSRS")

	// Verify toggle succeeded
	if !result.Success {
		t.Fatalf("ToggleDeckAlgorithm failed: %v", result.Error)
	}

	// Verify algorithm changed
	if deck.Algorithm != "FSRS" {
		t.Errorf("Expected deck algorithm FSRS, got %s", deck.Algorithm)
	}

	// Verify all cards converted
	if result.CardsConverted != 2 {
		t.Errorf("Expected 2 cards converted, got %d", result.CardsConverted)
	}

	// Verify card state
	for i, card := range deck.Cards {
		if card.Algorithm != "FSRS" {
			t.Errorf("Card %d: expected algorithm FSRS, got %s", i, card.Algorithm)
		}

		if card.Retention == 0 {
			t.Errorf("Card %d: expected non-zero retention", i)
		}

		if card.EaseBackup == 0 {
			t.Errorf("Card %d: expected non-zero ease_backup", i)
		}

		// Verify retention is in valid range
		if card.Retention < 0 || card.Retention > 1.0 {
			t.Errorf("Card %d: retention %.2f out of range [0.0-1.0]", i, card.Retention)
		}
	}
}

func TestToggleFSRSToSM2(t *testing.T) {
	// Create a deck with FSRS cards that have backups
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "FSRS",
		Cards: []model.Card{
			{
				ID:        "card1.md",
				Question:  "Q1",
				Answer:    "A1",
				Ease:      0, // Cleared after SM2->FSRS
				Retention: 0.5,
				EaseBackup: 2.5, // Original ease
			},
			{
				ID:        "card2.md",
				Question:  "Q2",
				Answer:    "A2",
				Ease:      0,
				Retention: 0.75,
				EaseBackup: 3.5,
			},
		},
	}

	result := ToggleDeckAlgorithm(deck, "SM2")

	// Verify toggle succeeded
	if !result.Success {
		t.Fatalf("ToggleDeckAlgorithm failed: %v", result.Error)
	}

	// Verify algorithm changed
	if deck.Algorithm != "SM2" {
		t.Errorf("Expected deck algorithm SM2, got %s", deck.Algorithm)
	}

	// Verify all cards converted
	if result.CardsConverted != 2 {
		t.Errorf("Expected 2 cards converted, got %d", result.CardsConverted)
	}

	// Verify exact restoration from backup
	if deck.Cards[0].Ease != 2.5 {
		t.Errorf("Card 0: expected ease 2.5 (restored from backup), got %.1f", deck.Cards[0].Ease)
	}

	if deck.Cards[1].Ease != 3.5 {
		t.Errorf("Card 1: expected ease 3.5 (restored from backup), got %.1f", deck.Cards[1].Ease)
	}

	// Verify retention and backup cleared
	for i, card := range deck.Cards {
		if card.Retention != 0 {
			t.Errorf("Card %d: expected retention 0, got %.2f", i, card.Retention)
		}

		if card.EaseBackup != 0 {
			t.Errorf("Card %d: expected ease_backup 0, got %.1f", i, card.EaseBackup)
		}
	}
}

func TestToggleSameAlgorithm(t *testing.T) {
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
		Cards: []model.Card{
			{
				ID:     "card1.md",
				Ease:   2.5,
				Interval: 3,
			},
		},
	}

	result := ToggleDeckAlgorithm(deck, "SM2")

	// Should succeed with 0 conversions
	if !result.Success {
		t.Fatalf("ToggleDeckAlgorithm failed: %v", result.Error)
	}

	if result.CardsConverted != 0 {
		t.Errorf("Expected 0 conversions for same algorithm, got %d", result.CardsConverted)
	}

	// Card should be unchanged
	if deck.Cards[0].Ease != 2.5 {
		t.Errorf("Card ease changed unexpectedly")
	}
}

func TestToggleNewCards(t *testing.T) {
	// Toggle with new cards (Ease=0, no prior history)
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
		Cards: []model.Card{
			{
				ID:       "card1.md",
				Question: "Q1",
				Ease:     0, // New card
				Interval: 0,
			},
		},
	}

	result := ToggleDeckAlgorithm(deck, "FSRS")

	if !result.Success {
		t.Fatalf("ToggleDeckAlgorithm failed: %v", result.Error)
	}

	// New card should get default retention
	if deck.Cards[0].Retention != 0.5 {
		t.Errorf("Expected new card retention 0.5, got %.2f", deck.Cards[0].Retention)
	}

	// New card should NOT have backup
	if deck.Cards[0].EaseBackup != 0 {
		t.Errorf("Expected new card EaseBackup 0, got %.1f", deck.Cards[0].EaseBackup)
	}
}

func TestToggleMixedCards(t *testing.T) {
	// Toggle with both old (reviewed) and new (never reviewed) cards
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
		Cards: []model.Card{
			{
				ID:       "card1.md",
				Ease:     2.5, // Old card
				Interval: 5,
			},
			{
				ID:       "card2.md",
				Ease:     0, // New card
				Interval: 0,
			},
		},
	}

	result := ToggleDeckAlgorithm(deck, "FSRS")

	if !result.Success {
		t.Fatalf("ToggleDeckAlgorithm failed: %v", result.Error)
	}

	// Old card: retention from conversion, backup set
	if deck.Cards[0].EaseBackup != 2.5 {
		t.Errorf("Old card: expected EaseBackup 2.5, got %.1f", deck.Cards[0].EaseBackup)
	}

	// New card: default retention, no backup
	if deck.Cards[1].Retention != 0.5 {
		t.Errorf("New card: expected retention 0.5, got %.2f", deck.Cards[1].Retention)
	}

	if deck.Cards[1].EaseBackup != 0 {
		t.Errorf("New card: expected EaseBackup 0, got %.1f", deck.Cards[1].EaseBackup)
	}
}

func TestToggleInvalidAlgorithm(t *testing.T) {
	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
	}

	result := ToggleDeckAlgorithm(deck, "INVALID")

	if result.Success {
		t.Errorf("Expected toggle to fail for invalid algorithm")
	}

	if result.Error == nil {
		t.Errorf("Expected error for invalid algorithm")
	}
}

func TestRollbackDeckAlgorithm(t *testing.T) {
	// Create original deck
	originalCard := model.Card{
		ID:        "card1.md",
		Ease:      2.5,
		Retention: 0,
		Algorithm: "SM2",
	}

	deck := &model.Deck{
		ID:        "test-deck",
		Algorithm: "SM2",
		Cards:     []model.Card{originalCard},
	}

	// Create backup
	backup := make([]model.Card, 1)
	backup[0] = originalCard

	// Modify deck (simulate failed toggle)
	deck.Algorithm = "FSRS"
	deck.Cards[0].Algorithm = "FSRS"
	deck.Cards[0].Retention = 0.5

	// Rollback
	RollbackDeckAlgorithm(deck, backup, "SM2")

	// Verify restored
	if deck.Algorithm != "SM2" {
		t.Errorf("Expected algorithm SM2 after rollback, got %s", deck.Algorithm)
	}

	if deck.Cards[0].Ease != 2.5 {
		t.Errorf("Expected ease 2.5 after rollback, got %.1f", deck.Cards[0].Ease)
	}

	if deck.Cards[0].Retention != 0 {
		t.Errorf("Expected retention 0 after rollback, got %.2f", deck.Cards[0].Retention)
	}
}
