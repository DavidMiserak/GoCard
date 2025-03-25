// internal/domain/card.go
package domain

import (
	"time"
)

// Card represents a flashcard with question, answer, and metadata
type Card struct {
	FilePath       string                 // Path to the card file (serves as identifier)
	Title          string                 // Card title
	Question       string                 // Question text (supports Markdown)
	Answer         string                 // Answer text (supports Markdown)
	Tags           []string               // Tags for categorization
	Created        time.Time              // Creation timestamp
	LastReviewed   time.Time              // Last review timestamp - from frontmatter
	ReviewInterval int                    // Current interval in days - from frontmatter
	Difficulty     int                    // Difficulty rating (0-5) - from frontmatter
	RawContent     string                 // Raw markdown content including frontmatter
	Frontmatter    map[string]interface{} // Parsed frontmatter
}

// IsDue determines if a card is due for review based on its interval
func (c *Card) IsDue() bool {
	if c.LastReviewed.IsZero() {
		// Card has never been reviewed
		return true
	}

	dueDate := c.LastReviewed.AddDate(0, 0, c.ReviewInterval)
	return time.Now().After(dueDate)
}

// GetDueDate returns the next due date for this card
func (c *Card) GetDueDate() time.Time {
	if c.LastReviewed.IsZero() {
		return time.Now()
	}
	return c.LastReviewed.AddDate(0, 0, c.ReviewInterval)
}

// NewCard creates a new Card with default values
func NewCard(filePath string) *Card {
	return &Card{
		FilePath:       filePath,
		ReviewInterval: 1, // Default interval is 1 day
		Difficulty:     3, // Default difficulty (middle of 0-5 range)
		Tags:           []string{},
		Frontmatter:    make(map[string]interface{}),
	}
}
