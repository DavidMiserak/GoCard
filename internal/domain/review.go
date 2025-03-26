// internal/domain/review.go
package domain

import (
	"errors"
	"time"
)

// ReviewSession represents an active review of cards
type ReviewSession struct {
	DeckPath    string    // Path of the deck being reviewed
	StartTime   time.Time // Time the session started
	CardPaths   []string  // File paths of cards in the session
	CurrentCard int       // Index of current card
	Completed   []bool    // Tracks which cards are completed
	Ratings     []int     // Rating given to each card
}

// ReviewSessionSummary contains statistics about a completed review session
type ReviewSessionSummary struct {
	DeckPath      string
	Duration      time.Duration
	CardsReviewed int
	AverageRating float64
	NewCards      int
	ReviewedCards int
}

// NewReviewSession creates a new review session for a deck with the given card paths
func NewReviewSession(deckPath string, cardPaths []string) *ReviewSession {
	completed := make([]bool, len(cardPaths))
	ratings := make([]int, len(cardPaths))

	return &ReviewSession{
		DeckPath:    deckPath,
		StartTime:   time.Now(),
		CardPaths:   cardPaths,
		CurrentCard: 0,
		Completed:   completed,
		Ratings:     ratings,
	}
}

// GetCurrentCardPath returns the path of the current card
func (rs *ReviewSession) GetCurrentCardPath() (string, error) {
	if rs.CurrentCard >= len(rs.CardPaths) {
		return "", errors.New("no more cards in the session")
	}
	return rs.CardPaths[rs.CurrentCard], nil
}

// SubmitRating records a rating for the current card and moves to the next
func (rs *ReviewSession) SubmitRating(rating int) error {
	if rs.CurrentCard >= len(rs.CardPaths) {
		return errors.New("no more cards in the session")
	}

	rs.Ratings[rs.CurrentCard] = rating
	rs.Completed[rs.CurrentCard] = true
	rs.CurrentCard++

	return nil
}

// IsComplete checks if all cards in the session have been reviewed
func (rs *ReviewSession) IsComplete() bool {
	// Check if all card paths have been processed
	return rs.CurrentCard >= len(rs.CardPaths)
}

// GenerateSummary creates a summary of the review session
func (rs *ReviewSession) GenerateSummary() ReviewSessionSummary {
	completedCount := 0
	totalRating := 0
	newCards := 0

	// Count completed cards by checking the Completed slice
	for i, completed := range rs.Completed {
		if completed {
			completedCount++
			totalRating += rs.Ratings[i]
		}
	}

	avgRating := 0.0
	if completedCount > 0 {
		avgRating = float64(totalRating) / float64(completedCount)
	}

	return ReviewSessionSummary{
		DeckPath:      rs.DeckPath,
		Duration:      time.Since(rs.StartTime),
		CardsReviewed: completedCount,
		AverageRating: avgRating,
		NewCards:      newCards,
		ReviewedCards: completedCount - newCards,
	}
}
