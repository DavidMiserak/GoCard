// File: internal/storage/interfaces.go

// Package storage implements the file-based storage system for GoCard.
// This file defines interfaces for all major components of the storage system.
package storage

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// CardOperationsInterface defines operations for managing individual cards
type CardOperationsInterface interface {
	CreateCard(title, question, answer string, tags []string) (*card.Card, error)
	CreateCardInDeck(title, question, answer string, tags []string, deckObj *deck.Deck) (*card.Card, error)
	LoadCard(path string) (*card.Card, error)
	SaveCard(cardObj *card.Card) error
	DeleteCard(cardObj *card.Card) error
	MoveCard(cardObj *card.Card, targetDeck *deck.Deck) error
	GetCardByPath(path string) (*card.Card, bool)
}

// DeckOperationsInterface defines operations for managing decks
type DeckOperationsInterface interface {
	CreateDeck(name string, parentDeck *deck.Deck) (*deck.Deck, error)
	DeleteDeck(deckObj *deck.Deck) error
	RenameDeck(deckObj *deck.Deck, newName string) error
	GetDeckByPath(path string) (*deck.Deck, error)
	GetDeckByRelativePath(relativePath string) (*deck.Deck, error)
}

// PersistenceInterface combines card and deck storage operations
type PersistenceInterface interface {
	CardOperationsInterface
	DeckOperationsInterface
}

// ContentRendererInterface defines operations for rendering content
type ContentRendererInterface interface {
	RenderMarkdown(content string) (string, error)
	RenderMarkdownWithTheme(content string, theme string) (string, error)
	GetAvailableSyntaxThemes() []string
	GetDefaultSyntaxTheme() string
}

// ReviewInterface defines operations for reviewing cards using spaced repetition
type ReviewInterface interface {
	ReviewCard(cardObj *card.Card, rating int) error
	GetDueCards() []*card.Card
	GetDueCardsInDeck(deckObj *deck.Deck) []*card.Card
	GetNextDueDate() time.Time
}

// StatsInterface defines operations for retrieving statistics
type StatsInterface interface {
	GetReviewStats() map[string]interface{}
	GetDeckStats(deckObj *deck.Deck) map[string]interface{}
}

// SearchInterface defines operations for searching and filtering cards
type SearchInterface interface {
	GetAllTags() []string
	GetCardsByTag(tag string) []*card.Card
	SearchCards(searchText string) []*card.Card
}

// FileWatcherInterface defines operations for file system monitoring
type FileWatcherInterface interface {
	WatchForChanges() error
	StopWatching() error
}

// LoggingInterface defines operations for controlling log output
type LoggingInterface interface {
	SetLogLevel(io.LogLevel)
	DisableLogging()
	EnableLogging()
}

// StoreInterface defines the main interface for the refactored storage system
// This combines all the component interfaces
type StoreInterface interface {
	PersistenceInterface
	ReviewInterface
	StatsInterface
	SearchInterface
	ContentRendererInterface
	FileWatcherInterface
	LoggingInterface

	// Additional store-specific operations
	GetCardCount() int
	GetDeckCount() int
	Close() error
}

// Ensure CardStore implements all these interfaces
var (
	_ CardOperationsInterface  = (*CardStore)(nil)
	_ DeckOperationsInterface  = (*CardStore)(nil)
	_ PersistenceInterface     = (*CardStore)(nil)
	_ ReviewInterface          = (*CardStore)(nil)
	_ StatsInterface           = (*CardStore)(nil)
	_ SearchInterface          = (*CardStore)(nil)
	_ ContentRendererInterface = (*CardStore)(nil)
	_ FileWatcherInterface     = (*CardStore)(nil)
	_ LoggingInterface         = (*CardStore)(nil)
	_ StoreInterface           = (*CardStore)(nil)
	_ CardStoreInterface       = (*CardStore)(nil)
)
