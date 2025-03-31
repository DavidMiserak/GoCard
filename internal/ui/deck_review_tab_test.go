// File: internal/ui/deck_review_tab_test.go

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

// Deck Review Tab Tests
func TestGetMatureCards(t *testing.T) {
	// Create a store with known card intervals
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:       "card-1",
						Interval: 10, // Not mature (< 21)
					},
					{
						ID:       "card-2",
						Interval: 21, // Mature (>= 21)
					},
					{
						ID:       "card-3",
						Interval: 30, // Mature
					},
				},
			},
		},
	}

	// Expected: 2 mature cards
	expectedCount := 2
	actualCount := getMatureCards(testStore)

	if actualCount != expectedCount {
		t.Errorf("Expected mature cards to be %d, got %d", expectedCount, actualCount)
	}
}

func TestCalculateSuccessRate(t *testing.T) {
	// Create a store with known ratings
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:           "card-1",
						LastReviewed: time.Now(),
						Rating:       5, // Success (rating >= 3)
					},
					{
						ID:           "card-2",
						LastReviewed: time.Now(),
						Rating:       3, // Success
					},
					{
						ID:           "card-3",
						LastReviewed: time.Now(),
						Rating:       2, // Failure
					},
					{
						ID:           "card-4",
						LastReviewed: time.Now(),
						Rating:       1, // Failure
					},
				},
			},
		},
	}

	// Expected success rate: 2 successful out of 4 = 50%
	expectedRate := 50
	actualRate := calculateSuccessRate(testStore)

	if actualRate != expectedRate {
		t.Errorf("Expected success rate to be %d%%, got %d%%", expectedRate, actualRate)
	}
}

func TestCalculateAverageInterval(t *testing.T) {
	// Create a store with known intervals
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:           "card-1",
						LastReviewed: time.Now(),
						Interval:     10,
					},
					{
						ID:           "card-2",
						LastReviewed: time.Now(),
						Interval:     20,
					},
				},
			},
		},
	}

	// Expected average: (10 + 20) / 2 = 15
	expectedAvg := 15.0
	actualAvg := calculateAverageInterval(testStore)

	if actualAvg != expectedAvg {
		t.Errorf("Expected average interval to be %.1f, got %.1f", expectedAvg, actualAvg)
	}
}

func TestFormatLastStudied(t *testing.T) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "Zero time",
			date:     time.Time{},
			expected: "Never",
		},
		{
			name:     "Today",
			date:     today.Add(2 * time.Hour), // Some time today
			expected: "Today",
		},
		{
			name:     "Yesterday",
			date:     yesterday.Add(2 * time.Hour), // Some time yesterday
			expected: "Yesterday",
		},
		{
			name:     "Earlier date",
			date:     twoDaysAgo,
			expected: twoDaysAgo.Format("Jan 2"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatLastStudied(test.date)
			if result != test.expected {
				t.Errorf("Expected formatLastStudied(%v) to be '%s', got '%s'", test.date, test.expected, result)
			}
		})
	}
}

func TestCalculateRatingDistribution(t *testing.T) {
	// Create a store with known ratings
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:           "card-1",
						LastReviewed: time.Now(),
						Rating:       1,
					},
					{
						ID:           "card-2",
						LastReviewed: time.Now(),
						Rating:       2,
					},
					{
						ID:           "card-3",
						LastReviewed: time.Now(),
						Rating:       2, // Another 2
					},
					{
						ID:           "card-4",
						LastReviewed: time.Now(),
						Rating:       3,
					},
					{
						ID:           "card-5",
						LastReviewed: time.Now(),
						Rating:       4,
					},
				},
			},
		},
	}

	distribution := calculateRatingDistribution(testStore)

	// Check each rating count
	expectedDistribution := map[int]int{
		1: 1, // One card with rating 1
		2: 2, // Two cards with rating 2
		3: 1, // One card with rating 3
		4: 1, // One card with rating 4
		5: 0, // No cards with rating 5
	}

	for rating, expectedCount := range expectedDistribution {
		if distribution[rating] != expectedCount {
			t.Errorf("Expected rating %d to have count %d, got %d", rating, expectedCount, distribution[rating])
		}
	}
}

func TestRenderDeckReviewStats(t *testing.T) {
	store := createTestStoreForDeckReview()

	// Just test that rendering doesn't panic and returns a non-empty string
	result := renderDeckReviewStats(store)

	if result == "" {
		t.Error("Expected renderDeckReviewStats to return a non-empty string")
	}

	// Check if the result contains expected headers
	expectedHeaders := []string{
		"Total Cards:",
		"Mature Cards:",
		"New Cards:",
		"Success Rate:",
		"Avg. Interval:",
		"Last Studied:",
		"Ratings Distribution",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Expected output to contain '%s'", header)
		}
	}
}

// Helper function to create a test store
func createTestStoreForDeckReview() *data.Store {
	return data.NewStore()
}
