// File: internal/data/store.go

package data

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
	"github.com/DavidMiserak/GoCard/internal/srs"
)

// Store manages all data for the application
type Store struct {
	Decks []model.Deck
}

// NewStore creates a new data store with dummy data
func NewStore() *Store {
	store := &Store{
		Decks: []model.Deck{},
	}

	// Add dummy data
	store.Decks = GetDummyDecks()

	return store
}

// NewStoreFromDir creates a new data store with decks from the specified directory
func NewStoreFromDir(dirPath string) (*Store, error) {
	store := &Store{
		Decks: []model.Deck{},
	}

	// List all subdirectories (each will be a deck)
	subdirs, err := listSubdirectories(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error listing subdirectories: %w", err)
	}

	// If no subdirectories found, treat the main directory as a single deck
	if len(subdirs) == 0 {
		deck, err := CreateDeckFromDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("error creating deck from directory: %w", err)
		}
		store.Decks = append(store.Decks, *deck)
		return store, nil
	}

	// Create decks from each subdirectory
	for _, subdir := range subdirs {
		deck, err := CreateDeckFromDir(subdir)
		if err != nil {
			// Log the error but continue with other subdirectories
			fmt.Printf("Warning: Error loading deck from %s: %v\n", subdir, err)
			continue
		}
		store.Decks = append(store.Decks, *deck)
	}

	// If no decks were loaded, use dummy data
	if len(store.Decks) == 0 {
		fmt.Println("No decks found in the specified directory. Using dummy data instead.")
		store.Decks = GetDummyDecks()
	}

	return store, nil
}

// listSubdirectories lists all immediate subdirectories in the given path
func listSubdirectories(dirPath string) ([]string, error) {
	var subdirs []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subdirPath := filepath.Join(dirPath, entry.Name())
			subdirs = append(subdirs, subdirPath)
		}
	}

	return subdirs, nil
}

// GetDecks returns all decks
func (s *Store) GetDecks() []model.Deck {
	return s.Decks
}

// GetDeck returns a deck by ID
func (s *Store) GetDeck(id string) (model.Deck, bool) {
	for _, deck := range s.Decks {
		if deck.ID == id {
			return deck, true
		}
	}
	return model.Deck{}, false
}

// GetDueCards returns cards due for review
func (s *Store) GetDueCards() []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		for _, card := range deck.Cards {
			if card.NextReview.Before(now) {
				dueCards = append(dueCards, card)
			}
		}
	}

	return dueCards
}

// GetDueCardsForDeck returns cards due for review in a specific deck
func (s *Store) GetDueCardsForDeck(deckID string) []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		if deck.ID == deckID {
			for _, card := range deck.Cards {
				if card.NextReview.Before(now) {
					dueCards = append(dueCards, card)
				}
			}
			break
		}
	}

	return dueCards
}

// UpdateCard updates a card in the store and returns whether it was found
func (s *Store) UpdateCard(updatedCard model.Card) bool {
	// Find and update the card in its deck
	for i, deck := range s.Decks {
		if deck.ID == updatedCard.DeckID {
			for j, card := range deck.Cards {
				if card.ID == updatedCard.ID {
					// Update the card
					s.Decks[i].Cards[j] = updatedCard
					return true
				}
			}
		}
	}
	return false
}

// UpdateDeckLastStudied updates the LastStudied timestamp for a deck
func (s *Store) UpdateDeckLastStudied(deckID string) bool {
	for i, deck := range s.Decks {
		if deck.ID == deckID {
			s.Decks[i].LastStudied = time.Now()
			return true
		}
	}
	return false
}

// ToggleDeckAlgorithm atomically converts all cards in a deck to a different algorithm
// Returns success/failure status and count of cards converted
// On error, the deck is rolled back to its original state (in-memory backup)
func (s *Store) ToggleDeckAlgorithm(deckID, targetAlgorithm string) (success bool, cardsConverted int, err error) {
	// Find the deck
	deckIndex := -1
	for i, deck := range s.Decks {
		if deck.ID == deckID {
			deckIndex = i
			break
		}
	}

	if deckIndex < 0 {
		return false, 0, fmt.Errorf("deck with ID %s not found", deckID)
	}

	deck := &s.Decks[deckIndex]
	previousAlgorithm := deck.Algorithm
	if previousAlgorithm == "" {
		previousAlgorithm = "SM2"
	}

	// Create backup of all cards before conversion
	backup := make([]model.Card, len(deck.Cards))
	for i, card := range deck.Cards {
		backup[i] = card
	}

	// Perform the toggle operation in-memory
	toggleResult := srs.ToggleDeckAlgorithm(deck, targetAlgorithm)
	if !toggleResult.Success {
		return false, 0, toggleResult.Error
	}

	// Try to persist all cards to disk atomically
	// If any write fails, rollback
	for _, card := range deck.Cards {
		// Skip cards without proper file paths
		if card.ID == "" || (!filepath.IsAbs(card.ID) &&
			!strings.Contains(card.ID, "/") && !strings.Contains(card.ID, "\\")) {
			continue
		}

		// Try to update the card file
		if err := UpdateCardFile(card); err != nil {
			// Write failed: rollback all changes
			srs.RollbackDeckAlgorithm(deck, backup, previousAlgorithm)
			return false, 0, fmt.Errorf("failed to persist card %s during toggle: %w", card.ID, err)
		}
	}

	// All writes succeeded
	return true, toggleResult.CardsConverted, nil
}

// SaveCardReview updates a card with its new review data and updates
// the parent deck's LastStudied timestamp
func (s *Store) SaveCardReview(card model.Card, rating int) bool {
	// Use the SRS algorithm to schedule the card (works with pointer)
	if err := srs.ScheduleCard(&card, rating); err != nil {
		fmt.Printf("Warning: failed to schedule card: %v\n", err)
		return false
	}

	// Update the card in the store
	cardUpdated := s.UpdateCard(card)

	// Persist the updated card to its markdown file immediately
	if cardUpdated {
		if err := s.SaveCardToMarkdown(card); err != nil {
			fmt.Printf("Warning: failed to persist card review to file: %v\n", err)
		}
	}

	// Update the deck's last studied timestamp
	deckUpdated := s.UpdateDeckLastStudied(card.DeckID)

	return cardUpdated && deckUpdated
}

// SaveCardToMarkdown saves SRS metadata for a single card back to its markdown file.
// It performs a surgical update of only the SRS fields in the frontmatter.
func (s *Store) SaveCardToMarkdown(card model.Card) error {
	// Skip cards without a proper file path
	if card.ID == "" || (!filepath.IsAbs(card.ID) &&
		!strings.Contains(card.ID, "/") && !strings.Contains(card.ID, "\\")) {
		return nil
	}

	// Verify the file exists before updating
	if _, err := os.Stat(card.ID); os.IsNotExist(err) {
		return nil
	}

	// Read the existing file content
	content, err := os.ReadFile(card.ID)
	if err != nil {
		return fmt.Errorf("error reading card file %s: %w", card.ID, err)
	}

	// Parse the content to extract front matter
	contentStr := string(content)
	fmStart := strings.Index(contentStr, "---")
	if fmStart < 0 {
		return nil
	}

	fmEnd := strings.Index(contentStr[fmStart+3:], "---")
	if fmEnd < 0 {
		return nil
	}
	fmEnd = fmStart + 3 + fmEnd

	frontMatter := contentStr[fmStart : fmEnd+3]
	bodyContent := contentStr[fmEnd+3:]

	// Update only the SRS-specific fields in front matter
	updatedFrontMatter := updateFrontMatterFields(frontMatter, card)

	// Combine updated front matter with original body content
	updatedContent := updatedFrontMatter + bodyContent

	// Write back to file
	if err := os.WriteFile(card.ID, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("error writing updated card file %s: %w", card.ID, err)
	}

	return nil
}

// SaveDeckToMarkdown saves SRS metadata for all cards in a deck back to their markdown files
func (s *Store) SaveDeckToMarkdown(deckID string) error {
	// Get the deck from the store
	deck, found := s.GetDeck(deckID)
	if !found {
		return fmt.Errorf("deck with ID %s not found", deckID)
	}

	// Only proceed if the deck ID looks like a valid directory path
	if !filepath.IsAbs(deck.ID) && !strings.Contains(deck.ID, "/") && !strings.Contains(deck.ID, "\\") {
		return nil
	}

	for _, card := range deck.Cards {
		if err := s.SaveCardToMarkdown(card); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to update only SRS-related fields in front matter
func updateFrontMatterFields(frontMatter string, card model.Card) string {
	// Regular expressions to update specific fields
	reviewIntervalRe := regexp.MustCompile(`(review_interval:\s*)[0-9.]+`)
	difficultyRe := regexp.MustCompile(`(difficulty:\s*)[0-9.]+`)
	lastReviewedRe := regexp.MustCompile(`(last_reviewed:\s*)[^\n]+`)

	// Format date in the YYYY-MM-DD format
	lastReviewedFormatted := card.LastReviewed.Format("2006-01-02")

	// Update each field if it exists
	if reviewIntervalRe.MatchString(frontMatter) {
		frontMatter = reviewIntervalRe.ReplaceAllString(frontMatter,
			fmt.Sprintf("${1}%d", card.Interval))
	}

	if difficultyRe.MatchString(frontMatter) {
		frontMatter = difficultyRe.ReplaceAllString(frontMatter,
			fmt.Sprintf("${1}%.1f", card.Ease))
	}

	if lastReviewedRe.MatchString(frontMatter) {
		frontMatter = lastReviewedRe.ReplaceAllString(frontMatter,
			fmt.Sprintf("${1}%s", lastReviewedFormatted))
	}

	return frontMatter
}
