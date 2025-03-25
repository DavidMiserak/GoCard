// internal/service/interfaces/storage_service.go
package interfaces

import (
	"github.com/DavidMiserak/GoCard/internal/domain"
)

// StorageService handles persistence of cards and decks
type StorageService interface {
	// Initialization and cleanup
	Initialize(rootDir string) error
	Close() error

	// Card operations
	LoadCard(filePath string) (domain.Card, error)
	UpdateCardMetadata(card domain.Card) error // Updates frontmatter for review state
	ListCardPaths(deckPath string) ([]string, error)

	// Frontmatter operations
	ParseFrontmatter(content []byte) (map[string]interface{}, []byte, error)
	UpdateFrontmatter(content []byte, updates map[string]interface{}) ([]byte, error)

	// Deck operations
	LoadDeck(dirPath string) (domain.Deck, error)
	ListDeckPaths(parentPath string) ([]string, error)

	// Query operations
	FindCardsByTag(tag string) ([]domain.Card, error)
	SearchCards(query string) ([]domain.Card, error)
}
