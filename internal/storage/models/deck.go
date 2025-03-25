// File: internal/storage/models/deck.go

// Package models contains the data models for the GoCard application.
package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Deck represents a collection of cards organized in a directory
type Deck struct {
	cardsMu    sync.RWMutex // Protects Cards slice
	subDecksMu sync.RWMutex // Protects SubDecks map
	statsMu    sync.RWMutex // Protects Statistics map
	parentMu   sync.RWMutex // Protects ParentDeck reference

	Name       string
	Path       string
	Cards      []*Card
	SubDecks   map[string]*Deck
	ParentDeck *Deck
	Statistics map[string]int
}

// NewDeck creates a new deck instance with validation
func NewDeck(path string, parent *Deck) (*Deck, error) {
	if path == "" {
		return nil, fmt.Errorf("deck path cannot be empty")
	}

	name := filepath.Base(path)
	if name == "" || name == "." || name == "/" {
		return nil, fmt.Errorf("invalid deck name derived from path: %s", path)
	}

	return &Deck{
		Name:       name,
		Path:       path,
		Cards:      make([]*Card, 0),
		SubDecks:   make(map[string]*Deck),
		ParentDeck: parent,
		Statistics: make(map[string]int),
	}, nil
}

// AddCard adds a card to this deck (thread-safe)
func (d *Deck) AddCard(card *Card) {
	if card == nil {
		return
	}

	d.cardsMu.Lock()
	d.Cards = append(d.Cards, card)
	d.cardsMu.Unlock()

	// Update statistics without holding the cards lock
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// RemoveCard removes a card from this deck (thread-safe)
func (d *Deck) RemoveCard(card *Card) bool {
	if card == nil {
		return false
	}

	d.cardsMu.Lock()
	defer d.cardsMu.Unlock()

	// Store filepath locally to avoid concurrent access
	cardFilePath := card.GetFilePath()
	cardTitle := card.GetTitle()

	// First pass: try filepath comparison
	for i, c := range d.Cards {
		if c.GetFilePath() == cardFilePath {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			// Don't call updateStatistics while holding the lock
			go d.updateStatisticsAsync() // Use async version to avoid deadlocks
			return true
		}
	}

	// Second pass: try title comparison as fallback
	for i, c := range d.Cards {
		if c.GetTitle() == cardTitle {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			// Don't call updateStatistics while holding the lock
			go d.updateStatisticsAsync() // Use async version to avoid deadlocks
			return true
		}
	}

	return false
}

// AddSubDeck adds a subdeck to this deck (thread-safe)
func (d *Deck) AddSubDeck(subDeck *Deck) {
	if subDeck == nil {
		return
	}

	d.subDecksMu.Lock()
	d.SubDecks[subDeck.Name] = subDeck
	// Save parent reference outside the lock to avoid deadlocks
	parent := d
	d.subDecksMu.Unlock()

	// Set the parent and update stats outside the lock
	subDeck.setParent(parent)
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// RemoveSubDeck removes a subdeck by name (thread-safe)
func (d *Deck) RemoveSubDeck(name string) {
	d.subDecksMu.Lock()
	delete(d.SubDecks, name)
	d.subDecksMu.Unlock()

	// Update statistics after removing a subdeck
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// UpdateStatistics updates the statistics for this deck (thread-safe)
func (d *Deck) UpdateStatistics() {
	go d.updateStatisticsAsync() // Use async version to avoid deadlocks
}

// setParent sets the parent deck - helper to avoid locking issues (thread-safe)
func (d *Deck) setParent(parent *Deck) {
	d.parentMu.Lock()
	d.ParentDeck = parent
	d.parentMu.Unlock()
}

// GetAllCards returns all cards in this deck and its subdecks (thread-safe)
func (d *Deck) GetAllCards() []*Card {
	// First get a copy of our own cards while holding the lock
	d.cardsMu.RLock()
	// Make a copy of cards slice and their filepaths
	ownCards := make([]*Card, len(d.Cards))
	cardPaths := make(map[*Card]string)
	for i, c := range d.Cards {
		ownCards[i] = c
		cardPaths[c] = c.GetFilePath() // Store filepath to avoid race
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
	seen := make(map[string]*Card)

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
			filepath := c.GetFilePath()
			seen[filepath] = c
		}
	}

	// Convert map to slice
	allCards := make([]*Card, 0, len(seen))
	for _, c := range seen {
		allCards = append(allCards, c)
	}

	return allCards
}

// GetDeckByPath returns the subdeck at the given relative path or nil if not found (thread-safe)
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

// AllDecks returns a flat slice of all decks (this deck and all subdecks) (thread-safe)
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

// PathFromRoot returns the path of this deck relative to the root deck (thread-safe)
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

// FullName returns a human-readable name including the full path (thread-safe)
func (d *Deck) FullName() string {
	pathFromRoot := d.PathFromRoot()
	if pathFromRoot == "" {
		return "Root"
	}
	return pathFromRoot
}

// CountAllCards returns the total number of cards in this deck and all subdecks (thread-safe)
func (d *Deck) CountAllCards() int {
	// Just use GetAllCards which is already thread-safe
	return len(d.GetAllCards())
}

// updateStatisticsAsync updates the cached statistics for this deck and its parents
// in a way that avoids deadlocks (thread-safe)
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

// GetName returns the deck's name (thread-safe)
func (d *Deck) GetName() string {
	// No need for locking as Name is never modified after initialization
	return d.Name
}

// GetPath returns the deck's path (thread-safe)
func (d *Deck) GetPath() string {
	// No need for locking as Path is never modified after initialization
	return d.Path
}

// GetParentDeck returns the parent deck (thread-safe)
func (d *Deck) GetParentDeck() *Deck {
	d.parentMu.RLock()
	defer d.parentMu.RUnlock()
	return d.ParentDeck
}

// GetCards returns a copy of the deck's cards (thread-safe)
func (d *Deck) GetCards() []*Card {
	d.cardsMu.RLock()
	defer d.cardsMu.RUnlock()

	// Return a copy to prevent concurrent modification
	cards := make([]*Card, len(d.Cards))
	copy(cards, d.Cards)
	return cards
}

// GetSubDecks returns a copy of the deck's subdecks (thread-safe)
func (d *Deck) GetSubDecks() map[string]*Deck {
	d.subDecksMu.RLock()
	defer d.subDecksMu.RUnlock()

	// Return a copy to prevent concurrent modification
	subDecks := make(map[string]*Deck, len(d.SubDecks))
	for k, v := range d.SubDecks {
		subDecks[k] = v
	}
	return subDecks
}

// GetStatistics returns a copy of the deck's statistics (thread-safe)
func (d *Deck) GetStatistics() map[string]int {
	d.statsMu.RLock()
	defer d.statsMu.RUnlock()

	// Return a copy to prevent concurrent modification
	stats := make(map[string]int, len(d.Statistics))
	for k, v := range d.Statistics {
		stats[k] = v
	}
	return stats
}

// String returns a string representation of the deck (thread-safe)
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

// GetCardsByTag returns all cards in this deck and its subdecks that have the given tag (thread-safe)
func (d *Deck) GetCardsByTag(tag string) []*Card {
	// Get all cards first (already thread-safe)
	allCards := d.GetAllCards()

	var result []*Card
	// Filter cards by tag
	for _, c := range allCards {
		tags := c.GetTags()
		for _, cardTag := range tags {
			if cardTag == tag {
				result = append(result, c)
				break
			}
		}
	}

	return result
}

// DirectAccessForBackwardCompatibility returns the underlying fields directly
// This method is to maintain backward compatibility during refactoring
// and should be removed once all code has been updated
func (d *Deck) DirectAccessForBackwardCompatibility() *Deck {
	return d
}
