// File: internal/deck/deck.go

package deck

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// Deck represents a collection of cards organized in a directory
type Deck struct {
	cardsMu    sync.RWMutex // Protects Cards slice
	subDecksMu sync.RWMutex // Protects SubDecks map
	statsMu    sync.RWMutex // Protects Statistics map
	parentMu   sync.RWMutex // Protects ParentDeck reference

	Name       string
	Path       string
	Cards      []*card.Card
	SubDecks   map[string]*Deck
	ParentDeck *Deck
	Statistics map[string]int
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
	d.cardsMu.Lock()
	d.Cards = append(d.Cards, card)
	d.cardsMu.Unlock()

	// Update statistics without holding the cards lock
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// RemoveCard removes a card from this deck
func (d *Deck) RemoveCard(card *card.Card) bool {
	d.cardsMu.Lock()
	defer d.cardsMu.Unlock()

	// First pass: try filepath comparison
	for i, c := range d.Cards {
		if c.FilePath == card.FilePath {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			// Don't call updateStatistics while holding the lock
			go d.updateStatisticsAsync() // Use async version to avoid deadlocks
			return true
		}
	}

	// Second pass: try title comparison as fallback
	for i, c := range d.Cards {
		if c.Title == card.Title {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			// Don't call updateStatistics while holding the lock
			go d.updateStatisticsAsync() // Use async version to avoid deadlocks
			return true
		}
	}

	return false
}

// AddSubDeck adds a subdeck to this deck
func (d *Deck) AddSubDeck(subDeck *Deck) {
	d.subDecksMu.Lock()
	d.SubDecks[subDeck.Name] = subDeck
	// Save parent reference outside the lock to avoid deadlocks
	parent := d
	d.subDecksMu.Unlock()

	// Set the parent and update stats outside the lock
	subDeck.setParent(parent)
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// RemoveSubDeck removes a subdeck by name (public method)
func (d *Deck) RemoveSubDeck(name string) {
	d.subDecksMu.Lock()
	delete(d.SubDecks, name)
	d.subDecksMu.Unlock()

	// Update statistics after removing a subdeck
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// UpdateStatistics updates the statistics for this deck (exported version)
func (d *Deck) UpdateStatistics() {
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// setParent sets the parent deck - helper to avoid locking issues
func (d *Deck) setParent(parent *Deck) {
	d.parentMu.Lock()
	d.ParentDeck = parent
	d.parentMu.Unlock()
}

// GetAllCards returns all cards in this deck and its subdecks
func (d *Deck) GetAllCards() []*card.Card {
	// First get a copy of our own cards while holding the lock
	d.cardsMu.RLock()
	// Make a copy of cards slice and their filepaths
	ownCards := make([]*card.Card, len(d.Cards))
	cardPaths := make(map[*card.Card]string)
	for i, c := range d.Cards {
		ownCards[i] = c
		cardPaths[c] = c.FilePath // Store filepath to avoid race
	}
	d.cardsMu.RUnlock()

	// Get a copy of subdeck references
	d.subDecksMu.RLock()
	subDecks := make([]*Deck, 0, len(d.SubDecks))
	for _, sd := range d.SubDecks {
		subDecks = append(subDecks, sd)
	}
	d.subDecksMu.RUnlock()

	// Use a map to deduplicate cards
	seen := make(map[string]*card.Card)

	// Add own cards to result using the captured filepaths
	for _, c := range ownCards {
		path := cardPaths[c]
		seen[path] = c
	}

	// Process subdecks without holding our lock
	for _, subDeck := range subDecks {
		subDeckCards := subDeck.GetAllCards()
		for _, c := range subDeckCards {
			// Safely capture the filepath to avoid race conditions
			filepath := c.FilePath
			seen[filepath] = c
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
func (d *Deck) GetDeckByPath(relativePath string) *Deck {
	if relativePath == "" || relativePath == "." {
		return d
	}

	parts := strings.Split(relativePath, "/")
	var currentDeck *Deck = d

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		currentDeck.subDecksMu.RLock()
		subDeck, exists := currentDeck.SubDecks[part]
		currentDeck.subDecksMu.RUnlock()

		if !exists {
			return nil
		}
		currentDeck = subDeck
	}

	return currentDeck
}

// AllDecks returns a flat slice of all decks (this deck and all subdecks)
func (d *Deck) AllDecks() []*Deck {
	// Get references to subdecks while holding the lock
	d.subDecksMu.RLock()
	subDecks := make([]*Deck, 0, len(d.SubDecks))
	for _, sd := range d.SubDecks {
		subDecks = append(subDecks, sd)
	}
	d.subDecksMu.RUnlock()

	// Start with this deck
	decks := []*Deck{d}

	// Process subdecks without holding our lock
	for _, subDeck := range subDecks {
		decks = append(decks, subDeck.AllDecks()...)
	}

	return decks
}

// PathFromRoot returns the path of this deck relative to the root deck
func (d *Deck) PathFromRoot() string {
	d.parentMu.RLock()
	parent := d.ParentDeck
	name := d.Name
	d.parentMu.RUnlock()

	if parent == nil {
		return ""
	}

	parentPath := parent.PathFromRoot()
	if parentPath == "" {
		return name
	}

	return filepath.Join(parentPath, name)
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
	// Just use GetAllCards which is already thread-safe
	return len(d.GetAllCards())
}

// updateStatisticsAsync updates the cached statistics for this deck and its parents
// in a way that avoids deadlocks
func (d *Deck) updateStatisticsAsync() {
	// Calculate total cards without holding a lock
	totalCards := d.CountAllCards()

	// Update the statistics
	d.statsMu.Lock()
	d.Statistics["total_cards"] = totalCards
	d.statsMu.Unlock()

	// Update parent deck statistics if needed
	d.parentMu.RLock()
	parent := d.ParentDeck
	d.parentMu.RUnlock()

	if parent != nil {
		// Start a new goroutine to update the parent's statistics
		// This prevents deadlocks in the statistics update chain
		go parent.updateStatisticsAsync()
	}
}

// String returns a string representation of the deck
func (d *Deck) String() string {
	name := d.FullName()

	d.cardsMu.RLock()
	cardCount := len(d.Cards)
	d.cardsMu.RUnlock()

	d.subDecksMu.RLock()
	subdeckCount := len(d.SubDecks)
	d.subDecksMu.RUnlock()

	return fmt.Sprintf("%s (%d cards, %d subdecks)",
		name, cardCount, subdeckCount)
}

// GetCardsByTag returns all cards in this deck and its subdecks that have the given tag
func (d *Deck) GetCardsByTag(tag string) []*card.Card {
	// Get all cards first (already thread-safe)
	allCards := d.GetAllCards()

	var result []*card.Card
	// Filter cards by tag
	for _, c := range allCards {
		for _, cardTag := range c.Tags {
			if cardTag == tag {
				result = append(result, c)
				break
			}
		}
	}

	return result
}
