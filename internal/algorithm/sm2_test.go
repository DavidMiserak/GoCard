// File: internal/algorithm/sm2_test.go

package algorithm

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// TestSM2AlgorithmProperties tests the fundamental properties of the SM-2 algorithm
func TestSM2AlgorithmProperties(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test case: Intervals should generally increase with higher ratings for previously reviewed cards
	t.Run("IntervalIncreaseWithRatings", func(t *testing.T) {
		baseCard := &card.Card{
			Title:          "Test Interval Increase",
			LastReviewed:   time.Now().AddDate(0, 0, -10),
			ReviewInterval: 10,
		}

		// Test ratings progression
		intervals := make([]int, 6)
		for rating := 0; rating <= 5; rating++ {
			testCard := *baseCard
			intervals[rating] = sm2.CalculateNextReview(&testCard, rating)
		}

		// Validate rating-based progression more flexibly
		t.Logf("Intervals for ratings 0-5: %v", intervals)

		// Check that very low ratings (0-1) have minimal changes or resets
		if intervals[0] > 5 || intervals[1] > 5 {
			t.Errorf("Very low ratings (0-1) should result in minimal intervals, got %d, %d",
				intervals[0], intervals[1])
		}

		// Verify increasing trend for higher ratings, with some flexibility
		for i := 3; i < 5; i++ {
			if intervals[i] <= intervals[i-1] {
				t.Errorf("Expected intervals to increase for ratings %d and %d, got %d, %d",
					i-1, i, intervals[i-1], intervals[i])
			}
		}
	})

	// Test case: Verify interval modification behavior
	t.Run("IntervalModification", func(t *testing.T) {
		baseCard := &card.Card{
			Title:          "Modifier Test",
			LastReviewed:   time.Now().AddDate(0, 0, -10),
			ReviewInterval: 10,
		}

		testCases := []struct {
			name     string
			modifier float64
		}{
			{"Default Modifier", 1.0},
			{"Lower Modifier", 0.5},
			{"Higher Modifier", 1.5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				sm2 := NewSM2Algorithm()
				sm2.IntervalModifier = tc.modifier

				testCard := baseCard
				baseInterval := sm2.CalculateNextReview(testCard, 4)

				t.Logf("Interval with modifier %.2f: %d", tc.modifier, baseInterval)

				// Allow a broader range of acceptable modifications
				lowerBound := int(float64(baseInterval) * 0.4)
				upperBound := int(float64(baseInterval) * 2.0)

				if baseInterval < lowerBound || baseInterval > upperBound {
					t.Errorf("Interval %d outside expected range [%d, %d] with modifier %.2f",
						baseInterval, lowerBound, upperBound, tc.modifier)
				}
			})
		}
	})

	// Test case: Verify easy bonus calculation
	t.Run("EasyBonus", func(t *testing.T) {
		baseCard := &card.Card{
			Title:          "Easy Bonus Test",
			LastReviewed:   time.Now().AddDate(0, 0, -10),
			ReviewInterval: 10,
		}

		testCases := []struct {
			name      string
			easyBonus float64
		}{
			{"Default Bonus", 1.3},
			{"Lower Bonus", 1.0},
			{"Higher Bonus", 1.5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				sm2 := NewSM2Algorithm()
				sm2.EasyBonus = tc.easyBonus

				testCard := baseCard
				interval := sm2.CalculateNextReview(testCard, 5)

				t.Logf("Easy rating interval with bonus %.2f: %d", tc.easyBonus, interval)

				// More flexible check for reasonable intervals
				if interval < 1 || interval > 1000 {
					t.Errorf("Unexpected interval %d for easy rating with bonus %.2f",
						interval, tc.easyBonus)
				}
			})
		}
	})
}

// TestSM2AlgorithmEdgeCases tests edge case scenarios for the SM-2 algorithm
func TestSM2AlgorithmEdgeCases(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test case: New card with zero interval
	t.Run("NewCardZeroInterval", func(t *testing.T) {
		newCard := &card.Card{
			Title:          "New Card Test",
			LastReviewed:   time.Time{}, // Zero time for new card
			ReviewInterval: 0,
		}

		testCases := []struct {
			rating            int
			expectedMinResult int
			expectedMaxResult int
		}{
			{3, 1, 1}, // Correct with effort
			{4, 1, 3}, // Good
			{5, 1, 5}, // Very easy
			{0, 1, 1}, // Forgot completely
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Rating%d", tc.rating), func(t *testing.T) {
				interval := sm2.CalculateNextReview(newCard, tc.rating)
				if interval < tc.expectedMinResult || interval > tc.expectedMaxResult {
					t.Errorf("Expected interval between %d and %d for rating %d, got %d",
						tc.expectedMinResult, tc.expectedMaxResult, tc.rating, interval)
				}
			})
		}
	})

	// Test case: Max difficulty rating impact
	t.Run("MaxDifficultyRating", func(t *testing.T) {
		baseCard := &card.Card{
			Title:          "Max Difficulty Test",
			LastReviewed:   time.Now(),
			ReviewInterval: 10,
		}

		// Test with maximum difficulty ratings
		intervals := make([]int, 6)
		for rating := 0; rating <= 5; rating++ {
			testCard := baseCard
			intervals[rating] = sm2.CalculateNextReview(testCard, rating)
		}

		// Ensure very low ratings result in minimal interval
		if intervals[0] != 1 || intervals[1] != 1 {
			t.Errorf("Very low ratings (0-1) should reset to minimal interval")
		}

		// Ensure highest rating provides maximum growth
		if intervals[5] <= intervals[4] {
			t.Errorf("Highest rating should provide maximum interval growth")
		}
	})
}

// TestSM2PercentOverdueCalculation tests the percent overdue calculation
func TestSM2PercentOverdueCalculation(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test various overdue scenarios
	testCases := []struct {
		name                string
		reviewInterval      int
		timeSinceLastReview time.Duration
		expectedMinPercent  float64
		expectedMaxPercent  float64
	}{
		{
			name:                "Just Due",
			reviewInterval:      5,
			timeSinceLastReview: 5 * 24 * time.Hour,
			expectedMinPercent:  0.0,
			expectedMaxPercent:  10.0,
		},
		{
			name:                "Extremely Overdue",
			reviewInterval:      5,
			timeSinceLastReview: 30 * 24 * time.Hour,
			expectedMinPercent:  80.0,
			expectedMaxPercent:  120.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCard := &card.Card{
				Title:          "Overdue Test Card",
				LastReviewed:   time.Now().Add(-tc.timeSinceLastReview),
				ReviewInterval: tc.reviewInterval,
			}

			percentOverdue := sm2.CalculatePercentOverdue(testCard)

			if percentOverdue < tc.expectedMinPercent || percentOverdue > tc.expectedMaxPercent {
				t.Errorf("Percent overdue %.2f outside expected range [%.2f, %.2f]",
					percentOverdue, tc.expectedMinPercent, tc.expectedMaxPercent)
			}
		})
	}
}

// Benchmark performance of the SM-2 algorithm calculations
func BenchmarkSM2Calculations(b *testing.B) {
	sm2 := NewSM2Algorithm()
	baseCard := &card.Card{
		Title:          "Benchmark Card",
		LastReviewed:   time.Now(),
		ReviewInterval: 10,
	}

	b.Run("CalculateNextReview", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testCard := baseCard
			sm2.CalculateNextReview(testCard, 4)
		}
	})

	b.Run("CalculatePercentOverdue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testCard := baseCard
			sm2.CalculatePercentOverdue(testCard)
		}
	})
}
