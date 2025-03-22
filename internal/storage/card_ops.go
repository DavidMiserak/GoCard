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

	// Add to Cards map first
	s.Cards[cardObj.FilePath] = cardObj

	// Add to deck directly instead of calling SaveCard
	deckObj.AddCard(cardObj)

	// Save to disk after adding to data structures
	content, err := parser.FormatCardAsMarkdown(cardObj)
	if err != nil {
		return nil, err
	}

	// Create the directory if needed
	dir := filepath.Dir(cardObj.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Write to file
	if err := os.WriteFile(cardObj.FilePath, content, 0644); err != nil {
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

	// Update our map
	s.Cards[cardObj.FilePath] = cardObj

	// Update deck organization if necessary
	dirPath := filepath.Dir(cardObj.FilePath)
	if deckObj, exists := s.Decks[dirPath]; exists {
		// Check if this card is already in the deck
		found := false
		for i, c := range deckObj.Cards {
			if c.FilePath == cardObj.FilePath {
				// Replace the existing card instead of adding a new one
				deckObj.Cards[i] = cardObj
				found = true
				break
			}
		}
		if !found {
			deckObj.AddCard(cardObj)
		}
	}

	return nil
}

// DeleteCard removes a card from the filesystem and from our map
func (s *CardStore) DeleteCard(cardObj *card.Card) error {
	if err := os.Remove(cardObj.FilePath); err != nil {
		return err
	}

	// Remove from the appropriate deck
	dirPath := filepath.Dir(cardObj.FilePath)
	if deckObj, exists := s.Decks[dirPath]; exists {
		deckObj.RemoveCard(cardObj)
	}

	delete(s.Cards, cardObj.FilePath)
	return nil
}

// MoveCard moves a card from one deck to another
func (s *CardStore) MoveCard(cardObj *card.Card, targetDeck *deck.Deck) error {
	// Get the current deck
	currentDirPath := filepath.Dir(cardObj.FilePath)
	currentDeck, exists := s.Decks[currentDirPath]
	if !exists {
		return fmt.Errorf("source deck not found for card: %s", cardObj.FilePath)
	}

	// Don't do anything if the card is already in the target deck
	if currentDeck == targetDeck {
		return nil
	}

	// Calculate the new file path
	fileName := filepath.Base(cardObj.FilePath)
	newFilePath := filepath.Join(targetDeck.Path, fileName)

	// Check if a card with the same filename already exists in the target deck
	if _, err := os.Stat(newFilePath); err == nil {
		return fmt.Errorf("a card with the same filename already exists in the target deck")
	}

	// Create the old file path before we modify the card
	oldFilePath := cardObj.FilePath

	// Move the file
	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		return fmt.Errorf("failed to move card file: %w", err)
	}

	// Update the card's filepath
	cardObj.FilePath = newFilePath

	// Update our maps
	delete(s.Cards, oldFilePath)
	s.Cards[newFilePath] = cardObj

	// Update the deck associations with some debugging
	fmt.Printf("Before removal: Current deck has %d cards\n", len(currentDeck.Cards))
	success := currentDeck.RemoveCard(cardObj)
	fmt.Printf("Removal successful: %v, Current deck now has %d cards\n", success, len(currentDeck.Cards))
	targetDeck.AddCard(cardObj)

	return nil
}

// FormatCardAsMarkdown is a helper method that proxies to the parser package
// This maintains compatibility with existing tests
func (s *CardStore) FormatCardAsMarkdown(cardObj *card.Card) ([]byte, error) {
	return parser.FormatCardAsMarkdown(cardObj)
}
