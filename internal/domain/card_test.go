// internal/domain/card_test.go
package domain

import (
	"testing"
	"time"
)

func TestCardIsDue(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		card     Card
		expected bool
	}{
		{
			name: "new card never reviewed",
			card: Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 1,
			},
			expected: true,
		},
		{
			name: "card reviewed today with 1 day interval",
			card: Card{
				LastReviewed:   time.Now(),
				ReviewInterval: 1,
			},
			expected: false,
		},
		{
			name: "card reviewed 2 days ago with 1 day interval",
			card: Card{
				LastReviewed:   time.Now().AddDate(0, 0, -2),
				ReviewInterval: 1,
			},
			expected: true,
		},
		{
			name: "card reviewed 5 days ago with 10 day interval",
			card: Card{
				LastReviewed:   time.Now().AddDate(0, 0, -5),
				ReviewInterval: 10,
			},
			expected: false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.card.IsDue()
			if result != tc.expected {
				t.Errorf("expected IsDue() to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestCardGetDueDate(t *testing.T) {
	// Test cases
	now := time.Now()
	testCases := []struct {
		name     string
		card     Card
		expected time.Time
	}{
		{
			name: "new card never reviewed",
			card: Card{
				LastReviewed:   time.Time{}, // zero value
				ReviewInterval: 1,
			},
			expected: now,
		},
		{
			name: "card with 1 day interval",
			card: Card{
				LastReviewed:   now.AddDate(0, 0, -1),
				ReviewInterval: 1,
			},
			expected: now.AddDate(0, 0, -1).AddDate(0, 0, 1),
		},
		{
			name: "card with 10 day interval",
			card: Card{
				LastReviewed:   now.AddDate(0, 0, -5),
				ReviewInterval: 10,
			},
			expected: now.AddDate(0, 0, -5).AddDate(0, 0, 10),
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.card.GetDueDate()

			// For the zero value case, we can only approximate
			if tc.card.LastReviewed.IsZero() {
				// Check that it's close to now (within 1 second)
				diff := result.Sub(now)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("expected due date close to now, got %v (diff: %v)", result, diff)
				}
				return
			}

			// Normal case comparison
			if !result.Equal(tc.expected) {
				t.Errorf("expected due date %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestNewCard(t *testing.T) {
	filePath := "/path/to/card.md"
	card := NewCard(filePath)

	if card.FilePath != filePath {
		t.Errorf("expected FilePath to be %s, got %s", filePath, card.FilePath)
	}

	if card.ReviewInterval != 1 {
		t.Errorf("expected default ReviewInterval to be 1, got %d", card.ReviewInterval)
	}

	if card.Difficulty != 3 {
		t.Errorf("expected default Difficulty to be 3, got %d", card.Difficulty)
	}

	if len(card.Tags) != 0 {
		t.Errorf("expected Tags to be empty, got %v", card.Tags)
	}

	if card.Frontmatter == nil {
		t.Errorf("expected Frontmatter to be initialized")
	}
}
