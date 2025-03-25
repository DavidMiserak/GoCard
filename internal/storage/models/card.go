// File: internal/storage/models/card.go

// Package models contains the data models for the GoCard application.
package models

import (
	"errors"
	"sync"
	"time"
)

// Card represents a flashcard with its metadata and content
type Card struct {
	mu             sync.RWMutex // Protects all fields
	FilePath       string       // Not stored in YAML, just for reference
	Title          string       // Extracted from markdown content
	Tags           []string     // Tags for categorization
	Created        time.Time    // Creation date
	LastReviewed   time.Time    // Last review date
	ReviewInterval int          // Days until next review
	Difficulty     int          // 0-5 difficulty rating
	Question       string       // Extracted from markdown content
	Answer         string       // Extracted from markdown content
}

// NewCard creates a new Card instance with validation
func NewCard(title, question, answer string, tags []string) (*Card, error) {
	// Validate required fields
	if title == "" {
		return nil, errors.New("card title cannot be empty")
	}

	if question == "" {
		return nil, errors.New("card question cannot be empty")
	}

	// Answer can be empty in some cases, so no validation for it

	// Create a copy of tags to prevent external modification
	tagsCopy := make([]string, len(tags))
	copy(tagsCopy, tags)

	return &Card{
		Title:          title,
		Tags:           tagsCopy,
		Question:       question,
		Answer:         answer,
		Created:        time.Now(),
		LastReviewed:   time.Time{},
		ReviewInterval: 0,
		Difficulty:     0,
	}, nil
}

// GetTitle returns the card's title (thread-safe)
func (c *Card) GetTitle() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Title
}

// SetTitle sets the card's title (thread-safe)
func (c *Card) SetTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Title = title
	return nil
}

// GetQuestion returns the card's question (thread-safe)
func (c *Card) GetQuestion() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Question
}

// SetQuestion sets the card's question (thread-safe)
func (c *Card) SetQuestion(question string) error {
	if question == "" {
		return errors.New("question cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Question = question
	return nil
}

// GetAnswer returns the card's answer (thread-safe)
func (c *Card) GetAnswer() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Answer
}

// SetAnswer sets the card's answer (thread-safe)
func (c *Card) SetAnswer(answer string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Answer = answer
}

// GetTags returns a copy of the card's tags (thread-safe)
func (c *Card) GetTags() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	tags := make([]string, len(c.Tags))
	copy(tags, c.Tags)
	return tags
}

// SetTags sets the card's tags (thread-safe)
func (c *Card) SetTags(tags []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to prevent external modification
	c.Tags = make([]string, len(tags))
	copy(c.Tags, tags)
}

// GetFilePath returns the card's file path (thread-safe)
func (c *Card) GetFilePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.FilePath
}

// SetFilePath sets the card's file path (thread-safe)
func (c *Card) SetFilePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FilePath = path
}

// GetCreatedTime returns the card's creation time (thread-safe)
func (c *Card) GetCreatedTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Created
}

// GetLastReviewedTime returns the card's last reviewed time (thread-safe)
func (c *Card) GetLastReviewedTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastReviewed
}

// SetLastReviewedTime sets the card's last reviewed time (thread-safe)
func (c *Card) SetLastReviewedTime(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastReviewed = t
}

// GetReviewInterval returns the card's review interval (thread-safe)
func (c *Card) GetReviewInterval() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ReviewInterval
}

// SetReviewInterval sets the card's review interval (thread-safe)
func (c *Card) SetReviewInterval(interval int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ReviewInterval = interval
}

// GetDifficulty returns the card's difficulty (thread-safe)
func (c *Card) GetDifficulty() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Difficulty
}

// SetDifficulty sets the card's difficulty (thread-safe)
func (c *Card) SetDifficulty(difficulty int) error {
	if difficulty < 0 || difficulty > 5 {
		return errors.New("difficulty must be between 0 and 5")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Difficulty = difficulty
	return nil
}

// DirectAccessForBackwardCompatibility returns the underlying fields directly
// This method is to maintain backward compatibility during refactoring
// and should be removed once all code has been updated
func (c *Card) DirectAccessForBackwardCompatibility() *Card {
	return c
}
