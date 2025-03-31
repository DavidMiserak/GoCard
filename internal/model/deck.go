// File: internal/model/deck.go

package model

import "time"

// Deck represents a collection of flashcards
type Deck struct {
	ID          string // Will be filepath of the deck (directory)
	Name        string // Will be base name of the directory
	Description string
	Cards       []Card // TODO: Make a tool to import Markdown files in directory to cards
	CreatedAt   time.Time
	LastStudied time.Time
}
