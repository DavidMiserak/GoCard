// internal/service/review/review_service.go
package review

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// DefaultReviewService implements the ReviewService interface
type DefaultReviewService struct {
	storage   interfaces.StorageService
	cardSvc   interfaces.CardService
	deckSvc   interfaces.DeckService
	algorithm *algorithm.SM2Algorithm
	session   *domain.ReviewSession
}

// NewReviewService creates a new review service
func NewReviewService(
	storage interfaces.StorageService,
	cardSvc interfaces.CardService,
	deckSvc interfaces.DeckService,
	algorithm *algorithm.SM2Algorithm,
) interfaces.ReviewService {
	return &DefaultReviewService{
		storage:   storage,
		cardSvc:   cardSvc,
		deckSvc:   deckSvc,
		algorithm: algorithm,
		session:   nil,
	}
}

// StartSession begins a new review session for a deck
func (rs *DefaultReviewService) StartSession(deckPath string) (domain.ReviewSession, error) {
	if rs.session != nil {
		return *rs.session, errors.New("review session already in progress")
	}

	// Get all due cards for the deck
	dueCards, err := rs.deckSvc.GetDueCards(deckPath)
	if err != nil {
		return domain.ReviewSession{}, fmt.Errorf("failed to get due cards: %w", err)
	}

	// Extract card paths
	var cardPaths []string
	for _, card := range dueCards {
		cardPaths = append(cardPaths, card.FilePath)
	}

	// Shuffle the cards
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(cardPaths), func(i, j int) {
		cardPaths[i], cardPaths[j] = cardPaths[j], cardPaths[i]
	})

	// Create a new session
	rs.session = domain.NewReviewSession(deckPath, cardPaths)

	return *rs.session, nil
}

// GetSession returns the current review session
func (rs *DefaultReviewService) GetSession() (domain.ReviewSession, error) {
	if rs.session == nil {
		return domain.ReviewSession{}, errors.New("no active review session")
	}
	return *rs.session, nil
}

// EndSession ends the current review session and returns a summary
func (rs *DefaultReviewService) EndSession() (interfaces.ReviewSessionSummary, error) {
	if rs.session == nil {
		return interfaces.ReviewSessionSummary{}, errors.New("no active review session")
	}

	// Generate summary
	summary := rs.session.GenerateSummary()

	// Create the interface summary object
	result := interfaces.ReviewSessionSummary{
		DeckPath:      summary.DeckPath,
		Duration:      summary.Duration,
		CardsReviewed: summary.CardsReviewed,
		AverageRating: summary.AverageRating,
		NewCards:      summary.NewCards,
		ReviewedCards: summary.ReviewedCards,
	}

	// Clear the session
	rs.session = nil

	return result, nil
}

// GetNextCard returns the next card in the session
func (rs *DefaultReviewService) GetNextCard() (domain.Card, error) {
	if rs.session == nil {
		return domain.Card{}, errors.New("no active review session")
	}

	if rs.session.IsComplete() {
		return domain.Card{}, errors.New("review session is complete")
	}

	cardPath, err := rs.session.GetCurrentCardPath()
	if err != nil {
		return domain.Card{}, err
	}

	return rs.cardSvc.GetCard(cardPath)
}

// SubmitRating submits a rating for the current card and advances to the next
func (rs *DefaultReviewService) SubmitRating(rating int) error {
	if rs.session == nil {
		return errors.New("no active review session")
	}

	if rs.session.IsComplete() {
		return errors.New("review session is complete")
	}

	// Get the current card path
	cardPath, err := rs.session.GetCurrentCardPath()
	if err != nil {
		return err
	}

	// Update the card with the rating
	if err := rs.cardSvc.ReviewCard(cardPath, rating); err != nil {
		return fmt.Errorf("failed to update card review data: %w", err)
	}

	// Advance the session to the next card
	return rs.session.SubmitRating(rating)
}

// GetSessionStats returns statistics about the current session
func (rs *DefaultReviewService) GetSessionStats() (map[string]interface{}, error) {
	if rs.session == nil {
		return nil, errors.New("no active review session")
	}

	// Count completed and remaining cards
	completed := 0
	for _, isComplete := range rs.session.Completed {
		if isComplete {
			completed++
		}
	}

	remaining := len(rs.session.CardPaths) - completed

	// Calculate average rating for completed cards
	totalRating := 0
	count := 0
	for i, isComplete := range rs.session.Completed {
		if isComplete {
			totalRating += rs.session.Ratings[i]
			count++
		}
	}

	avgRating := 0.0
	if count > 0 {
		avgRating = float64(totalRating) / float64(count)
	}

	// Current progress
	progress := 0.0
	if len(rs.session.CardPaths) > 0 {
		progress = float64(completed) / float64(len(rs.session.CardPaths)) * 100.0
	}

	// Create stats map
	stats := map[string]interface{}{
		"total_cards":     len(rs.session.CardPaths),
		"completed_cards": completed,
		"remaining_cards": remaining,
		"average_rating":  avgRating,
		"progress":        progress,
		"start_time":      rs.session.StartTime,
		"duration":        time.Since(rs.session.StartTime),
	}

	return stats, nil
}

// Ensure DefaultReviewService implements ReviewService
var _ interfaces.ReviewService = (*DefaultReviewService)(nil)
