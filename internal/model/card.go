// File: internal/model/card.go

package model

import "time"

// TODO: Make a tool to import Markdown files in directory to cards

// Card represents a flashcard
type Card struct {
	ID           string // Will be the filepath of the card
	Question     string
	Answer       string
	DeckID       string // Will the filepath of the deck (directory)
	LastReviewed time.Time
	NextReview   time.Time
	Ease         float64
	Interval     int // in days
	Rating       int // 1-5 rating per SmartMemo2 Algorithm
}
