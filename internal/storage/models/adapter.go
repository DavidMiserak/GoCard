// File: internal/storage/models/adapter.go

// Package models contains the data models for the GoCard application.
package models

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
)

// ConvertToLegacyCard converts a models.Card to the legacy card.Card format
func ConvertToLegacyCard(modernCard *Card) *card.Card {
	if modernCard == nil {
		return nil
	}

	return &card.Card{
		FilePath:       modernCard.GetFilePath(),
		Title:          modernCard.GetTitle(),
		Tags:           modernCard.GetTags(),
		Created:        modernCard.GetCreatedTime(),
		LastReviewed:   modernCard.GetLastReviewedTime(),
		ReviewInterval: modernCard.GetReviewInterval(),
		Difficulty:     modernCard.GetDifficulty(),
		Question:       modernCard.GetQuestion(),
		Answer:         modernCard.GetAnswer(),
	}
}

// ConvertFromLegacyCard converts a legacy card.Card to the new models.Card format
func ConvertFromLegacyCard(legacyCard *card.Card) (*Card, error) {
	if legacyCard == nil {
		return nil, nil
	}

	// Create using constructor for validation
	newCard, err := NewCard(
		legacyCard.Title,
		legacyCard.Question,
		legacyCard.Answer,
		legacyCard.Tags,
	)
	if err != nil {
		return nil, err
	}

	// Set additional fields that aren't part of the constructor
	newCard.SetFilePath(legacyCard.FilePath)
	newCard.SetLastReviewedTime(legacyCard.LastReviewed)
	newCard.SetReviewInterval(legacyCard.ReviewInterval)
	if err := newCard.SetDifficulty(legacyCard.Difficulty); err != nil {
		// If difficulty is out of range, default to 0
		err := newCard.SetDifficulty(0)
		if err != nil {
			return nil, err
		}
	}

	return newCard, nil
}

// ConvertToLegacyDeck converts a models.Deck to the legacy deck.Deck format
func ConvertToLegacyDeck(modernDeck *Deck) *deck.Deck {
	if modernDeck == nil {
		return nil
	}

	// Convert cards to legacy format
	cards := modernDeck.GetCards()
	legacyCards := make([]*card.Card, len(cards))
	for i, c := range cards {
		legacyCards[i] = ConvertToLegacyCard(c)
	}

	// Get parent deck
	parentDeck := modernDeck.GetParentDeck()
	var legacyParent *deck.Deck
	if parentDeck != nil {
		legacyParent = ConvertToLegacyDeck(parentDeck)
	}

	// Create legacy deck
	legacyDeck := deck.NewDeck(modernDeck.GetPath(), legacyParent)

	// Convert subdecks recursively
	subDecks := modernDeck.GetSubDecks()
	for _, subDeck := range subDecks {
		legacySubDeck := ConvertToLegacyDeck(subDeck)
		legacyDeck.AddSubDeck(legacySubDeck)
	}

	// Add cards to legacy deck
	for _, legacyCard := range legacyCards {
		legacyDeck.AddCard(legacyCard)
	}

	return legacyDeck
}

// ConvertFromLegacyDeck converts a legacy deck.Deck to the new models.Deck format
func ConvertFromLegacyDeck(legacyDeck *deck.Deck) (*Deck, error) {
	if legacyDeck == nil {
		return nil, nil
	}

	// Create new deck
	modernDeck, err := NewDeck(legacyDeck.Path, nil)
	if err != nil {
		return nil, err
	}

	// Convert cards
	for _, legacyCard := range legacyDeck.Cards {
		modernCard, err := ConvertFromLegacyCard(legacyCard)
		if err != nil {
			continue // Skip cards that can't be converted
		}
		modernDeck.AddCard(modernCard)
	}

	return modernDeck, nil
}

// LegacyCardAdapter wraps a legacy card to implement CardInterface
type LegacyCardAdapter struct {
	card *card.Card
}

// NewLegacyCardAdapter creates a new adapter for a legacy card
func NewLegacyCardAdapter(c *card.Card) CardInterface {
	return &LegacyCardAdapter{card: c}
}

// GetLegacyCard returns the underlying legacy card
func (a *LegacyCardAdapter) GetLegacyCard() *card.Card {
	return a.card
}

// Implement CardInterface methods
func (a *LegacyCardAdapter) GetTitle() string               { return a.card.Title }
func (a *LegacyCardAdapter) GetQuestion() string            { return a.card.Question }
func (a *LegacyCardAdapter) GetAnswer() string              { return a.card.Answer }
func (a *LegacyCardAdapter) GetTags() []string              { return a.card.Tags }
func (a *LegacyCardAdapter) GetFilePath() string            { return a.card.FilePath }
func (a *LegacyCardAdapter) GetCreatedTime() time.Time      { return a.card.Created }
func (a *LegacyCardAdapter) GetLastReviewedTime() time.Time { return a.card.LastReviewed }
func (a *LegacyCardAdapter) GetReviewInterval() int         { return a.card.ReviewInterval }
func (a *LegacyCardAdapter) GetDifficulty() int             { return a.card.Difficulty }

// LegacyDeckAdapter wraps a legacy deck to implement DeckInterface
type LegacyDeckAdapter struct {
	deck *deck.Deck
}

// NewLegacyDeckAdapter creates a new adapter for a legacy deck
func NewLegacyDeckAdapter(d *deck.Deck) DeckInterface {
	return &LegacyDeckAdapter{deck: d}
}

// GetLegacyDeck returns the underlying legacy deck
func (a *LegacyDeckAdapter) GetLegacyDeck() *deck.Deck {
	return a.deck
}

// Implement DeckInterface methods
func (a *LegacyDeckAdapter) GetName() string { return a.deck.Name }
func (a *LegacyDeckAdapter) GetPath() string { return a.deck.Path }
func (a *LegacyDeckAdapter) GetCards() []CardInterface {
	cards := a.deck.Cards
	result := make([]CardInterface, len(cards))
	for i, c := range cards {
		result[i] = NewLegacyCardAdapter(c)
	}
	return result
}
func (a *LegacyDeckAdapter) GetSubDecks() map[string]DeckInterface {
	subDecks := make(map[string]DeckInterface)
	for name, d := range a.deck.SubDecks {
		subDecks[name] = NewLegacyDeckAdapter(d)
	}
	return subDecks
}
func (a *LegacyDeckAdapter) GetParentDeck() DeckInterface {
	if a.deck.ParentDeck == nil {
		return nil
	}
	return NewLegacyDeckAdapter(a.deck.ParentDeck)
}
func (a *LegacyDeckAdapter) GetAllCards() []CardInterface {
	cards := a.deck.GetAllCards()
	result := make([]CardInterface, len(cards))
	for i, c := range cards {
		result[i] = NewLegacyCardAdapter(c)
	}
	return result
}
