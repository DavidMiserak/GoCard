// File: internal/deck/deck.go
package deck

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// Deck represents a collection of cards organized in a directory
type Deck struct {
	Name       string           // Deck name (directory name)
	Path       string           // Directory path
	Cards      []*card.Card     // Cards directly in this deck
	SubDecks   map[string]*Deck // Child decks mapped by name
	ParentDeck *Deck            // Parent deck (nil for root deck)
	Statistics map[string]int   // Deck statistics cache
}

// NewDeck creates a new deck instance
func NewDeck(path string, parent *Deck) *Deck {
	name := filepath.Base(path)
	return &Deck{
		Name:       name,
		Path:       path,
		Cards:      make([]*card.Card, 0),
		SubDecks:   make(map[string]*Deck),
		ParentDeck: parent,
		Statistics: make(map[string]int),
	}
}

// AddCard adds a card to this deck
func (d *Deck) AddCard(card *card.Card) {
	d.Cards = append(d.Cards, card)
	d.updateStatistics()
}

// RemoveCard removes a card from this deck
func (d *Deck) RemoveCard(card *card.Card) bool {
	// First pass: try filepath comparison
	for i, c := range d.Cards {
		if c.FilePath == card.FilePath {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			d.updateStatistics()
			return true
		}
	}

	// Second pass: try title comparison as fallback
	for i, c := range d.Cards {
		if c.Title == card.Title {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			d.updateStatistics()
			return true
		}
	}

	return false
}

// AddSubDeck adds a subdeck to this deck
func (d *Deck) AddSubDeck(subDeck *Deck) {
	d.SubDecks[subDeck.Name] = subDeck
	subDeck.ParentDeck = d
	d.updateStatistics()
}

// GetAllCards returns all cards in this deck and its subdecks
func (d *Deck) GetAllCards() []*card.Card {
	seen := make(map[string]*card.Card) // Use map to deduplicate by filepath

	// Add cards from this deck
	for _, c := range d.Cards {
		seen[c.FilePath] = c
	}

	// Add cards from subdecks
	for _, subDeck := range d.SubDecks {
		subDeckCards := subDeck.GetAllCards()
		for _, c := range subDeckCards {
			seen[c.FilePath] = c
		}
	}

	// Convert map to slice
	allCards := make([]*card.Card, 0, len(seen))
	for _, c := range seen {
		allCards = append(allCards, c)
	}

	return allCards
}

// GetDeckByPath returns the subdeck at the given relative path or nil if not found
// Path should be relative to this deck, using "/" as separator (e.g., "algorithms/sorting")
func (d *Deck) GetDeckByPath(relativePath string) *Deck {
	if relativePath == "" || relativePath == "." {
		return d
	}

	parts := strings.Split(relativePath, "/")
	currentDeck := d

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		subDeck, exists := currentDeck.SubDecks[part]
		if !exists {
			return nil
		}
		currentDeck = subDeck
	}

	return currentDeck
}

// AllDecks returns a flat slice of all decks (this deck and all subdecks)
func (d *Deck) AllDecks() []*Deck {
	decks := []*Deck{d}

	for _, subDeck := range d.SubDecks {
		decks = append(decks, subDeck.AllDecks()...)
	}

	return decks
}

// PathFromRoot returns the path of this deck relative to the root deck
func (d *Deck) PathFromRoot() string {
	if d.ParentDeck == nil {
		return ""
	}

	if d.ParentDeck.ParentDeck == nil {
		return d.Name
	}

	return filepath.Join(d.ParentDeck.PathFromRoot(), d.Name)
}

// FullName returns a human-readable name including the full path
func (d *Deck) FullName() string {
	pathFromRoot := d.PathFromRoot()
	if pathFromRoot == "" {
		return "Root"
	}
	return pathFromRoot
}

// CountAllCards returns the total number of cards in this deck and all subdecks
func (d *Deck) CountAllCards() int {
	seen := make(map[string]bool) // Track filepaths to avoid duplicates

	// Add cards in this deck
	for _, c := range d.Cards {
		seen[c.FilePath] = true
	}

	// Add cards from all subdecks recursively
	for _, subDeck := range d.SubDecks {
		subDeckCards := subDeck.GetAllCards() // Get all cards from subdeck
		for _, c := range subDeckCards {
			seen[c.FilePath] = true
		}
	}

	return len(seen)
}

// updateStatistics updates the cached statistics for this deck
func (d *Deck) updateStatistics() {
	// Calculate statistics
	totalCards := d.CountAllCards()
	d.Statistics["total_cards"] = totalCards

	// Update parent deck statistics
	if d.ParentDeck != nil {
		d.ParentDeck.updateStatistics()
	}
}

// String returns a string representation of the deck
func (d *Deck) String() string {
	return fmt.Sprintf("%s (%d cards, %d subdecks)",
		d.FullName(),
		len(d.Cards),
		len(d.SubDecks))
}

// GetCardsByTag returns all cards in this deck and its subdecks that have the given tag
func (d *Deck) GetCardsByTag(tag string) []*card.Card {
	var result []*card.Card

	// Check cards in this deck
	for _, c := range d.Cards {
		for _, cardTag := range c.Tags {
			if cardTag == tag {
				result = append(result, c)
				break
			}
		}
	}

	// Check cards in subdecks
	for _, subDeck := range d.SubDecks {
		result = append(result, subDeck.GetCardsByTag(tag)...)
	}

	return result
}
