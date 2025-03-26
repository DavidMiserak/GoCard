// pkg/algorithm/sm2_test.go
package algorithm

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
)

func TestNewSM2Algorithm(t *testing.T) {
	sm2 := NewSM2Algorithm()

	if sm2.EasyBonus != DefaultEasyBonus {
		t.Errorf("expected EasyBonus to be %f, got %f", DefaultEasyBonus, sm2.EasyBonus)
	}

	if sm2.IntervalModifier != DefaultIntervalModifier {
		t.Errorf("expected IntervalModifier to be %f, got %f", DefaultIntervalModifier, sm2.IntervalModifier)
	}

	if sm2.MaxInterval != DefaultMaxInterval {
		t.Errorf("expected MaxInterval to be %d, got %d", DefaultMaxInterval, sm2.MaxInterval)
	}
}

func TestCalculateNextInterval(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test cases
	testCases := []struct {
		name                 string
		card                 domain.Card
		rating               int
		expectedNextInterval int
	}{
		{
			name: "new card with rating 0 (fail)",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 0,
			},
			rating:               0,
			expectedNextInterval: 1, // Reset to 1 day
		},
		{
			name: "new card with rating 3 (difficult)",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 0,
			},
			rating:               3,
			expectedNextInterval: 1,
		},
		{
			name: "new card with rating 4 (good)",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 0,
			},
			rating:               4,
			expectedNextInterval: 2,
		},
		{
			name: "new card with rating 5 (easy)",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 0,
			},
			rating:               5,
			expectedNextInterval: 3,
		},
		{
			name: "established card (interval 10) with rating 2 (fail)",
			card: domain.Card{
				LastReviewed:   time.Now().AddDate(0, 0, -10),
				ReviewInterval: 10,
			},
			rating:               2,
			expectedNextInterval: 1, // Reset to 1 day
		},
		{
			name: "established card (interval 10) with rating 4 (good)",
			card: domain.Card{
				LastReviewed:   time.Now().AddDate(0, 0, -10),
				ReviewInterval: 10,
			},
			rating:               4,
			expectedNextInterval: 11, // Increment by at least 1
		},
		{
			name: "established card (interval 10) with rating 5 (easy)",
			card: domain.Card{
				LastReviewed:   time.Now().AddDate(0, 0, -10),
				ReviewInterval: 10,
			},
			rating:               5,
			expectedNextInterval: 13, // Should apply easy bonus
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sm2.CalculateNextInterval(tc.card, tc.rating)
			if result != tc.expectedNextInterval {
				t.Errorf("expected next interval %d, got %d", tc.expectedNextInterval, result)
			}
		})
	}
}

func TestIsDue(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test cases
	testCases := []struct {
		name     string
		card     domain.Card
		expected bool
	}{
		{
			name: "new card never reviewed",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 1,
			},
			expected: true,
		},
		{
			name: "card reviewed today with 1 day interval",
			card: domain.Card{
				LastReviewed:   time.Now(),
				ReviewInterval: 1,
			},
			expected: false,
		},
		{
			name: "card reviewed 2 days ago with 1 day interval",
			card: domain.Card{
				LastReviewed:   time.Now().AddDate(0, 0, -2),
				ReviewInterval: 1,
			},
			expected: true,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sm2.IsDue(tc.card)
			if result != tc.expected {
				t.Errorf("expected IsDue() to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestCalculateEaseFactor(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Test cases
	testCases := []struct {
		difficulty     int
		expectedFactor float64
	}{
		{0, 2.5},
		{1, 2.26},
		{2, 2.02},
		{3, 1.78},
		{4, 1.54},
		{5, 1.3},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("difficulty_%d", tc.difficulty), func(t *testing.T) {
			result := sm2.CalculateEaseFactor(tc.difficulty)
			// Allow small floating point differences
			if result < tc.expectedFactor-0.01 || result > tc.expectedFactor+0.01 {
				t.Errorf("expected ease factor %f for difficulty %d, got %f",
					tc.expectedFactor, tc.difficulty, result)
			}
		})
	}
}

// Test for GetDueDate
func TestGetDueDate(t *testing.T) {
	sm2 := NewSM2Algorithm()

	// Create test cases with a fixed reference time (midnight)
	referenceTime := time.Date(2025, 3, 25, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		card     domain.Card
		expected time.Time
	}{
		{
			name: "new card never reviewed",
			card: domain.Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 1,
			},
			// Special case: for never-reviewed cards, expect "now" - we'll test differently
		},
		{
			name: "card with 1 day interval",
			card: domain.Card{
				LastReviewed:   referenceTime,
				ReviewInterval: 1,
			},
			expected: time.Date(2025, 3, 26, 0, 0, 0, 0, time.UTC), // 1 day later, midnight
		},
		{
			name: "card with 7 day interval",
			card: domain.Card{
				LastReviewed:   referenceTime,
				ReviewInterval: 7,
			},
			expected: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), // 7 days later, midnight
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sm2.GetDueDate(tc.card)

			if tc.card.LastReviewed.IsZero() {
				// For never-reviewed cards, verify result is close to current time
				timeDiff := time.Since(result)
				if timeDiff > time.Second*2 || timeDiff < -time.Second*2 {
					t.Errorf("expected time close to now, got difference of %v", timeDiff)
				}
				return
			}

			// For normal cases, compare exact times
			if !result.Equal(tc.expected) {
				t.Errorf("expected due date %v, got %v", tc.expected, result)
			}
		})
	}
}
