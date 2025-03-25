// internal/service/interfaces/review_service.go
package interfaces

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
)

// ReviewSessionSummary contains statistics about a completed review session
type ReviewSessionSummary struct {
	DeckPath      string
	Duration      time.Duration
	CardsReviewed int
	AverageRating float64
	NewCards      int
	ReviewedCards int
}

// ReviewService manages the review process
type ReviewService interface {
	// Session management
	StartSession(deckPath string) (domain.ReviewSession, error)
	GetSession() (domain.ReviewSession, error)
	EndSession() (ReviewSessionSummary, error)

	// Card review operations
	GetNextCard() (domain.Card, error)
	SubmitRating(rating int) error
	GetSessionStats() (map[string]interface{}, error)
}
