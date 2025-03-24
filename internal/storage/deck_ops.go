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

// CreateDeck creates a new deck directory
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

	// Create the directory
	if err := os.MkdirAll(deckPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create deck directory: %w", err)
	}

	// Create the deck object
	newDeck := deck.NewDeck(deckPath, parentDeck)
	s.Decks[deckPath] = newDeck

	// Add as subdeck to parent
	parentDeck.AddSubDeck(newDeck)

	return newDeck, nil
}

// DeleteDeck removes a deck directory and all contained cards and subdecks
func (s *CardStore) DeleteDeck(deckObj *deck.Deck) error {
	// Don't allow deleting the root deck
	if deckObj == s.RootDeck {
		return fmt.Errorf("cannot delete the root deck")
	}

	// Remove the directory and all its contents
	if err := os.RemoveAll(deckObj.Path); err != nil {
		return fmt.Errorf("failed to delete deck directory: %w", err)
	}

	// Remove all cards in this deck and its subdecks from our maps
	for _, card := range deckObj.GetAllCards() {
		delete(s.Cards, card.FilePath)
	}

	// Remove all subdecks from our map
	for _, subDeck := range deckObj.AllDecks() {
		if subDeck != deckObj { // Skip the deck itself, we'll remove it separately
			delete(s.Decks, subDeck.Path)
		}
	}

	// Remove the deck from its parent
	if deckObj.ParentDeck != nil {
		delete(deckObj.ParentDeck.SubDecks, deckObj.Name)
	}

	// Remove the deck from our map
	delete(s.Decks, deckObj.Path)

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

	// Rename the directory
	if err := os.Rename(deckObj.Path, newPath); err != nil {
		return fmt.Errorf("failed to rename deck directory: %w", err)
	}

	// Update the deck object
	oldPath := deckObj.Path
	deckObj.Path = newPath
	deckObj.Name = sanitizedName

	// Update the deck in our map
	delete(s.Decks, oldPath)
	s.Decks[newPath] = deckObj

	// Update the parent deck's subdeck map
	if deckObj.ParentDeck != nil {
		delete(deckObj.ParentDeck.SubDecks, filepath.Base(oldPath))
		deckObj.ParentDeck.SubDecks[sanitizedName] = deckObj
	}

	// Update paths for all cards in this deck
	for _, cardObj := range deckObj.Cards {
		oldCardPath := cardObj.FilePath
		fileName := filepath.Base(oldCardPath)
		newCardPath := filepath.Join(newPath, fileName)

		// Update the card's filepath
		cardObj.FilePath = newCardPath

		// Update our card map
		delete(s.Cards, oldCardPath)
		s.Cards[newCardPath] = cardObj
	}

	// Recursively update paths for all subdecks and their cards
	for _, subDeck := range deckObj.SubDecks {
		// The recursive directory rename is handled by the OS
		// We just need to update our internal references
		subDeckOldPath := subDeck.Path
		subDeckNewPath := filepath.Join(newPath, subDeck.Name)
		subDeck.Path = subDeckNewPath

		// Update the deck in our map
		delete(s.Decks, subDeckOldPath)
		s.Decks[subDeckNewPath] = subDeck

		// Update paths for all cards in this subdeck
		for _, cardObj := range subDeck.Cards {
			oldCardPath := cardObj.FilePath
			fileName := filepath.Base(oldCardPath)
			newCardPath := filepath.Join(subDeckNewPath, fileName)

			// Update the card's filepath
			cardObj.FilePath = newCardPath

			// Update our card map
			delete(s.Cards, oldCardPath)
			s.Cards[newCardPath] = cardObj
		}
	}

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
		deckObj, exists := s.Decks[path]
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
	deckObj, exists := s.Decks[absPath]
	if !exists {
		return nil, fmt.Errorf("deck not found: %s", relativePath)
	}

	return deckObj, nil
}
