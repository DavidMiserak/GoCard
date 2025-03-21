// File: internal/card/card.go
package card

import (
	"time"
)

// Card represents a flashcard with its metadata and content
type Card struct {
	FilePath       string    // Not stored in YAML, just for reference
	Title          string    // Extracted from markdown content
	Tags           []string  // Tags for categorization
	Created        time.Time // Creation date
	LastReviewed   time.Time // Last review date
	ReviewInterval int       // Days until next review
	Difficulty     int       // 0-5 difficulty rating
	Question       string    // Extracted from markdown content
	Answer         string    // Extracted from markdown content
}

// NewCard creates a new Card instance
func NewCard(title, question, answer string, tags []string) *Card {
	return &Card{
		Title:          title,
		Tags:           tags,
		Question:       question,
		Answer:         answer,
		Created:        time.Now(),
		LastReviewed:   time.Time{},
		ReviewInterval: 0,
		Difficulty:     0,
	}
}
