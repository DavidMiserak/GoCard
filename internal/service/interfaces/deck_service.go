// internal/service/interfaces/deck_service.go
package interfaces

import (
	"github.com/DavidMiserak/GoCard/internal/domain"
)

// DeckService manages operations on decks and card collections
type DeckService interface {
	// Deck read operations
	GetDeck(deckPath string) (domain.Deck, error)

	// Deck hierarchy operations
	GetSubdecks(deckPath string) ([]domain.Deck, error)
	GetParentDeck(deckPath string) (domain.Deck, error)

	// Card collection operations
	GetCards(deckPath string) ([]domain.Card, error)
	GetDueCards(deckPath string) ([]domain.Card, error)
	GetCardStats(deckPath string) (map[string]int, error)
}
