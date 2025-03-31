// File: internal/model/deck.go

package model

import "time"

// Deck represents a collection of flashcards
type Deck struct {
	ID          string
	Name        string
	Description string
	Cards       []Card
	CreatedAt   time.Time
	LastStudied time.Time
}
