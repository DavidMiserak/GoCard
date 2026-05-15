// File: internal/data/store_toggle_test.go

package data

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestStoreToggleDeckAlgorithmSM2ToFSRS(t *testing.T) {
	// Create a temporary directory for test deck
	tempDir, err := os.MkdirTemp("", "toggle-test-sm2-fsrs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test cards
	now := time.Now()
	cards := []model.Card{
		{
			ID:           filepath.Join(tempDir, "card1.md"),
			Question:     "What is SM-2?",
			Answer:       "SuperMemo 2 algorithm",
			DeckID:       tempDir,
			LastReviewed: now.AddDate(0, 0, -5),
			NextReview:   now.AddDate(0, 0, 3),
			Ease:         2.5,
			Interval:     5,
			Algorithm:    "SM2",
		},
		{
			ID:           filepath.Join(tempDir, "card2.md"),
			Question:     "What is FSRS?",
			Answer:       "Free Spaced Repetition Scheduler",
			DeckID:       tempDir,
			LastReviewed: now.AddDate(0, 0, -10),
			NextReview:   now.AddDate(0, 0, 7),
			Ease:         3.5,
			Interval:     10,
			Algorithm:    "SM2",
		},
	}

	// Write cards to disk
	for _, card := range cards {
		if err := WriteCard(card, card.ID); err != nil {
			t.Fatalf("Failed to write card: %v", err)
		}
	}

	// Create store with deck
	deck := &model.Deck{
		ID:          tempDir,
		Name:        "Test Deck",
		Algorithm:   "SM2",
		CreatedAt:   now,
		LastStudied: now,
		Cards:       cards,
	}

	store := &Store{
		Decks: []model.Deck{*deck},
	}

	// Toggle to FSRS
	success, converted, err := store.ToggleDeckAlgorithm(tempDir, "FSRS")

	if !success {
		t.Fatalf("Toggle failed: %v", err)
	}

	if converted != 2 {
		t.Errorf("Expected 2 cards converted, got %d", converted)
	}

	// Verify in-memory deck was updated
	updatedDeck, found := store.GetDeck(tempDir)
	if !found {
		t.Fatalf("Deck not found after toggle")
	}

	if updatedDeck.Algorithm != "FSRS" {
		t.Errorf("Expected algorithm FSRS, got %s", updatedDeck.Algorithm)
	}

	// Verify cards were updated
	for i, card := range updatedDeck.Cards {
		if card.Algorithm != "FSRS" {
			t.Errorf("Card %d: expected algorithm FSRS, got %s", i, card.Algorithm)
		}

		if card.Retention == 0 {
			t.Errorf("Card %d: expected non-zero retention", i)
		}

		if card.EaseBackup == 0 {
			t.Errorf("Card %d: expected non-zero ease_backup", i)
		}
	}

	// Verify persistence: read cards from disk
	readCard1, err := ParseMarkdownFile(filepath.Join(tempDir, "card1.md"))
	if err != nil {
		t.Fatalf("Failed to read card1: %v", err)
	}

	if readCard1.FrontMatter.Algorithm != "FSRS" {
		t.Errorf("Card1 disk: expected algorithm FSRS, got %s", readCard1.FrontMatter.Algorithm)
	}

	if readCard1.FrontMatter.EaseBackup == 0 {
		t.Errorf("Card1 disk: expected non-zero ease_backup, got %.1f", readCard1.FrontMatter.EaseBackup)
	}

	if readCard1.FrontMatter.Retention == 0 {
		t.Errorf("Card1 disk: expected non-zero retention, got %.2f", readCard1.FrontMatter.Retention)
	}
}

func TestStoreToggleDeckAlgorithmFSRSToSM2(t *testing.T) {
	// Create a temporary directory for test deck
	tempDir, err := os.MkdirTemp("", "toggle-test-fsrs-sm2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create FSRS cards with backups
	now := time.Now()
	cards := []model.Card{
		{
			ID:           filepath.Join(tempDir, "card1.md"),
			Question:     "Q1",
			Answer:       "A1",
			DeckID:       tempDir,
			LastReviewed: now.AddDate(0, 0, -5),
			NextReview:   now.AddDate(0, 0, 3),
			Ease:         0,
			Retention:    0.5,
			EaseBackup:   2.5,
			Interval:     5,
			Algorithm:    "FSRS",
		},
		{
			ID:           filepath.Join(tempDir, "card2.md"),
			Question:     "Q2",
			Answer:       "A2",
			DeckID:       tempDir,
			LastReviewed: now.AddDate(0, 0, -10),
			NextReview:   now.AddDate(0, 0, 7),
			Ease:         0,
			Retention:    0.75,
			EaseBackup:   3.5,
			Interval:     10,
			Algorithm:    "FSRS",
		},
	}

	// Write cards to disk
	for _, card := range cards {
		if err := WriteCard(card, card.ID); err != nil {
			t.Fatalf("Failed to write card: %v", err)
		}
	}

	// Create store with deck
	deck := &model.Deck{
		ID:          tempDir,
		Name:        "Test Deck",
		Algorithm:   "FSRS",
		CreatedAt:   now,
		LastStudied: now,
		Cards:       cards,
	}

	store := &Store{
		Decks: []model.Deck{*deck},
	}

	// Toggle back to SM2
	success, converted, err := store.ToggleDeckAlgorithm(tempDir, "SM2")

	if !success {
		t.Fatalf("Toggle failed: %v", err)
	}

	if converted != 2 {
		t.Errorf("Expected 2 cards converted, got %d", converted)
	}

	// Verify rollback: Ease restored from EaseBackup
	updatedDeck, found := store.GetDeck(tempDir)
	if !found {
		t.Fatalf("Deck not found after toggle")
	}

	if updatedDeck.Cards[0].Ease != 2.5 {
		t.Errorf("Card 0: expected ease 2.5 (restored), got %.1f", updatedDeck.Cards[0].Ease)
	}

	if updatedDeck.Cards[1].Ease != 3.5 {
		t.Errorf("Card 1: expected ease 3.5 (restored), got %.1f", updatedDeck.Cards[1].Ease)
	}

	// Verify persistence: read from disk
	readCard1, err := ParseMarkdownFile(filepath.Join(tempDir, "card1.md"))
	if err != nil {
		t.Fatalf("Failed to read card1: %v", err)
	}

	modelCard1 := readCard1.ToModelCard(tempDir)
	if modelCard1.Ease != 2.5 {
		t.Errorf("Card1 disk: expected ease 2.5, got %.1f", modelCard1.Ease)
	}

	if modelCard1.Algorithm != "SM2" {
		t.Errorf("Card1 disk: expected algorithm SM2, got %s", modelCard1.Algorithm)
	}
}

func TestStoreToggleLargeDeck(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "toggle-test-large")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create 100 cards
	now := time.Now()
	cards := make([]model.Card, 100)
	for i := 0; i < 100; i++ {
		cards[i] = model.Card{
			ID:           filepath.Join(tempDir, fmt.Sprintf("card_%03d.md", i)),
			Question:     "Q" + string(rune(i)),
			Answer:       "A" + string(rune(i)),
			DeckID:       tempDir,
			LastReviewed: now,
			NextReview:   now.AddDate(0, 0, 3),
			Ease:         2.5,
			Interval:     3,
			Algorithm:    "SM2",
		}

		if err := WriteCard(cards[i], cards[i].ID); err != nil {
			t.Fatalf("Failed to write card %d: %v", i, err)
		}
	}

	deck := &model.Deck{
		ID:          tempDir,
		Name:        "Large Deck",
		Algorithm:   "SM2",
		CreatedAt:   now,
		LastStudied: now,
		Cards:       cards,
	}

	store := &Store{
		Decks: []model.Deck{*deck},
	}

	// Toggle to FSRS
	success, converted, err := store.ToggleDeckAlgorithm(tempDir, "FSRS")

	if !success {
		t.Fatalf("Toggle failed: %v", err)
	}

	if converted != 100 {
		t.Errorf("Expected 100 cards converted, got %d", converted)
	}

	// Verify all cards converted
	updatedDeck, _ := store.GetDeck(tempDir)
	for i, card := range updatedDeck.Cards {
		if card.Algorithm != "FSRS" {
			t.Errorf("Card %d: algorithm not FSRS", i)
		}
		if card.Retention == 0 {
			t.Errorf("Card %d: retention not set", i)
		}
	}
}

func TestStoreTogglePersistsAcrossReload(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "toggle-test-reload")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create and write a card
	now := time.Now()
	card := model.Card{
		ID:           filepath.Join(tempDir, "card1.md"),
		Question:     "Q1",
		Answer:       "A1",
		DeckID:       tempDir,
		LastReviewed: now,
		NextReview:   now.AddDate(0, 0, 3),
		Ease:         2.5,
		Interval:     3,
		Algorithm:    "SM2",
	}

	if err := WriteCard(card, card.ID); err != nil {
		t.Fatalf("Failed to write card: %v", err)
	}

	// Create first store and toggle
	deck1 := &model.Deck{
		ID:          tempDir,
		Name:        "Test",
		Algorithm:   "SM2",
		CreatedAt:   now,
		LastStudied: now,
		Cards:       []model.Card{card},
	}

	store1 := &Store{Decks: []model.Deck{*deck1}}
	success, _, err := store1.ToggleDeckAlgorithm(tempDir, "FSRS")
	if !success {
		t.Fatalf("First toggle failed: %v", err)
	}

	// Create second store and reload from disk
	deck2, err := CreateDeckFromDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to reload deck: %v", err)
	}

	store2 := &Store{Decks: []model.Deck{*deck2}}

	// Verify card-level algorithm persisted (this is what matters for MVP)
	// Note: deck.Algorithm defaults to SM2 on reload (deck-level setting not persisted to disk)
	// But card.Algorithm IS persisted, which is the critical data
	reloadedDeck, _ := store2.GetDeck(tempDir)

	if len(reloadedDeck.Cards) == 0 {
		t.Fatalf("No cards loaded after reload")
	}

	if reloadedDeck.Cards[0].Algorithm != "FSRS" {
		t.Errorf("Card algorithm not persisted: expected FSRS, got %s", reloadedDeck.Cards[0].Algorithm)
	}

	if reloadedDeck.Cards[0].Retention == 0 {
		t.Errorf("Card retention not persisted")
	}

	if reloadedDeck.Cards[0].EaseBackup == 0 {
		t.Errorf("Card ease_backup not persisted")
	}
}

func TestStoreToggleWithEaseToRetentionAccuracy(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "toggle-test-ease-retention")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create cards with specific ease values for testing formula accuracy
	now := time.Now()
	testCases := []struct {
		ease           float64
		expectedRetention float64
		tolerance      float64
	}{
		{1.3, 0.1, 0.01},
		{1.8, 0.27, 0.01},
		{2.5, 0.5, 0.01},
		{3.5, 0.83, 0.01},
		{4.0, 1.0, 0.01},
	}

	cards := make([]model.Card, len(testCases))
	for i, tc := range testCases {
		cards[i] = model.Card{
			ID:           filepath.Join(tempDir, fmt.Sprintf("card_%d.md", i)),
			Question:     "Q",
			Answer:       "A",
			DeckID:       tempDir,
			LastReviewed: now,
			NextReview:   now.AddDate(0, 0, 3),
			Ease:         tc.ease,
			Interval:     3,
			Algorithm:    "SM2",
		}

		if err := WriteCard(cards[i], cards[i].ID); err != nil {
			t.Fatalf("Failed to write card: %v", err)
		}
	}

	deck := &model.Deck{
		ID:          tempDir,
		Name:        "Test",
		Algorithm:   "SM2",
		CreatedAt:   now,
		LastStudied: now,
		Cards:       cards,
	}

	store := &Store{Decks: []model.Deck{*deck}}
	success, _, _ := store.ToggleDeckAlgorithm(tempDir, "FSRS")
	if !success {
		t.Fatalf("Toggle failed")
	}

	// Verify retention values within tolerance
	updatedDeck, _ := store.GetDeck(tempDir)
	for i, card := range updatedDeck.Cards {
		expected := testCases[i].expectedRetention
		tolerance := testCases[i].tolerance
		if card.Retention < expected-tolerance || card.Retention > expected+tolerance {
			t.Errorf("Card %d: retention %.2f outside expected range [%.2f-%.2f]",
				i, card.Retention, expected-tolerance, expected+tolerance)
		}
	}
}
