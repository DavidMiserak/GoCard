// File: internal/storage/deck_ops.go

// Package storage implements the file-based storage system for GoCard.
// This file contains operations related to managing decks (directories of cards).
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/deck"
)

// CreateDeck creates a new deck directory with thread-safety
func (s *CardStore) CreateDeck(name string, parentDeck *deck.Deck) (*deck.Deck, error) {
	// Sanitize name for filesystem
	sanitizedName := strings.ToLower(name)
	sanitizedName = strings.ReplaceAll(sanitizedName, " ", "-")
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "-")

	// Determine the path for the new deck
	var parentPath string
	if parentDeck == nil {
		parentDeck = s.RootDeck
		parentPath = s.RootDir
	} else {
		parentPath = parentDeck.Path
	}

	deckPath := filepath.Join(parentPath, sanitizedName)

	// Check if the directory already exists
	if _, err := os.Stat(deckPath); err == nil {
		return nil, fmt.Errorf("deck already exists: %s", deckPath)
	}

	// Create the directory (filesystem operation outside the lock)
	if err := os.MkdirAll(deckPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create deck directory: %w", err)
	}

	// Create the new deck object
	newDeck := deck.NewDeck(deckPath, parentDeck)

	s.decksMu.Lock()
	s.Decks[deckPath] = newDeck
	s.decksMu.Unlock()

	// Add as subdeck to parent (this will handle parent deck locking)
	parentDeck.AddSubDeck(newDeck)

	return newDeck, nil
}

// DeleteDeck removes a deck directory and all contained cards and subdecks
func (s *CardStore) DeleteDeck(deckObj *deck.Deck) error {
	// Don't allow deleting the root deck
	if deckObj == s.RootDeck {
		return fmt.Errorf("cannot delete the root deck")
	}

	// Get all cards and decks to be removed
	allCards := deckObj.GetAllCards()
	allSubDecks := deckObj.AllDecks()
	parentDeck := deckObj.ParentDeck

	// Remove the directory and all its contents (filesystem operation outside lock)
	if err := os.RemoveAll(deckObj.Path); err != nil {
		return fmt.Errorf("failed to delete deck directory: %w", err)
	}

	// Remove all cards in this deck and its subdecks from our maps
	s.cardsMu.Lock()
	for _, card := range allCards {
		delete(s.Cards, card.FilePath)
	}
	s.cardsMu.Unlock()

	// Remove all subdecks from our map (including the deck itself)
	s.decksMu.Lock()
	for _, subDeck := range allSubDecks {
		if subDeck != deckObj { // Skip the deck itself for now
			delete(s.Decks, subDeck.Path)
		}
	}
	// Now remove the deck itself
	delete(s.Decks, deckObj.Path)
	s.decksMu.Unlock()

	// Remove the deck from its parent (thread-safe)
	if parentDeck != nil {
		parentDeck.RemoveSubDeck(deckObj.Name)
	}

	return nil
}

// RenameDeck renames a deck directory
func (s *CardStore) RenameDeck(deckObj *deck.Deck, newName string) error {
	// Don't allow renaming the root deck
	if deckObj == s.RootDeck {
		return fmt.Errorf("cannot rename the root deck")
	}

	// Sanitize name for filesystem
	sanitizedName := strings.ToLower(newName)
	sanitizedName = strings.ReplaceAll(sanitizedName, " ", "-")
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "-")

	// Calculate the new path
	parentPath := filepath.Dir(deckObj.Path)
	newPath := filepath.Join(parentPath, sanitizedName)

	// Check if the new path already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("deck with name %s already exists", newName)
	}

	// Get references to things we'll need to update
	parentDeck := deckObj.ParentDeck
	oldPath := deckObj.Path

	// Get all cards in this deck
	cardsInDeck := deckObj.GetAllCards()

	// Get all subdecks
	allSubDecks := deckObj.AllDecks()

	// Rename the directory (filesystem operation outside lock)
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename deck directory: %w", err)
	}

	// Update the in-memory structures

	// Update our deck map references
	s.decksMu.Lock()
	delete(s.Decks, oldPath) // Remove old reference

	// Need to update deck paths since the filesystem paths changed
	// First update the main deck's path
	deckObj.Path = newPath
	deckObj.Name = sanitizedName
	s.Decks[newPath] = deckObj

	// Update all subdeck paths
	for _, subDeck := range allSubDecks {
		if subDeck == deckObj {
			continue // Skip the deck itself as we already updated it
		}

		// Calculate the new subdeck path
		oldSubPath := subDeck.Path
		relPath, err := filepath.Rel(oldPath, oldSubPath)
		if err != nil {
			s.logger.Error("Error calculating relative path: %v", err)
			continue
		}

		newSubPath := filepath.Join(newPath, relPath)

		// Update the subdeck's path in our map
		delete(s.Decks, oldSubPath)
		subDeck.Path = newSubPath
		s.Decks[newSubPath] = subDeck
	}
	s.decksMu.Unlock()

	// Update the parent's subdeck reference if it exists
	if parentDeck != nil {
		// The parent deck needs to update its SubDecks map to reflect the name change
		// This is handled by the deck's public methods
		parentDeck.RemoveSubDeck(filepath.Base(oldPath))
		parentDeck.AddSubDeck(deckObj)
	}

	// Update card filepaths in our maps
	s.cardsMu.Lock()
	for _, cardObj := range cardsInDeck {
		oldCardPath := cardObj.FilePath
		relPath, err := filepath.Rel(oldPath, oldCardPath)
		if err != nil {
			s.logger.Error("Error calculating relative path for card: %v", err)
			continue
		}

		newCardPath := filepath.Join(newPath, relPath)

		// Update the card's filepath
		delete(s.Cards, oldCardPath)
		cardObj.FilePath = newCardPath
		s.Cards[newCardPath] = cardObj
	}
	s.cardsMu.Unlock()

	return nil
}

// GetDeckByPath returns the deck at the given path
func (s *CardStore) GetDeckByPath(path string) (*deck.Deck, error) {
	// If path is empty or ".", return the root deck
	if path == "" || path == "." {
		return s.RootDeck, nil
	}

	// Check if the path is absolute
	if filepath.IsAbs(path) {
		s.decksMu.RLock()
		deckObj, exists := s.Decks[path]
		s.decksMu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("deck not found: %s", path)
		}
		return deckObj, nil
	}

	// Otherwise, treat it as relative to the root deck
	return s.GetDeckByRelativePath(path)
}

// GetDeckByRelativePath returns the deck at the given path relative to the root
func (s *CardStore) GetDeckByRelativePath(relativePath string) (*deck.Deck, error) {
	// If the path is empty or ".", return the root deck
	if relativePath == "" || relativePath == "." {
		return s.RootDeck, nil
	}

	// Convert the relative path to an absolute path
	absPath := filepath.Join(s.RootDir, relativePath)

	// Look up the deck
	s.decksMu.RLock()
	deckObj, exists := s.Decks[absPath]
	s.decksMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("deck not found: %s", relativePath)
	}

	return deckObj, nil
}
