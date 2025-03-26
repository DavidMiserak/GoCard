// internal/service/card/card_service.go
package card

import (
	"fmt"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// DefaultCardService implements the CardService interface
type DefaultCardService struct {
	storage   interfaces.StorageService
	algorithm *algorithm.SM2Algorithm
}

// NewCardService creates a new card service
func NewCardService(storage interfaces.StorageService, algorithm *algorithm.SM2Algorithm) interfaces.CardService {
	return &DefaultCardService{
		storage:   storage,
		algorithm: algorithm,
	}
}

// GetCard retrieves a card by path
func (cs *DefaultCardService) GetCard(cardPath string) (domain.Card, error) {
	return cs.storage.LoadCard(cardPath)
}

// ReviewCard updates a card's review metadata based on the rating
func (cs *DefaultCardService) ReviewCard(cardPath string, rating int) error {
	// Load the card
	card, err := cs.storage.LoadCard(cardPath)
	if err != nil {
		return fmt.Errorf("failed to load card for review: %w", err)
	}

	// Calculate the next interval
	nextInterval := cs.algorithm.CalculateNextInterval(card, rating)

	// Update card metadata
	card.LastReviewed = time.Now()
	card.ReviewInterval = nextInterval

	// Update difficulty based on rating (optional)
	if rating >= 0 && rating <= 5 {
		// We can optionally update the difficulty based on the rating
		// Lower ratings mean higher difficulty
		card.Difficulty = 5 - rating
		if card.Difficulty < 0 {
			card.Difficulty = 0
		}
	}

	// Save the updated card
	return cs.storage.UpdateCardMetadata(card)
}

// IsDue checks if a card is due for review
func (cs *DefaultCardService) IsDue(cardPath string) bool {
	card, err := cs.storage.LoadCard(cardPath)
	if err != nil {
		return false
	}

	// Direct implementation instead of delegating to algorithm
	// This ensures consistency with our fixes
	if card.LastReviewed.IsZero() {
		// Card has never been reviewed - it's always due
		return true
	}

	// Calculate due date and compare with today
	dueDate := card.LastReviewed.AddDate(0, 0, card.ReviewInterval)
	now := time.Now()

	// Compare dates only (ignoring time components)
	dueYear, dueMonth, dueDay := dueDate.Date()
	nowYear, nowMonth, nowDay := now.Date()

	dueDate = time.Date(dueYear, dueMonth, dueDay, 0, 0, 0, 0, dueDate.Location())
	now = time.Date(nowYear, nowMonth, nowDay, 0, 0, 0, 0, now.Location())

	// Card is due if today is on or after the due date
	return now.After(dueDate) || now.Equal(dueDate)
}

// GetDueDate returns the next due date for a card
func (cs *DefaultCardService) GetDueDate(cardPath string) time.Time {
	card, err := cs.storage.LoadCard(cardPath)
	if err != nil {
		return time.Time{} // Return zero time if card can't be loaded
	}

	if card.LastReviewed.IsZero() {
		return time.Now() // For never-reviewed cards, return current time
	}

	// Calculate due date directly from the card's data
	return card.LastReviewed.AddDate(0, 0, card.ReviewInterval)
}

// Ensure DefaultCardService implements CardService
var _ interfaces.CardService = (*DefaultCardService)(nil)
