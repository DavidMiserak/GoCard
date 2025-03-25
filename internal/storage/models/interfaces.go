// File: internal/storage/models/interfaces.go
package models

import "time"

// CardInterface defines the core methods for a card
type CardInterface interface {
	GetTitle() string
	GetQuestion() string
	GetAnswer() string
	GetTags() []string
	GetFilePath() string
	GetCreatedTime() time.Time
	GetLastReviewedTime() time.Time
	GetReviewInterval() int
	GetDifficulty() int
}

// DeckInterface defines the core methods for a deck
type DeckInterface interface {
	GetName() string
	GetPath() string
	GetCards() []CardInterface
	GetSubDecks() map[string]DeckInterface
	GetParentDeck() DeckInterface
	GetAllCards() []CardInterface
}
