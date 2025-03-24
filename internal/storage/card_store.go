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
	"sync"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// CardStore manages the file-based storage of flashcards
type CardStore struct {
	cardsMu   sync.RWMutex // Mutex for the Cards map
	decksMu   sync.RWMutex // Mutex for the Decks map
	watcherMu sync.Mutex   // Mutex for watcher operations

	RootDir  string
	Cards    map[string]*card.Card
	Decks    map[string]*deck.Deck
	RootDeck *deck.Deck
	logger   *io.Logger
	watcher  *io.FileWatcher
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

	// Create a logger for storage operations
	logger := io.NewLogger(os.Stdout, io.INFO)

	store := &CardStore{
		RootDir: absRootDir,
		Cards:   make(map[string]*card.Card),
		Decks:   make(map[string]*deck.Deck),
		logger:  logger,
		watcher: nil, // Initialize to nil
	}

	// Create the root deck
	store.RootDeck = deck.NewDeck(absRootDir, nil)

	store.decksMu.Lock()
	store.Decks[absRootDir] = store.RootDeck
	store.decksMu.Unlock()

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

// GetCardCount returns the number of cards in the store (thread-safe)
func (s *CardStore) GetCardCount() int {
	s.cardsMu.RLock()
	defer s.cardsMu.RUnlock()
	return len(s.Cards)
}

// GetDeckCount returns the number of decks in the store (thread-safe)
func (s *CardStore) GetDeckCount() int {
	s.decksMu.RLock()
	defer s.decksMu.RUnlock()
	return len(s.Decks)
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

		// Thread-safe updates
		s.cardsMu.Lock()
		s.Cards[path] = cardObj
		s.cardsMu.Unlock()

		// Add the card to the appropriate deck
		dirPath := filepath.Dir(path)

		s.decksMu.RLock()
		deckObj, exists := s.Decks[dirPath]
		s.decksMu.RUnlock()

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

		s.decksMu.RLock()
		parentDeck, exists := s.Decks[parentPath]
		s.decksMu.RUnlock()

		if !exists {
			return fmt.Errorf("parent deck not found for %s", path)
		}

		// Create the new deck
		newDeck := deck.NewDeck(absPath, parentDeck)

		s.decksMu.Lock()
		s.Decks[absPath] = newDeck
		s.decksMu.Unlock()

		// Add as subdeck to parent
		parentDeck.AddSubDeck(newDeck)

		return nil
	})
}

// WatchForChanges monitors the file system for changes to cards
func (s *CardStore) WatchForChanges() error {
	// Lock the watcher mutex to protect from concurrent access
	s.watcherMu.Lock()
	defer s.watcherMu.Unlock()

	// If we already have a watcher, stop it first
	if s.watcher != nil {
		if err := s.watcher.Stop(); err != nil {
			s.logger.Warn("Failed to stop existing watcher: %v", err)
		}
		s.watcher = nil
	}

	// Make sure the root directory exists
	if _, err := os.Stat(s.RootDir); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", s.RootDir)
	}

	// Create a new file watcher for the root directory
	watcher, err := io.NewFileWatcher(s.RootDir)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Set logger for the watcher
	watcher.SetLogger(s.logger)

	// Start the watcher
	if err := watcher.Start(); err != nil {
		s.watcher = nil
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	// Store the watcher in the CardStore
	s.watcher = watcher

	// Start a goroutine to process events
	// Get a local copy of watcher to avoid race conditions
	localWatcher := s.watcher
	go func() {
		for {
			select {
			case event, ok := <-localWatcher.Events():
				if !ok {
					// Channel was closed, exit the goroutine
					return
				}
				// Handle the event in a separate goroutine to avoid blocking
				go s.handleFileEvent(event)
			case err, ok := <-localWatcher.Errors():
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

	// Get the directory path for this file
	dirPath := filepath.Dir(event.Path)

	switch event.Operation {
	case "create", "write":
		// File was created or modified - reload it
		s.logger.Debug("Card file changed: %s", event.Path)

		// Retry loading the card a few times in case the file is still being written
		var cardObj *card.Card
		var loadErr error
		for attempts := 0; attempts < 3; attempts++ {
			cardObj, loadErr = s.LoadCard(event.Path)
			if loadErr == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if loadErr != nil {
			s.logger.Error("Error loading card after retries: %v", loadErr)
			return
		}

		// Thread-safe updates to maps
		s.cardsMu.Lock()
		s.Cards[event.Path] = cardObj
		s.cardsMu.Unlock()

		s.decksMu.RLock()
		deckObj, exists := s.Decks[dirPath]
		s.decksMu.RUnlock()

		if !exists {
			s.logger.Warn("Deck not found for directory: %s", dirPath)
			return
		}

		// Update the card in the deck (first remove if exists, then add)
		deckObj.RemoveCard(cardObj)
		deckObj.AddCard(cardObj)

	case "remove":
		// File was deleted - remove from our maps
		s.cardsMu.Lock()
		cardObj, exists := s.Cards[event.Path]
		if exists {
			s.logger.Debug("Card file deleted: %s", event.Path)
			delete(s.Cards, event.Path)
		}
		s.cardsMu.Unlock()

		if exists {
			s.decksMu.RLock()
			deckObj, deckExists := s.Decks[dirPath]
			s.decksMu.RUnlock()

			// Remove from deck if both card and deck exist
			if deckExists {
				deckObj.RemoveCard(cardObj)
			}
		}
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

		// Find the parent deck - thread-safe read
		parentPath := filepath.Dir(event.Path)

		s.decksMu.RLock()
		parentDeck, exists := s.Decks[parentPath]
		s.decksMu.RUnlock()

		if !exists {
			s.logger.Warn("Parent deck not found for %s", event.Path)
			return
		}

		// Create a new deck for this directory
		newDeck := deck.NewDeck(event.Path, parentDeck)

		// Thread-safe update to maps
		s.decksMu.Lock()
		s.Decks[event.Path] = newDeck
		s.decksMu.Unlock()

		// Add as subdeck to parent (thread-safe)
		parentDeck.AddSubDeck(newDeck)

	case "remove":
		// Thread-safe check if this was a known deck directory
		s.decksMu.RLock()
		deckObj, exists := s.Decks[event.Path]
		s.decksMu.RUnlock()

		if !exists {
			return // Not a known deck
		}

		s.logger.Debug("Deck directory removed: %s", event.Path)

		// Get all cards from this deck for removal (thread-safe)
		allCards := deckObj.GetAllCards()
		allDecks := deckObj.AllDecks()

		// Thread-safe updates to maps
		s.cardsMu.Lock()
		// Remove all cards in this deck and its subdecks from our maps
		for _, cardObj := range allCards {
			delete(s.Cards, cardObj.FilePath)
		}
		s.cardsMu.Unlock()

		s.decksMu.Lock()
		// Remove all subdecks from our map
		for _, subDeck := range allDecks {
			if subDeck != deckObj { // Skip the deck itself, we'll remove it separately
				delete(s.Decks, subDeck.Path)
			}
		}

		// Remove the deck from our map
		delete(s.Decks, event.Path)
		s.decksMu.Unlock()

		// Get the parent deck
		parentDeck := deckObj.ParentDeck
		if parentDeck != nil {
			// Use the public method to remove the subdeck from parent
			parentDeck.RemoveSubDeck(deckObj.Name)
		}
	}
}

// StopWatching stops the file watcher
func (s *CardStore) StopWatching() error {
	s.watcherMu.Lock()
	defer s.watcherMu.Unlock()

	if s.watcher != nil {
		err := s.watcher.Stop()
		s.watcher = nil
		return err
	}
	return nil
}

// Close cleans up resources used by the CardStore
func (s *CardStore) Close() error {
	return s.StopWatching()
}

// Helper methods to safely access maps
func (s *CardStore) getCard(path string) (*card.Card, bool) {
	s.cardsMu.RLock()
	defer s.cardsMu.RUnlock()
	card, exists := s.Cards[path]
	return card, exists
}

// GetCardByPath returns a card by its filepath (thread-safe)
func (s *CardStore) GetCardByPath(path string) (*card.Card, bool) {
	return s.getCard(path)
}

func (s *CardStore) getDeckByPath(path string) (*deck.Deck, bool) {
	s.decksMu.RLock()
	defer s.decksMu.RUnlock()
	deck, exists := s.Decks[path]
	return deck, exists
}

// removeSubDeckFromParent removes a subdeck from its parent
func (s *CardStore) removeSubDeckFromParent(parentDeck *deck.Deck, subDeckName string) {
	parentDeck.RemoveSubDeck(subDeckName)
}
