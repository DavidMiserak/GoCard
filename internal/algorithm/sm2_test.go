// File: internal/algorithm/sm2_test.go

package algorithm

import (
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
)

func TestSM2Algorithm(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test case 1: New card, rating = 3 (hard but correct)
	t.Run("NewCard_Rating3", func(t *testing.T) {
		testCard := &card.Card{
			Title:          "Test Card 1",
			Question:       "Question 1",
			Answer:         "Answer 1",
			Tags:           []string{"test"},
			Created:        time.Now(),
			LastReviewed:   time.Time{}, // Zero time
			ReviewInterval: 0,
			Difficulty:     0,
		}

		interval := sm2.CalculateNextReview(testCard, 3)

		if interval != 1 {
			t.Errorf("Expected interval of 1 day, got %d", interval)
		}

		if testCard.Difficulty != 3 {
			t.Errorf("Expected difficulty of 3, got %d", testCard.Difficulty)
		}

		if testCard.LastReviewed.IsZero() {
			t.Error("Expected LastReviewed to be updated")
		}
	})

	// Test case 2: New card, rating = 5 (very easy)
	t.Run("NewCard_Rating5", func(t *testing.T) {
		testCard := &card.Card{
			Title:          "Test Card 2",
			Question:       "Question 2",
			Answer:         "Answer 2",
			Tags:           []string{"test"},
			Created:        time.Now(),
			LastReviewed:   time.Time{}, // Zero time
			ReviewInterval: 0,
			Difficulty:     0,
		}

		interval := sm2.CalculateNextReview(testCard, 5)

		if interval != 5 {
			t.Errorf("Expected interval of 5 days, got %d", interval)
		}

		if testCard.Difficulty != 5 {
			t.Errorf("Expected difficulty of 5, got %d", testCard.Difficulty)
		}
	})

	// Test case 3: Previously reviewed card, rating = 4 (good)
	t.Run("ReviewedCard_Rating4", func(t *testing.T) {
		// Card with previous review and interval
		testCard := &card.Card{
			Title:          "Test Card 3",
			Question:       "Question 3",
			Answer:         "Answer 3",
			Tags:           []string{"test"},
			Created:        time.Now().AddDate(0, 0, -10), // 10 days ago
			LastReviewed:   time.Now().AddDate(0, 0, -5),  // 5 days ago
			ReviewInterval: 5,
			Difficulty:     4,
		}

		interval := sm2.CalculateNextReview(testCard, 4)

		// Expected: 5 days * 1.8 = 9 days
		expectedInterval := int(float64(5) * 1.8)
		if interval != expectedInterval {
			t.Errorf("Expected interval of %d days, got %d", expectedInterval, interval)
		}
	})

	// Test case 4: Low rating resets interval
	t.Run("LowRating_ResetsInterval", func(t *testing.T) {
		testCard := &card.Card{
			Title:          "Test Card 4",
			Question:       "Question 4",
			Answer:         "Answer 4",
			Tags:           []string{"test"},
			Created:        time.Now().AddDate(0, 0, -20), // 20 days ago
			LastReviewed:   time.Now().AddDate(0, 0, -10), // 10 days ago
			ReviewInterval: 15,
			Difficulty:     4,
		}

		interval := sm2.CalculateNextReview(testCard, 2)

		if interval != 1 {
			t.Errorf("Expected interval to reset to 1 day, got %d", interval)
		}

		if testCard.Difficulty != 2 {
			t.Errorf("Expected difficulty of 2, got %d", testCard.Difficulty)
		}
	})

	// Test case 5: Test IsDue function
	t.Run("IsDue_Function", func(t *testing.T) {
		// Card that is due (last reviewed 10 days ago, interval was 5 days)
		dueCard := &card.Card{
			Title:          "Due Card",
			LastReviewed:   time.Now().AddDate(0, 0, -10),
			ReviewInterval: 5,
		}

		if !sm2.IsDue(dueCard) {
			t.Error("Expected card to be due")
		}

		// Card that is not due yet (last reviewed 2 days ago, interval is 5 days)
		notDueCard := &card.Card{
			Title:          "Not Due Card",
			LastReviewed:   time.Now().AddDate(0, 0, -2),
			ReviewInterval: 5,
		}

		if sm2.IsDue(notDueCard) {
			t.Error("Expected card to not be due")
		}

		// New card should be due
		newCard := &card.Card{
			Title:          "New Card",
			LastReviewed:   time.Time{}, // Zero time
			ReviewInterval: 0,
		}

		if !sm2.IsDue(newCard) {
			t.Error("Expected new card to be due")
		}
	})

	// Test case 6: Test CalculatePercentOverdue function
	t.Run("PercentOverdue", func(t *testing.T) {
		// A card that is exactly one interval overdue
		overdueDays := 5
		lastReviewed := time.Now().AddDate(0, 0, -(overdueDays * 2)) // 10 days ago

		testCard := &card.Card{
			Title:          "Overdue Card",
			LastReviewed:   lastReviewed,
			ReviewInterval: overdueDays, // 5 day interval, so it's been overdue for 5 days
		}

		percentOverdue := sm2.CalculatePercentOverdue(testCard)

		// It should be approximately 100% overdue (1 full interval)
		if percentOverdue < 95 || percentOverdue > 105 {
			t.Errorf("Expected ~100%% overdue, got %.2f%%", percentOverdue)
		}

		// A card that is not due yet
		notDueCard := &card.Card{
			Title:          "Not Due Card",
			LastReviewed:   time.Now(),
			ReviewInterval: 5,
		}

		percentOverdue = sm2.CalculatePercentOverdue(notDueCard)

		if percentOverdue != 0 {
			t.Errorf("Expected 0%% overdue for card not due yet, got %.2f%%", percentOverdue)
		}
	})

	// Test case 7: Test EstimateEase function
	t.Run("EstimateEase", func(t *testing.T) {
		// Cards with different difficulty levels
		hardCard := &card.Card{
			Title:      "Hard Card",
			Difficulty: 0,
		}

		mediumCard := &card.Card{
			Title:      "Medium Card",
			Difficulty: 3,
		}

		easyCard := &card.Card{
			Title:      "Easy Card",
			Difficulty: 5,
		}

		hardEase := sm2.EstimateEase(hardCard)
		mediumEase := sm2.EstimateEase(mediumCard)
		easyEase := sm2.EstimateEase(easyCard)

		// Verify ease values are within expected ranges
		if hardEase < 1.2 || hardEase > 1.4 {
			t.Errorf("Expected hard card ease around 1.3, got %.2f", hardEase)
		}

		if mediumEase < 1.9 || mediumEase > 2.4 {
			t.Errorf("Expected medium card ease around 2.1, got %.2f", mediumEase)
		}

		if easyEase < 2.9 || easyEase > 3.1 {
			t.Errorf("Expected easy card ease around 3.0, got %.2f", easyEase)
		}

		// Verify ordering
		if !(hardEase < mediumEase && mediumEase < easyEase) {
			t.Errorf("Ease values should increase with ease: %.2f, %.2f, %.2f",
				hardEase, mediumEase, easyEase)
		}
	})
}
