// File: internal/srs/algorithm.go

package srs

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

// Default values for SM-2 algorithm
const (
	defaultEase     = 2.5  // Initial ease factor
	minEase         = 1.3  // Minimum ease factor
	easeModifier    = 0.15 // How much ease changes based on rating
	maxInterval     = 365  // Maximum interval in days
	easyBonus       = 1.3  // Multiplier for "easy" cards
	defaultInterval = 1    // Default interval for new cards
)

// ScheduleCard updates a card based on the user's rating (1-5)
// and returns the updated card
//
// Rating scale:
// 1 - Blackout (complete failure)
// 2 - Wrong (significant difficulty)
// 3 - Hard (correct with difficulty)
// 4 - Good (correct with some effort)
// 5 - Easy (correct with no effort)
func ScheduleCard(card model.Card, rating int) model.Card {
	// Update the last reviewed time
	card.LastReviewed = time.Now()

	// Store the user's rating
	card.Rating = rating

	// Calculate new interval and ease based on rating
	switch rating {
	case 1: // Blackout
		// Reset the interval, reduce ease
		card.Interval = 1
		card.Ease = maxFloat(card.Ease-0.3, minEase)

	case 2: // Wrong
		// Reset the interval, reduce ease
		card.Interval = 1
		card.Ease = maxFloat(card.Ease-0.2, minEase)

	case 3: // Hard
		// Slight increase in interval, reduce ease
		if card.Interval == 0 {
			card.Interval = 1
		} else {
			card.Interval = int(float64(card.Interval) * 1.2)
		}
		card.Ease = maxFloat(card.Ease-easeModifier, minEase)

	case 4: // Good
		// Standard increase in interval
		switch card.Interval {
		case 0:
			card.Interval = defaultInterval
		case 1:
			card.Interval = 3
		default:
			card.Interval = int(float64(card.Interval) * card.Ease)
		}
		// Ease remains the same

	case 5: // Easy
		// Larger increase in interval, increase ease
		switch card.Interval {
		case 0:
			card.Interval = defaultInterval * 2
		case 1:
			card.Interval = 4
		default:
			card.Interval = int(float64(card.Interval) * card.Ease * easyBonus)
		}
		card.Ease = minFloat(card.Ease+easeModifier, 4.0)
	}

	// Cap the interval at the maximum
	card.Interval = minInt(card.Interval, maxInterval)

	// Set the next review date
	card.NextReview = time.Now().AddDate(0, 0, card.Interval)

	return card
}

// InitializeNewCard initializes a new card with default SRS values
func InitializeNewCard(card model.Card) model.Card {
	// Set default values for a new card
	if card.Ease == 0 {
		card.Ease = defaultEase
	}
	card.Interval = 0
	card.NextReview = time.Now() // Due immediately

	return card
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
