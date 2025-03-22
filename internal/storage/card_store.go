// Package storage implements the file-based storage system for GoCard.
// It manages persisting cards and decks to the filesystem as markdown files.
package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
)

// CardStore manages the file-based storage of flashcards
type CardStore struct {
	RootDir  string                // Root directory for all decks
	Cards    map[string]*card.Card // Map of filepath to Card
	Decks    map[string]*deck.Deck // Map of directory path to Deck
	RootDeck *deck.Deck            // The root deck (representing RootDir)
}

// NewCardStore creates a new CardStore with the given root directory
func NewCardStore(rootDir string) (*CardStore, error) {
	// Ensure the directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		if err := os.MkdirAll(rootDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Get absolute path for the root directory to ensure consistent paths
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	store := &CardStore{
		RootDir: absRootDir,
		Cards:   make(map[string]*card.Card),
		Decks:   make(map[string]*deck.Deck),
	}

	// Create the root deck
	store.RootDeck = deck.NewDeck(absRootDir, nil)
	store.Decks[absRootDir] = store.RootDeck

	// Load all cards and organize into deck structure
	if err := store.LoadAllCards(); err != nil {
		return nil, err
	}

	return store, nil
}

// LoadAllCards scans the root directory and loads all markdown files as cards
func (s *CardStore) LoadAllCards() error {
	// First, discover all directories and create the deck structure
	if err := s.discoverDecks(); err != nil {
		return err
	}

	// Then load all cards and organize them into the appropriate decks
	return filepath.WalkDir(s.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Load the card
		cardObj, err := s.LoadCard(path)
		if err != nil {
			return fmt.Errorf("failed to load card %s: %w", path, err)
		}

		s.Cards[path] = cardObj

		// Add the card to the appropriate deck
		dirPath := filepath.Dir(path)
		deckObj, exists := s.Decks[dirPath]
		if !exists {
			// This shouldn't happen if discoverDecks worked correctly
			return fmt.Errorf("deck not found for directory: %s", dirPath)
		}
		deckObj.AddCard(cardObj)

		return nil
	})
}

// discoverDecks builds the deck hierarchy by scanning directories
func (s *CardStore) discoverDecks() error {
	return filepath.WalkDir(s.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process directories
		if !d.IsDir() {
			return nil
		}

		// Skip the root directory as we already created that deck
		if path == s.RootDir {
			return nil
		}

		// Create a deck for this directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Find the parent deck
		parentPath := filepath.Dir(absPath)
		parentDeck, exists := s.Decks[parentPath]
		if !exists {
			return fmt.Errorf("parent deck not found for %s", path)
		}

		// Create the new deck
		newDeck := deck.NewDeck(absPath, parentDeck)
		s.Decks[absPath] = newDeck

		// Add as subdeck to parent
		parentDeck.AddSubDeck(newDeck)

		return nil
	})
}

// WatchForChanges monitors the file system for changes to cards
// This is a placeholder for a more sophisticated file watcher
func (s *CardStore) WatchForChanges() {
	// In a real implementation, you'd use something like fsnotify
	// to watch for file changes and reload cards as needed
	fmt.Println("File watching not implemented yet")
}
