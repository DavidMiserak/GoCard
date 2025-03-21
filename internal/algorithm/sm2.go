// File: internal/algorithm/sm2.go
package algorithm

import (
	"math"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// SM2Algorithm implements the SuperMemo 2 spaced repetition algorithm
type SM2Algorithm struct {
	// EasyBonus is a multiplier for intervals when a card is rated as easy (4-5)
	EasyBonus float64
	// IntervalModifier is a global scaling factor for all intervals
	IntervalModifier float64
}

// Create global SM2 instance
var SM2 = NewSM2Algorithm()

// NewSM2Algorithm creates a new SM2Algorithm with default parameters
func NewSM2Algorithm() *SM2Algorithm {
	return &SM2Algorithm{
		EasyBonus:        1.3,
		IntervalModifier: 1.0,
	}
}

// CalculateNextReview applies the SM2 algorithm to calculate the next review date and interval
// based on the user's performance rating (0-5) and the card's current state.
// Returns the new interval in days.
func (sm2 *SM2Algorithm) CalculateNextReview(cardObj *card.Card, rating int) int {
	// Ensure rating is within bounds
	if rating < 0 {
		rating = 0
	} else if rating > 5 {
		rating = 5
	}

	// Update card difficulty based on rating
	cardObj.Difficulty = rating

	var interval int

	// Calculate the interval based on the SM-2 algorithm
	if rating < 3 {
		// If rating is less than 3, reset to learning phase
		interval = 1
	} else {
		// If this is the first successful review
		if cardObj.LastReviewed.IsZero() || cardObj.ReviewInterval == 0 {
			if rating == 3 {
				interval = 1
			} else if rating == 4 {
				interval = 3
			} else { // rating == 5
				interval = 5
			}
		} else {
			// Calculate new interval for cards with previous successful reviews
			interval = cardObj.ReviewInterval

			// Apply SM-2 formula for increasing intervals
			if rating == 3 {
				// Hard but correct response - small increase
				interval = int(float64(interval) * 1.2)
			} else if rating == 4 {
				// Good response - standard increase
				interval = int(float64(interval) * 1.8)
			} else { // rating == 5
				// Easy response - apply easy bonus
				interval = int(float64(interval) * 2.5 * sm2.EasyBonus)
			}
		}

		// Apply global interval modifier
		interval = int(float64(interval) * sm2.IntervalModifier)

		// Ensure minimum interval of 1 day
		if interval < 1 {
			interval = 1
		}
	}

	// Update card state
	cardObj.LastReviewed = time.Now()
	cardObj.ReviewInterval = interval

	return interval
}

// CalculateDueDate returns the next due date for a card based on its last review and interval
func (sm2 *SM2Algorithm) CalculateDueDate(cardObj *card.Card) time.Time {
	// If the card has never been reviewed, it's due now
	if cardObj.LastReviewed.IsZero() {
		return time.Now()
	}

	// Calculate the next due date based on last review + interval
	return cardObj.LastReviewed.AddDate(0, 0, cardObj.ReviewInterval)
}

// CalculatePercentOverdue returns how overdue a card is as a percentage
// 0% = due exactly now, 100% = one full interval overdue
func (sm2 *SM2Algorithm) CalculatePercentOverdue(cardObj *card.Card) float64 {
	// If the card has never been reviewed, it's fully overdue
	if cardObj.LastReviewed.IsZero() {
		return 100.0
	}

	// If the card is not due yet, return 0
	dueDate := sm2.CalculateDueDate(cardObj)
	if time.Now().Before(dueDate) {
		return 0.0
	}

	// Calculate overdue percentage
	intervalDuration := time.Duration(cardObj.ReviewInterval) * 24 * time.Hour
	overdueDuration := time.Since(dueDate)
	percentOverdue := (float64(overdueDuration) / float64(intervalDuration)) * 100.0

	// Cap at 100% for very overdue cards
	return math.Min(percentOverdue, 100.0)
}

// EstimateEase returns an estimated ease factor (how easy the card is to remember)
// based on review history and ratings
func (sm2 *SM2Algorithm) EstimateEase(cardObj *card.Card) float64 {
	// Simple implementation: scale from difficulty (0-5)
	// 0 = most difficult (ease = 1.3), 5 = easiest (ease = 3.0)
	baseEase := 1.3
	difficultyFactor := float64(cardObj.Difficulty) / 5.0
	easeRange := 1.7 // 3.0 - 1.3

	return baseEase + (difficultyFactor * easeRange)
}

// IsDue returns true if a card is due for review
func (sm2 *SM2Algorithm) IsDue(cardObj *card.Card) bool {
	// If the card has never been reviewed, it's due now
	if cardObj.LastReviewed.IsZero() {
		return true
	}

	// Check if the current time is after the due date
	dueDate := sm2.CalculateDueDate(cardObj)
	return time.Now().After(dueDate) || time.Now().Equal(dueDate)
}
