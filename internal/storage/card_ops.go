// File: internal/storage/card_ops.go

// Package storage implements the file-based storage system for GoCard.
// This file contains operations related to managing individual flashcards.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage/parser"
)

// CreateCard creates a new card with the given title, question, and answer
func (s *CardStore) CreateCard(title, question, answer string, tags []string) (*card.Card, error) {
	return s.CreateCardInDeck(title, question, answer, tags, s.RootDeck)
}

// CreateCardInDeck creates a new card in the specified deck
func (s *CardStore) CreateCardInDeck(title, question, answer string, tags []string, deckObj *deck.Deck) (*card.Card, error) {
	cardObj := &card.Card{
		Title:          title,
		Tags:           tags,
		Created:        time.Now(), // Set created time to now immediately
		LastReviewed:   time.Time{},
		ReviewInterval: 0,
		Difficulty:     0,
		Question:       question,
		Answer:         answer,
	}

	// Create a filename from the title or use a timestamp if no title
	filename := "card_" + time.Now().Format("20060102_150405") + ".md"
	if cardObj.Title != "" {
		// Convert title to a filename-friendly format
		filename = strings.ToLower(cardObj.Title)
		filename = strings.ReplaceAll(filename, " ", "-")
		filename = strings.ReplaceAll(filename, "/", "-")
		filename += ".md"
	}

	// Create the filepath within the deck directory
	cardObj.FilePath = filepath.Join(deckObj.Path, filename)

	// Format the card content - do this outside the lock
	content, err := parser.FormatCardAsMarkdown(cardObj)
	if err != nil {
		return nil, err
	}

	// Create the directory if needed - filesystem operation outside lock
	dir := filepath.Dir(cardObj.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Lock before modifying shared data structures
	s.cardsMu.Lock()
	// Add to Cards map
	s.Cards[cardObj.FilePath] = cardObj
	s.cardsMu.Unlock()

	// Add to deck - this method is already synchronized
	deckObj.AddCard(cardObj)

	// Write to file - filesystem operation outside lock
	if err := os.WriteFile(cardObj.FilePath, content, 0644); err != nil {
		// If file write fails, we should clean up
		s.cardsMu.Lock()
		delete(s.Cards, cardObj.FilePath)
		s.cardsMu.Unlock()
		deckObj.RemoveCard(cardObj)
		return nil, err
	}

	return cardObj, nil
}

// LoadCard loads a single card from a markdown file
func (s *CardStore) LoadCard(path string) (*card.Card, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the markdown file
	cardObj, err := parser.ParseMarkdown(content)
	if err != nil {
		return nil, err
	}

	cardObj.FilePath = path
	return cardObj, nil
}

// SaveCard writes a card to its file
func (s *CardStore) SaveCard(cardObj *card.Card) error {
	// If the card is new and doesn't have a filepath, create one
	if cardObj.FilePath == "" {
		// Create a filename from the title or use a timestamp if no title
		filename := "card_" + time.Now().Format("20060102_150405") + ".md"
		if cardObj.Title != "" {
			// Convert title to a filename-friendly format
			filename = strings.ToLower(cardObj.Title)
			filename = strings.ReplaceAll(filename, " ", "-")
			filename = strings.ReplaceAll(filename, "/", "-")
			filename += ".md"
		}

		// Create the filepath within the root directory
		cardObj.FilePath = filepath.Join(s.RootDir, filename)
	}

	// Format the card as markdown
	content, err := parser.FormatCardAsMarkdown(cardObj)
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(cardObj.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(cardObj.FilePath, content, 0644); err != nil {
		return err
	}

	// Update our map - thread-safe
	s.cardsMu.Lock()
	s.Cards[cardObj.FilePath] = cardObj
	s.cardsMu.Unlock()

	// Update deck organization if necessary
	dirPath := filepath.Dir(cardObj.FilePath)

	s.decksMu.RLock()
	deckObj, exists := s.Decks[dirPath]
	s.decksMu.RUnlock()

	if exists {
		// Remove card if it exists then add it back (both operations are thread-safe)
		deckObj.RemoveCard(cardObj)
		deckObj.AddCard(cardObj)
	}

	return nil
}

// DeleteCard removes a card from the filesystem and from our map
func (s *CardStore) DeleteCard(cardObj *card.Card) error {
	// Capture filepath before any potential modification
	filePath := cardObj.FilePath

	if err := os.Remove(filePath); err != nil {
		return err
	}

	// Remove from the appropriate deck
	dirPath := filepath.Dir(filePath)

	s.decksMu.RLock()
	deckObj, exists := s.Decks[dirPath]
	s.decksMu.RUnlock()

	if exists {
		deckObj.RemoveCard(cardObj)
	}

	s.cardsMu.Lock()
	delete(s.Cards, filePath)
	s.cardsMu.Unlock()

	return nil
}

// MoveCard moves a card from one deck to another
func (s *CardStore) MoveCard(cardObj *card.Card, targetDeck *deck.Deck) error {
	// Capture filepath before any potential modification to avoid race conditions
	oldFilePath := cardObj.FilePath

	// Get the current deck
	currentDirPath := filepath.Dir(oldFilePath)

	s.decksMu.RLock()
	currentDeck, exists := s.Decks[currentDirPath]
	s.decksMu.RUnlock()

	if !exists {
		return fmt.Errorf("source deck not found for card: %s", oldFilePath)
	}

	// Don't do anything if the card is already in the target deck
	if currentDeck == targetDeck {
		return nil
	}

	// Calculate the new file path
	fileName := filepath.Base(oldFilePath)
	newFilePath := filepath.Join(targetDeck.Path, fileName)

	// Check if a card with the same filename already exists in the target deck
	if _, err := os.Stat(newFilePath); err == nil {
		return fmt.Errorf("a card with the same filename already exists in the target deck")
	}

	// Create a new card with updated filepath
	newCard := &card.Card{
		Title:          cardObj.Title,
		Tags:           cardObj.Tags,
		Created:        cardObj.Created,
		LastReviewed:   cardObj.LastReviewed,
		ReviewInterval: cardObj.ReviewInterval,
		Difficulty:     cardObj.Difficulty,
		Question:       cardObj.Question,
		Answer:         cardObj.Answer,
		FilePath:       newFilePath,
	}

	// Move the file to the target location
	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		return fmt.Errorf("failed to move card file: %w", err)
	}

	// Update our maps (first add new, then remove old)
	s.cardsMu.Lock()
	s.Cards[newFilePath] = newCard
	delete(s.Cards, oldFilePath)
	s.cardsMu.Unlock()

	// Add to new deck first, then remove from old deck
	targetDeck.AddCard(newCard)

	// Remove the original card from the source deck
	// We don't modify the original card at all
	currentDeck.RemoveCard(cardObj)

	// For tests that expect the filepath to be updated, return a mutex-protected reference
	// to the card from the CardStore instead of using the original card
	// This can be done by exposing a GetCardByPath method that uses the existing thread-safe
	// getCard method.

	return nil
}

// FormatCardAsMarkdown is a helper method that proxies to the parser package
// This maintains compatibility with existing tests
func (s *CardStore) FormatCardAsMarkdown(cardObj *card.Card) ([]byte, error) {
	return parser.FormatCardAsMarkdown(cardObj)
}
