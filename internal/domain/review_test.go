// internal/domain/review_test.go
package domain

import (
	"testing"
	"time"
)

func TestNewReviewSession(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		deckPath  string
		cardPaths []string
		expectLen int
	}{
		{
			name:      "empty session",
			deckPath:  "/test/cards/deck1",
			cardPaths: []string{},
			expectLen: 0,
		},
		{
			name:      "session with cards",
			deckPath:  "/test/cards/deck1",
			cardPaths: []string{"card1.md", "card2.md", "card3.md"},
			expectLen: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new review session
			session := NewReviewSession(tc.deckPath, tc.cardPaths)

			// Check deck path
			if session.DeckPath != tc.deckPath {
				t.Errorf("expected DeckPath to be %s, got %s", tc.deckPath, session.DeckPath)
			}

			// Check card paths
			if len(session.CardPaths) != tc.expectLen {
				t.Errorf("expected %d card paths, got %d", tc.expectLen, len(session.CardPaths))
			}

			// Check completed and ratings arrays
			if len(session.Completed) != tc.expectLen {
				t.Errorf("expected %d completed flags, got %d", tc.expectLen, len(session.Completed))
			}
			if len(session.Ratings) != tc.expectLen {
				t.Errorf("expected %d ratings, got %d", tc.expectLen, len(session.Ratings))
			}

			// Check current card index
			if session.CurrentCard != 0 {
				t.Errorf("expected CurrentCard to be 0, got %d", session.CurrentCard)
			}

			// Check start time is recent
			timeDiff := time.Since(session.StartTime)
			if timeDiff > 5*time.Second {
				t.Errorf("expected StartTime to be recent, got %v ago", timeDiff)
			}
		})
	}
}

func TestGetCurrentCardPath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name         string
		cardPaths    []string
		currentCard  int
		expectError  bool
		expectedPath string
	}{
		{
			name:         "first card",
			cardPaths:    []string{"card1.md", "card2.md", "card3.md"},
			currentCard:  0,
			expectError:  false,
			expectedPath: "card1.md",
		},
		{
			name:         "middle card",
			cardPaths:    []string{"card1.md", "card2.md", "card3.md"},
			currentCard:  1,
			expectError:  false,
			expectedPath: "card2.md",
		},
		{
			name:         "out of bounds",
			cardPaths:    []string{"card1.md", "card2.md"},
			currentCard:  2,
			expectError:  true,
			expectedPath: "",
		},
		{
			name:         "empty session",
			cardPaths:    []string{},
			currentCard:  0,
			expectError:  true,
			expectedPath: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a review session
			session := NewReviewSession("/test/deck", tc.cardPaths)
			session.CurrentCard = tc.currentCard

			// Get the current card path
			path, err := session.GetCurrentCardPath()

			// Check error
			if tc.expectError && err == nil {
				t.Error("expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("did not expect an error but got: %v", err)
			}

			// Check path if no error
			if !tc.expectError {
				if path != tc.expectedPath {
					t.Errorf("expected path %s, got %s", tc.expectedPath, path)
				}
			}
		})
	}
}

func TestSubmitRating(t *testing.T) {
	// Test cases
	testCases := []struct {
		name         string
		cardPaths    []string
		currentCard  int
		rating       int
		expectError  bool
		expectedNext int
	}{
		{
			name:         "submit first rating",
			cardPaths:    []string{"card1.md", "card2.md", "card3.md"},
			currentCard:  0,
			rating:       4,
			expectError:  false,
			expectedNext: 1,
		},
		{
			name:         "submit last rating",
			cardPaths:    []string{"card1.md", "card2.md"},
			currentCard:  1,
			rating:       5,
			expectError:  false,
			expectedNext: 2,
		},
		{
			name:         "out of bounds",
			cardPaths:    []string{"card1.md", "card2.md"},
			currentCard:  2,
			rating:       3,
			expectError:  true,
			expectedNext: 2, // Should not change
		},
		{
			name:         "empty session",
			cardPaths:    []string{},
			currentCard:  0,
			rating:       4,
			expectError:  true,
			expectedNext: 0, // Should not change
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a review session
			session := NewReviewSession("/test/deck", tc.cardPaths)
			session.CurrentCard = tc.currentCard

			// Submit the rating
			err := session.SubmitRating(tc.rating)

			// Check error
			if tc.expectError && err == nil {
				t.Error("expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("did not expect an error but got: %v", err)
			}

			// Check that current card index was updated correctly
			if session.CurrentCard != tc.expectedNext {
				t.Errorf("expected CurrentCard to be %d, got %d", tc.expectedNext, session.CurrentCard)
			}

			// If no error, check that the rating was recorded
			if !tc.expectError {
				if session.Ratings[tc.currentCard] != tc.rating {
					t.Errorf("expected rating %d to be recorded, got %d", tc.rating, session.Ratings[tc.currentCard])
				}
				if !session.Completed[tc.currentCard] {
					t.Error("expected card to be marked as completed")
				}
			}
		})
	}
}

func TestIsComplete(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		cardPaths   []string
		currentCard int
		expected    bool
	}{
		{
			name:        "not started",
			cardPaths:   []string{"card1.md", "card2.md", "card3.md"},
			currentCard: 0,
			expected:    false,
		},
		{
			name:        "in progress",
			cardPaths:   []string{"card1.md", "card2.md", "card3.md"},
			currentCard: 1,
			expected:    false,
		},
		{
			name:        "complete",
			cardPaths:   []string{"card1.md", "card2.md", "card3.md"},
			currentCard: 3,
			expected:    true,
		},
		{
			name:        "empty session",
			cardPaths:   []string{},
			currentCard: 0,
			expected:    true, // Edge case: empty session is considered complete
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a review session
			session := NewReviewSession("/test/deck", tc.cardPaths)
			session.CurrentCard = tc.currentCard

			// Check if complete
			result := session.IsComplete()

			if result != tc.expected {
				t.Errorf("expected IsComplete() to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	// Create a session with some completed cards and ratings
	session := NewReviewSession("/test/deck", []string{"card1.md", "card2.md", "card3.md"})

	// Simulate submitting ratings
	_ = session.SubmitRating(5) // Card 1: 5
	_ = session.SubmitRating(3) // Card 2: 3
	// Card 3 is not rated

	// Get summary
	summary := session.GenerateSummary()

	// Check summary properties
	if summary.DeckPath != "/test/deck" {
		t.Errorf("expected DeckPath to be /test/deck, got %s", summary.DeckPath)
	}

	if summary.CardsReviewed != 2 {
		t.Errorf("expected CardsReviewed to be 2, got %d", summary.CardsReviewed)
	}

	// Average rating should be (5+3)/2 = 4.0
	expectedAvg := 4.0
	if summary.AverageRating != expectedAvg {
		t.Errorf("expected AverageRating to be %.1f, got %.1f", expectedAvg, summary.AverageRating)
	}

	// Duration should be reasonable (less than a minute)
	if summary.Duration > time.Minute {
		t.Errorf("expected Duration to be reasonable, got %v", summary.Duration)
	}
}
