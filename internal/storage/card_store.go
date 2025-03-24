// File: internal/storage/card_store.go

// Package storage implements the file-based storage system for GoCard.
// It manages persisting cards and decks to the filesystem as markdown files.
package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// CardStore manages the file-based storage of flashcards
type CardStore struct {
	RootDir  string                // Root directory for all decks
	Cards    map[string]*card.Card // Map of filepath to Card
	Decks    map[string]*deck.Deck // Map of directory path to Deck
	RootDeck *deck.Deck            // The root deck (representing RootDir)
	logger   *io.Logger            // Logger for file operations
}

// FileWatcher holds the current file watcher instance
var watcher *io.FileWatcher

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

	// Create a logger for storage operations
	logger := io.NewLogger(os.Stdout, io.INFO)

	store := &CardStore{
		RootDir: absRootDir,
		Cards:   make(map[string]*card.Card),
		Decks:   make(map[string]*deck.Deck),
		logger:  logger,
	}

	// Create the root deck
	store.RootDeck = deck.NewDeck(absRootDir, nil)
	store.Decks[absRootDir] = store.RootDeck

	// Load all cards and organize into deck structure
	if err := store.LoadAllCards(); err != nil {
		return nil, err
	}

	// Start watching for file changes
	if err := store.WatchForChanges(); err != nil {
		// Just log the error but continue - file watching is not critical
		logger.Warn("Failed to start file watcher: %v", err)
	}

	return store, nil
}

// SetLogLevel sets the minimum log level for this store
func (s *CardStore) SetLogLevel(level io.LogLevel) {
	s.logger.SetLevel(level)
}

// DisableLogging disables logging for this store
func (s *CardStore) DisableLogging() {
	s.logger.SetEnabled(false)
}

// EnableLogging enables logging for this store
func (s *CardStore) EnableLogging() {
	s.logger.SetEnabled(true)
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
func (s *CardStore) WatchForChanges() error {
	// If we already have a watcher, stop it first
	if watcher != nil {
		if err := watcher.Stop(); err != nil {
			s.logger.Warn("Failed to stop existing watcher: %v", err)
		}
		watcher = nil
	}

	// Make sure the root directory exists
	if _, err := os.Stat(s.RootDir); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", s.RootDir)
	}

	// Create a new file watcher for the root directory
	var err error
	watcher, err = io.NewFileWatcher(s.RootDir)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Start the watcher
	if err := watcher.Start(); err != nil {
		watcher = nil
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	// Start a goroutine to process events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events():
				if !ok {
					// Channel was closed, exit the goroutine
					return
				}
				// Handle the event in a separate goroutine to avoid blocking
				go s.handleFileEvent(event)
			case err, ok := <-watcher.Errors():
				if !ok {
					// Channel was closed, exit the goroutine
					return
				}
				s.logger.Error("File watcher error: %v", err)
			}
		}
	}()

	s.logger.Info("File watcher started for: %s", s.RootDir)
	return nil
}

// handleFileEvent processes file events from the watcher
func (s *CardStore) handleFileEvent(event io.FileEvent) {
	// Add some delay to allow file operations to complete
	time.Sleep(100 * time.Millisecond)

	// Check if this is a directory operation
	fi, err := os.Stat(event.Path)
	if err != nil {
		if event.Operation != "remove" {
			// If it's not a remove operation and we can't stat the file, just log and return
			s.logger.Debug("Cannot stat path %s: %v", event.Path, err)
			return
		}
		// For remove operations, we can continue even if we can't stat the file
	} else if fi.IsDir() {
		s.handleDirectoryEvent(event)
		return
	}

	// For files, only process markdown files
	if !strings.HasSuffix(event.Path, ".md") {
		return
	}

	// Skip files that might be temporary or in the middle of being edited
	// Many text editors create temporary files with patterns like .~, .swp, etc.
	baseName := filepath.Base(event.Path)
	if strings.HasPrefix(baseName, ".") || strings.HasPrefix(baseName, "~") ||
		strings.HasSuffix(baseName, ".tmp") || strings.HasSuffix(baseName, ".swp") {
		return
	}

	// Get the directory path for this file
	dirPath := filepath.Dir(event.Path)

	switch event.Operation {
	case "create", "write":
		// File was created or modified - reload it
		s.logger.Debug("Card file changed: %s", event.Path)

		// Retry loading the card a few times in case the file is still being written
		var card *card.Card
		var loadErr error
		for attempts := 0; attempts < 3; attempts++ {
			card, loadErr = s.LoadCard(event.Path)
			if loadErr == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if loadErr != nil {
			s.logger.Error("Error loading card after retries: %v", loadErr)
			return
		}

		// Update our maps
		s.Cards[event.Path] = card

		// Update deck organization if necessary
		deckObj, exists := s.Decks[dirPath]
		if !exists {
			s.logger.Warn("Deck not found for directory: %s", dirPath)
			return
		}

		// Check if this card is already in the deck
		found := false
		for i, c := range deckObj.Cards {
			if c.FilePath == event.Path {
				// Replace the existing card instead of adding a new one
				deckObj.Cards[i] = card
				found = true
				break
			}
		}
		if !found {
			deckObj.AddCard(card)
		}

	case "remove":
		// File was deleted - remove from our maps
		if cardObj, exists := s.Cards[event.Path]; exists {
			s.logger.Debug("Card file deleted: %s", event.Path)

			// Remove from the appropriate deck
			deckObj, deckExists := s.Decks[dirPath]
			if deckExists {
				deckObj.RemoveCard(cardObj)
			}

			delete(s.Cards, event.Path)
		}

	case "rename":
		// For renames, we need to rely on the create/delete events
		// fsnotify will generate a remove for the old path and a create for the new path
		s.logger.Debug("Card file renamed (processing as remove): %s", event.Path)
	}
}

// handleDirectoryEvent handles events for directories (deck operations)
func (s *CardStore) handleDirectoryEvent(event io.FileEvent) {
	// Check if this is a directory
	fi, err := os.Stat(event.Path)
	isDir := err == nil && fi.IsDir()

	switch event.Operation {
	case "create":
		if !isDir {
			return // Not a directory
		}

		s.logger.Debug("New directory detected: %s", event.Path)

		// Find the parent deck
		parentPath := filepath.Dir(event.Path)
		parentDeck, exists := s.Decks[parentPath]
		if !exists {
			s.logger.Warn("Parent deck not found for %s", event.Path)
			return
		}

		// Create a new deck for this directory
		newDeck := deck.NewDeck(event.Path, parentDeck)
		s.Decks[event.Path] = newDeck

		// Add as subdeck to parent
		parentDeck.AddSubDeck(newDeck)

	case "remove":
		// Check if this was a known deck directory
		deckObj, exists := s.Decks[event.Path]
		if !exists {
			return // Not a known deck
		}

		s.logger.Debug("Deck directory removed: %s", event.Path)

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
		delete(s.Decks, event.Path)
	}
}

// StopWatching stops the file watcher
func (s *CardStore) StopWatching() error {
	if watcher != nil {
		err := watcher.Stop()
		watcher = nil
		return err
	}
	return nil
}

// Close cleans up resources used by the CardStore
func (s *CardStore) Close() error {
	return s.StopWatching()
}
