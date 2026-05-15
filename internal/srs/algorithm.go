// File: internal/srs/algorithm.go

package srs

import (
	"fmt"
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

// ScheduleResult contains the output of a scheduling operation
type ScheduleResult struct {
	NextReviewDate time.Time // Next time card should be reviewed
	Interval       int       // Interval in days until next review
	EaseOrRetention float64  // SM2: Ease (1.3-4.0), FSRS: Retention (0-1)
	LastReview     time.Time // When the card was last reviewed
}

// Algorithm interface defines the contract for different SRS algorithms
type Algorithm interface {
	// Schedule computes next review date and updates card state
	// Returns ScheduleResult with next review date and updated parameters
	Schedule(card *model.Card, rating int) ScheduleResult
}

// SM2Scheduler implements the Algorithm interface for SM-2
type SM2Scheduler struct{}

// NewSM2Scheduler creates a new SM2Scheduler
func NewSM2Scheduler() *SM2Scheduler {
	return &SM2Scheduler{}
}

// Schedule implements the Algorithm interface for SM-2
func (s *SM2Scheduler) Schedule(card *model.Card, rating int) ScheduleResult {
	now := time.Now()
	card.LastReviewed = now
	card.Rating = rating

	// Calculate new interval and ease based on rating
	switch rating {
	case 1: // Blackout
		card.Interval = 1
		card.Ease = maxFloat(card.Ease-0.3, minEase)

	case 2: // Wrong
		card.Interval = 1
		card.Ease = maxFloat(card.Ease-0.2, minEase)

	case 3: // Hard
		if card.Interval == 0 {
			card.Interval = 1
		} else {
			card.Interval = int(float64(card.Interval) * 1.2)
		}
		card.Ease = maxFloat(card.Ease-easeModifier, minEase)

	case 4: // Good
		switch card.Interval {
		case 0:
			card.Interval = defaultInterval
		case 1:
			card.Interval = 3
		default:
			card.Interval = int(float64(card.Interval) * card.Ease)
		}

	case 5: // Easy
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

	card.Interval = minInt(card.Interval, maxInterval)
	card.NextReview = now.AddDate(0, 0, card.Interval)

	return ScheduleResult{
		NextReviewDate:  card.NextReview,
		Interval:        card.Interval,
		EaseOrRetention: card.Ease,
		LastReview:      card.LastReviewed,
	}
}

// FSRSScheduler implements the Algorithm interface for FSRS (placeholder)
type FSRSScheduler struct{}

// NewFSRSScheduler creates a new FSRSScheduler
func NewFSRSScheduler() *FSRSScheduler {
	return &FSRSScheduler{}
}

// Schedule implements the Algorithm interface for FSRS
// MVP uses approximate formula; post-MVP parameter tuning will improve accuracy
func (s *FSRSScheduler) Schedule(card *model.Card, rating int) ScheduleResult {
	now := time.Now()
	card.LastReviewed = now
	card.Rating = rating

	// Approximate FSRS formula (placeholder for MVP)
	// Real FSRS has 17 parameters; this uses fixed approximate values
	// Rates response difficulty: 1-5 maps to different retention targets

	retentionTarget := 0.9
	switch rating {
	case 1: // Blackout
		retentionTarget = 0.2
	case 2: // Wrong
		retentionTarget = 0.4
	case 3: // Hard
		retentionTarget = 0.6
	case 4: // Good
		retentionTarget = 0.85
	case 5: // Easy
		retentionTarget = 0.95
	}

	// Update retention parameter
	card.Retention = retentionTarget

	// Compute interval based on retention target
	// Approximate formula: higher retention = longer interval
	if card.Interval == 0 {
		card.Interval = 1
	} else {
		// FSRS tends to produce longer intervals for high retention
		card.Interval = int(float64(card.Interval) * (1.0 + retentionTarget))
	}

	card.Interval = minInt(card.Interval, maxInterval)
	card.NextReview = now.AddDate(0, 0, card.Interval)

	return ScheduleResult{
		NextReviewDate:  card.NextReview,
		Interval:        card.Interval,
		EaseOrRetention: card.Retention,
		LastReview:      card.LastReviewed,
	}
}

// ScheduleCard dispatcher routes to the appropriate scheduler based on card.Algorithm
func ScheduleCard(card *model.Card, rating int) error {
	if card == nil {
		return fmt.Errorf("card is nil")
	}

	// Default to SM2 if algorithm is not set
	algorithm := card.Algorithm
	if algorithm == "" {
		algorithm = "SM2"
	}

	var scheduler Algorithm
	switch algorithm {
	case "SM2":
		scheduler = NewSM2Scheduler()
	case "FSRS":
		scheduler = NewFSRSScheduler()
	default:
		return fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	result := scheduler.Schedule(card, rating)

	// Update card with scheduled values
	card.NextReview = result.NextReviewDate
	card.Interval = result.Interval

	// Update the appropriate parameter based on algorithm
	switch algorithm {
	case "SM2":
		card.Ease = result.EaseOrRetention
	case "FSRS":
		card.Retention = result.EaseOrRetention
	}

	return nil
}

// InitializeNewCard initializes a new card with default SRS values
func InitializeNewCard(card *model.Card) *model.Card {
	if card.Ease == 0 {
		card.Ease = defaultEase
	}
	if card.Algorithm == "" {
		card.Algorithm = "SM2"
	}
	card.Interval = 0
	card.NextReview = time.Now()

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
