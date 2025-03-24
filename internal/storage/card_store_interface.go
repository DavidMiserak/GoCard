// File: internal/storage/card_store_interface.go

// Package storage implements the file-based storage system for GoCard.
package storage

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// CardStoreInterface defines the interface for card storage operations
// This allows for mocking in tests and flexibility in implementation
type CardStoreInterface interface {
	// Card operations
	CreateCard(title, question, answer string, tags []string) (*card.Card, error)
	CreateCardInDeck(title, question, answer string, tags []string, deckObj *deck.Deck) (*card.Card, error)
	LoadCard(path string) (*card.Card, error)
	SaveCard(cardObj *card.Card) error
	DeleteCard(cardObj *card.Card) error
	MoveCard(cardObj *card.Card, targetDeck *deck.Deck) error

	// Deck operations
	CreateDeck(name string, parentDeck *deck.Deck) (*deck.Deck, error)
	DeleteDeck(deckObj *deck.Deck) error
	RenameDeck(deckObj *deck.Deck, newName string) error
	GetDeckByPath(path string) (*deck.Deck, error)
	GetDeckByRelativePath(relativePath string) (*deck.Deck, error)

	// Review operations
	ReviewCard(cardObj *card.Card, rating int) error
	GetDueCards() []*card.Card
	GetDueCardsInDeck(deckObj *deck.Deck) []*card.Card
	GetNextDueDate() time.Time

	// Search operations
	GetAllTags() []string
	GetCardsByTag(tag string) []*card.Card
	SearchCards(searchText string) []*card.Card

	// Statistics operations
	GetReviewStats() map[string]interface{}
	GetDeckStats(deckObj *deck.Deck) map[string]interface{}

	// UI compatibility operations
	RenderMarkdown(content string) (string, error)
	RenderMarkdownWithTheme(content string, theme string) (string, error)
	GetAvailableSyntaxThemes() []string
	GetDefaultSyntaxTheme() string

	// File watching operations
	WatchForChanges() error
	StopWatching() error

	// Logging operations
	SetLogLevel(level io.LogLevel)
	DisableLogging()
	EnableLogging()

	// Cleanup
	Close() error
}

// Ensure CardStore implements CardStoreInterface
var _ CardStoreInterface = (*CardStore)(nil)
