// test/integration/card_review_test.go
package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/service/card"
	"github.com/DavidMiserak/GoCard/internal/service/deck"
	"github.com/DavidMiserak/GoCard/internal/service/review"
	"github.com/DavidMiserak/GoCard/internal/service/storage"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// setupIntegrationTest creates a test environment with a temporary directory and services
func setupIntegrationTest(t *testing.T) (string, *storage.FileSystemStorage, *card.DefaultCardService, *deck.DefaultDeckService, *review.DefaultReviewService, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gocard-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Initialize services with real implementations
	storageService := storage.NewFileSystemStorage()
	if err := storageService.Initialize(tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create the algorithm
	alg := algorithm.NewSM2Algorithm()

	// Create services
	cardService := card.NewCardService(storageService, alg).(*card.DefaultCardService)
	deckService := deck.NewDeckService(storageService, cardService).(*deck.DefaultDeckService)
	reviewService := review.NewReviewService(storageService, cardService, deckService, alg).(*review.DefaultReviewService)

	// Return cleanup function
	cleanup := func() {
		storageService.Close()
		os.RemoveAll(tempDir)
	}

	return tempDir, storageService, cardService, deckService, reviewService, cleanup
}

// createSampleDeck creates a test deck with sample cards
func createSampleDeck(t *testing.T, rootDir string) (string, []string) {
	// Create deck directory
	deckPath := filepath.Join(rootDir, "TestDeck")
	if err := os.MkdirAll(deckPath, 0755); err != nil {
		t.Fatalf("Failed to create deck directory: %v", err)
	}

	// Sample card content
	cardContents := []struct {
		filename string
		content  string
	}{
		{
			filename: "new_card.md",
			content: `---
title: New Card
tags:
  - test
difficulty: 3
---
# Question

What is this card testing?

---

Testing new cards in review sessions.
`,
		},
		{
			filename: "due_card.md",
			content: `---
title: Due Card
tags:
  - test
difficulty: 2
last_reviewed: 2023-01-01
review_interval: 7
---
# Question

Is this card due for review?

---

Yes, this card is due for review!
`,
		},
		{
			filename: "not_due_card.md",
			content: `---
title: Not Due Card
tags:
  - test
difficulty: 1
last_reviewed: 2099-01-01
review_interval: 14
---
# Question

Is this card due for review?

---

No, this card is not due for review yet.
`,
		},
	}

	var cardPaths []string
	for _, card := range cardContents {
		cardPath := filepath.Join(deckPath, card.filename)
		if err := os.WriteFile(cardPath, []byte(card.content), 0644); err != nil {
			t.Fatalf("Failed to create card file %s: %v", card.filename, err)
		}
		cardPaths = append(cardPaths, cardPath)
	}

	return deckPath, cardPaths
}

// TestCardReviewCycle tests the full cycle of reviewing cards
func TestCardReviewCycle(t *testing.T) {
	// Setup test environment
	rootDir, _, cardService, deckService, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create sample deck and cards
	deckPath, _ := createSampleDeck(t, rootDir)

	// STEP 1: Get due cards directly from deck service
	dueCards, err := deckService.GetDueCards(deckPath)
	if err != nil {
		t.Fatalf("Failed to get due cards: %v", err)
	}

	// Should have 2 due cards (new_card and due_card)
	if len(dueCards) != 2 {
		t.Errorf("Expected 2 due cards, got %d", len(dueCards))
	}

	// Keep track of reviewed cards for verification
	reviewedCardPaths := make(map[string]bool)

	// Process each due card individually to avoid session state issues
	for _, dueCard := range dueCards {
		cardPath := dueCard.FilePath

		t.Logf("Processing card: %s (path: %s)", dueCard.Title, cardPath)

		// Record the original state
		originalCard, err := cardService.GetCard(cardPath)
		if err != nil {
			t.Fatalf("Failed to get original card state: %v", err)
		}

		wasNewCard := originalCard.LastReviewed.IsZero()
		originalInterval := originalCard.ReviewInterval

		t.Logf("Before review - Card: %s, LastReviewed: %v, Interval: %d",
			originalCard.Title, originalCard.LastReviewed, originalCard.ReviewInterval)

		// Choose rating based on the card
		var rating int
		if originalCard.Title == "New Card" {
			rating = 5 // Easy
		} else {
			rating = 3 // Hard but recalled
		}

		// Review the card directly without using a session
		err = cardService.ReviewCard(cardPath, rating)
		if err != nil {
			t.Fatalf("Failed to review card %s: %v", originalCard.Title, err)
		}

		reviewedCardPaths[cardPath] = true

		// Get updated card state
		updatedCard, err := cardService.GetCard(cardPath)
		if err != nil {
			t.Fatalf("Failed to get updated card state: %v", err)
		}

		t.Logf("After review - Card: %s, LastReviewed: %v, Interval: %d",
			updatedCard.Title, updatedCard.LastReviewed, updatedCard.ReviewInterval)

		// Verify LastReviewed was updated
		if updatedCard.LastReviewed.IsZero() {
			t.Errorf("LastReviewed should be set after review: %s", updatedCard.Title)
		}

		if !wasNewCard && !updatedCard.LastReviewed.After(originalCard.LastReviewed) {
			t.Errorf("LastReviewed should be updated for card: %s. Before: %v, After: %v",
				updatedCard.Title, originalCard.LastReviewed, updatedCard.LastReviewed)
		}

		// Verify interval was updated correctly
		if rating <= 2 {
			if updatedCard.ReviewInterval != 1 {
				t.Errorf("Failed card should have interval reset to 1, got %d", updatedCard.ReviewInterval)
			}
		} else if wasNewCard {
			// New cards with successful rating should have specific intervals based on SM-2
			if rating == 3 && updatedCard.ReviewInterval != 1 {
				t.Errorf("New card with rating 3 should have interval of 1, got %d", updatedCard.ReviewInterval)
			} else if rating == 4 && updatedCard.ReviewInterval != 2 {
				t.Errorf("New card with rating 4 should have interval of 2, got %d", updatedCard.ReviewInterval)
			} else if rating == 5 && updatedCard.ReviewInterval != 3 {
				t.Errorf("New card with rating 5 should have interval of 3, got %d", updatedCard.ReviewInterval)
			}
		} else if originalInterval > 0 && rating > 2 {
			// Reviewed cards with successful rating should increase interval
			if updatedCard.ReviewInterval <= originalInterval {
				t.Errorf("Card interval should increase after successful review, original: %d, new: %d",
					originalInterval, updatedCard.ReviewInterval)
			}
		}

		// Check that the file was updated
		content, err := os.ReadFile(cardPath)
		if err != nil {
			t.Fatalf("Failed to read card file: %v", err)
		}

		contentStr := string(content)
		t.Logf("Card file content:\n%s", contentStr)

		if !strings.Contains(contentStr, "last_reviewed:") {
			t.Errorf("Card file should contain last_reviewed field in frontmatter")
		}
		if !strings.Contains(contentStr, "review_interval:") {
			t.Errorf("Card file should contain review_interval field in frontmatter")
		}
	}

	// Verify all cards were reviewed
	if len(reviewedCardPaths) != 2 {
		t.Errorf("Expected to review 2 cards, actually reviewed %d", len(reviewedCardPaths))
	}

	// STEP 7: Verify cards are no longer due
	dueCardsAfter, err := deckService.GetDueCards(deckPath)
	if err != nil {
		t.Fatalf("Failed to get due cards after review: %v", err)
	}

	// None of the reviewed cards should be due now
	if len(dueCardsAfter) != 0 {
		t.Errorf("Expected 0 due cards after review, got %d", len(dueCardsAfter))
		for _, card := range dueCardsAfter {
			t.Logf("Still due: %s (last reviewed: %v, interval: %d)",
				card.Title, card.LastReviewed, card.ReviewInterval)
		}
	}
}

// TestSimpleReviewSession tests a simplified version of the review process
func TestSimpleReviewSession(t *testing.T) {
	// Setup test environment
	rootDir, _, _, deckService, reviewService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create sample deck and cards
	deckPath, _ := createSampleDeck(t, rootDir)

	// Get initial due cards
	dueCards, err := deckService.GetDueCards(deckPath)
	if err != nil {
		t.Fatalf("Failed to get initial due cards: %v", err)
	}
	initialDueCount := len(dueCards)
	t.Logf("Due cards before review: %d", initialDueCount)

	// Debug: Log all due card details
	for _, card := range dueCards {
		t.Logf("Due Card: %s (Path: %s)", card.Title, card.FilePath)
	}

	// Start a review session
	session, err := reviewService.StartSession(deckPath)
	if err != nil {
		t.Fatalf("Failed to start review session: %v", err)
	}

	// Debug: Log session details
	t.Logf("Session card paths: %v", session.CardPaths)
	t.Logf("Session total cards: %d", len(session.CardPaths))
	t.Logf("Session current card: %d", session.CurrentCard)

	// Process cards
	processedCount := 0
	for processedCount < initialDueCount {
		// Check session completion status
		if session.IsComplete() {
			t.Logf("Session reported complete after processing %d cards", processedCount)
			break
		}

		// Get the next card
		card, err := reviewService.GetNextCard()
		if err != nil {
			t.Fatalf("Failed to get next card (processed %d): %v", processedCount, err)
		}

		t.Logf("Reviewing card %d: %s", processedCount+1, card.Title)

		// Submit a fixed rating
		err = reviewService.SubmitRating(5) // Always "Easy"
		if err != nil {
			t.Fatalf("Failed to submit rating for card %d: %v", processedCount+1, err)
		}

		processedCount++

		// Retrieve the updated session to check its state
		updatedSession, err := reviewService.GetSession()
		if err != nil {
			t.Fatalf("Failed to get updated session: %v", err)
		}
		session = updatedSession

		t.Logf("After review - Current card: %d, Is Complete: %v",
			session.CurrentCard, session.IsComplete())
	}

	// End the session
	summary, err := reviewService.EndSession()
	if err != nil {
		t.Fatalf("Failed to end review session: %v", err)
	}

	t.Logf("Session summary: %d cards reviewed", summary.CardsReviewed)

	// Verify the number of cards reviewed
	if summary.CardsReviewed != initialDueCount {
		t.Errorf("Expected to review %d cards, but reviewed %d",
			initialDueCount, summary.CardsReviewed)
	}

	// Verify cards are no longer due
	dueCardsAfter, err := deckService.GetDueCards(deckPath)
	if err != nil {
		t.Fatalf("Failed to get due cards after review: %v", err)
	}
	if len(dueCardsAfter) != 0 {
		t.Errorf("Expected 0 due cards after review, got %d", len(dueCardsAfter))
		for _, card := range dueCardsAfter {
			t.Logf("Still due: %s", card.Title)
		}
	}
}

// TestEdgeCases tests various edge cases in the review process
func TestEdgeCases(t *testing.T) {
	// Setup test environment
	rootDir, _, _, _, reviewService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create sample deck and cards
	deckPath, _ := createSampleDeck(t, rootDir)

	// Test: Empty deck with no due cards
	// First, review all cards to make them not due
	session, _ := reviewService.StartSession(deckPath)
	processedCount := 0

	for !session.IsComplete() && processedCount < 10 { // Safety limit
		processedCount++
		_, err := reviewService.GetNextCard()
		if err != nil {
			break
		}
		_ = reviewService.SubmitRating(5)
	}

	_, _ = reviewService.EndSession()

	// Now try to start a new session
	emptySession, err := reviewService.StartSession(deckPath)
	if err != nil {
		t.Fatalf("StartSession should succeed even with no due cards: %v", err)
	}
	if len(emptySession.CardPaths) != 0 {
		t.Errorf("Session should have 0 cards when none are due, got %d", len(emptySession.CardPaths))
	}

	// Verify the session is already complete
	if !emptySession.IsComplete() {
		t.Errorf("Empty session should be marked as complete")
	}

	// Test: Out-of-bounds session operations
	_, err = reviewService.GetNextCard()
	if err == nil {
		t.Errorf("GetNextCard should return error when session is complete")
	}

	err = reviewService.SubmitRating(5)
	if err == nil {
		t.Errorf("SubmitRating should return error when session is complete")
	}
}
