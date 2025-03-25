// File: internal/storage/interfaces.go

package storage

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// All existing interface definitions remain the same as in your current file...

// CardStoreInterface combines all interfaces for comprehensive card storage operations
type CardStoreInterface interface {
	// Card operations
	CreateCard(title, question, answer string, tags []string) (*card.Card, error)
	CreateCardInDeck(title, question, answer string, tags []string, deckObj *deck.Deck) (*card.Card, error)
	LoadCard(path string) (*card.Card, error)
	SaveCard(cardObj *card.Card) error
	DeleteCard(cardObj *card.Card) error
	MoveCard(cardObj *card.Card, targetDeck *deck.Deck) error
	GetCardByPath(path string) (*card.Card, bool)

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

	// Stats
	GetReviewStats() map[string]interface{}
	GetDeckStats(deckObj *deck.Deck) map[string]interface{}
	GetCardCount() int
	GetDeckCount() int

	// Search
	GetAllTags() []string
	GetCardsByTag(tag string) []*card.Card
	SearchCards(searchText string) []*card.Card

	// Content rendering
	RenderMarkdown(content string) (string, error)
	RenderMarkdownWithTheme(content string, theme string) (string, error)
	GetAvailableSyntaxThemes() []string
	GetDefaultSyntaxTheme() string

	// File watching
	WatchForChanges() error
	StopWatching() error

	// Logging
	SetLogLevel(io.LogLevel)
	DisableLogging()
	EnableLogging()

	// Close resources
	Close() error
}

// Ensure CardStore implements all the interfaces
var (
	_ CardStoreInterface = (*CardStore)(nil)
)
