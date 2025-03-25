// internal/service/deck/deck_service.go
package deck

import (
	"fmt"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
)

// DefaultDeckService implements the DeckService interface
type DefaultDeckService struct {
	storage interfaces.StorageService
	cardSvc interfaces.CardService
}

// NewDeckService creates a new deck service
func NewDeckService(storage interfaces.StorageService, cardSvc interfaces.CardService) interfaces.DeckService {
	return &DefaultDeckService{
		storage: storage,
		cardSvc: cardSvc,
	}
}

// GetDeck retrieves a deck by path
func (ds *DefaultDeckService) GetDeck(deckPath string) (domain.Deck, error) {
	return ds.storage.LoadDeck(deckPath)
}

// GetSubdecks finds all subdirectories (subdecks) in a deck
func (ds *DefaultDeckService) GetSubdecks(deckPath string) ([]domain.Deck, error) {
	paths, err := ds.storage.ListDeckPaths(deckPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list subdeck paths: %w", err)
	}

	var decks []domain.Deck
	for _, path := range paths {
		deck, err := ds.storage.LoadDeck(path)
		if err != nil {
			// Log but continue
			continue
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// GetParentDeck returns the parent deck of a given deck
func (ds *DefaultDeckService) GetParentDeck(deckPath string) (domain.Deck, error) {
	parentPath := filepath.Dir(deckPath)
	if parentPath == deckPath {
		return domain.Deck{}, fmt.Errorf("deck has no parent")
	}
	return ds.storage.LoadDeck(parentPath)
}

// GetCards retrieves all cards in a deck
func (ds *DefaultDeckService) GetCards(deckPath string) ([]domain.Card, error) {
	cardPaths, err := ds.storage.ListCardPaths(deckPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list card paths: %w", err)
	}

	var cards []domain.Card
	for _, path := range cardPaths {
		card, err := ds.storage.LoadCard(path)
		if err != nil {
			// Log but continue
			continue
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// GetDueCards retrieves all due cards in a deck
func (ds *DefaultDeckService) GetDueCards(deckPath string) ([]domain.Card, error) {
	cards, err := ds.GetCards(deckPath)
	if err != nil {
		return nil, err
	}

	var dueCards []domain.Card
	for _, card := range cards {
		if ds.cardSvc.IsDue(card.FilePath) {
			dueCards = append(dueCards, card)
		}
	}

	return dueCards, nil
}

// GetCardStats returns statistics about cards in a deck
func (ds *DefaultDeckService) GetCardStats(deckPath string) (map[string]int, error) {
	cards, err := ds.GetCards(deckPath)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":   len(cards),
		"due":     0,
		"new":     0,
		"learned": 0,
	}

	for _, card := range cards {
		if card.LastReviewed.IsZero() {
			stats["new"]++
		} else {
			stats["learned"]++
		}

		if ds.cardSvc.IsDue(card.FilePath) {
			stats["due"]++
		}
	}

	return stats, nil
}

// Ensure DefaultDeckService implements DeckService
var _ interfaces.DeckService = (*DefaultDeckService)(nil)
