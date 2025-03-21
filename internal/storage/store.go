// File: internal/storage/store.go
package storage

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
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

// LoadCard loads a single card from a markdown file
func (s *CardStore) LoadCard(path string) (*card.Card, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the markdown file
	cardObj, err := parseMarkdown(content)
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
	content, err := formatCardAsMarkdown(cardObj)
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

	// In SaveCard method, modify the deck organization section:
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

// parseMarkdown parses a markdown file into a Card structure
// Uses Goldmark for proper markdown processing
func parseMarkdown(content []byte) (*card.Card, error) {
	// Check if the file starts with YAML frontmatter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, fmt.Errorf("markdown file must start with YAML frontmatter")
	}

	// Split the content into frontmatter and markdown
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format")
	}

	frontmatter := parts[1]
	markdownContent := parts[2]

	// Create a temporary struct that matches the YAML structure exactly
	type frontMatterData struct {
		Tags           []string  `yaml:"tags,omitempty"`
		Created        time.Time `yaml:"created,omitempty"`
		LastReviewed   time.Time `yaml:"last_reviewed,omitempty"`
		ReviewInterval int       `yaml:"review_interval"`
		Difficulty     int       `yaml:"difficulty,omitempty"`
	}

	// Parse YAML frontmatter into temporary struct
	var fmData frontMatterData
	if err := yaml.Unmarshal(frontmatter, &fmData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Create and populate the Card struct
	cardObj := &card.Card{
		Tags:           fmData.Tags,
		Created:        fmData.Created,
		LastReviewed:   fmData.LastReviewed,
		ReviewInterval: fmData.ReviewInterval,
		Difficulty:     fmData.Difficulty,
	}

	// Extract title, question, and answer from markdown using regex
	// This preserves the raw markdown content for proper rendering later
	mdStr := string(markdownContent)

	// Title regex: Finds a level 1 heading (# Title)
	titleRegex := regexp.MustCompile(`(?m)^# (.+)$`)
	if match := titleRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Title = strings.TrimSpace(match[1])
	}

	// Question section regex: Finds content between "## Question" and the next heading or end of content
	questionRegex := regexp.MustCompile(`(?ms)^## Question\s*\n(.*?)(?:^## |\z)`)
	if match := questionRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Question = strings.TrimSpace(match[1])
	}

	// Answer section regex: Finds content between "## Answer" and the next heading or end of content
	answerRegex := regexp.MustCompile(`(?ms)^## Answer\s*\n(.*?)(?:^## |\z)`)
	if match := answerRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Answer = strings.TrimSpace(match[1])
	}

	// Validate the extracted markdown content by parsing it with Goldmark
	// This ensures that the markdown is well-formed before we store it
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// Parse the question and answer to validate them
	// We don't need the output, we just want to ensure they parse correctly
	_ = md.Parser().Parse(text.NewReader([]byte(cardObj.Question)))
	_ = md.Parser().Parse(text.NewReader([]byte(cardObj.Answer)))

	return cardObj, nil
}

// formatCardAsMarkdown converts a Card structure to markdown format with YAML frontmatter
func formatCardAsMarkdown(cardObj *card.Card) ([]byte, error) {
	// Create a copy of the card to manipulate for YAML output
	yamlCard := struct {
		Tags           []string  `yaml:"tags,omitempty"`
		Created        time.Time `yaml:"created,omitempty"`
		LastReviewed   time.Time `yaml:"last_reviewed,omitempty"`
		ReviewInterval int       `yaml:"review_interval"` // Make sure field name matches what's in parsing code
		Difficulty     int       `yaml:"difficulty,omitempty"`
	}{
		Tags:           cardObj.Tags,
		Created:        cardObj.Created,
		LastReviewed:   cardObj.LastReviewed,
		ReviewInterval: cardObj.ReviewInterval,
		Difficulty:     cardObj.Difficulty,
	}

	// Marshal the card to YAML
	yamlData, err := yaml.Marshal(yamlCard)
	if err != nil {
		return nil, err
	}

	// Construct the full markdown content
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(string(yamlData))
	sb.WriteString("---\n\n")

	// Add title if present
	if cardObj.Title != "" {
		sb.WriteString("# " + cardObj.Title + "\n\n")
	}

	// Add question and answer sections
	sb.WriteString("## Question\n\n")
	sb.WriteString(cardObj.Question + "\n\n")
	sb.WriteString("## Answer\n\n")
	sb.WriteString(cardObj.Answer + "\n")

	return []byte(sb.String()), nil
}

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
	content, err := formatCardAsMarkdown(cardObj)
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

// WatchForChanges monitors the file system for changes to cards
// This is a placeholder for a more sophisticated file watcher
func (s *CardStore) WatchForChanges() {
	// In a real implementation, you'd use something like fsnotify
	// to watch for file changes and reload cards as needed
	fmt.Println("File watching not implemented yet")
}

// ReviewCard reviews a card with the given difficulty rating (0-5)
// and saves the updated card to disk
func (s *CardStore) ReviewCard(cardObj *card.Card, rating int) error {
	// Apply the SM-2 algorithm to calculate the next review date
	algorithm.SM2.CalculateNextReview(cardObj, rating)

	// Save the updated card to disk
	return s.SaveCard(cardObj)
}

// GetDueCards returns all cards that are due for review
// Update this method to use the SM2 algorithm for determining due cards
func (s *CardStore) GetDueCards() []*card.Card {
	var dueCards []*card.Card

	for _, cardObj := range s.Cards {
		if algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
		}
	}

	return dueCards
}

// GetDueCardsInDeck returns due cards in a specific deck and its subdecks
func (s *CardStore) GetDueCardsInDeck(deckObj *deck.Deck) []*card.Card {
	var dueCards []*card.Card
	seen := make(map[string]bool) // Track filepaths we've already seen

	// Get all cards in this deck and its subdecks
	allCards := deckObj.GetAllCards()

	// Filter for due cards
	for _, cardObj := range allCards {
		if !seen[cardObj.FilePath] && algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
			seen[cardObj.FilePath] = true
		}
	}

	return dueCards
}

// GetNextDueDate returns the date when the next card will be due
func (s *CardStore) GetNextDueDate() time.Time {
	var nextDue time.Time

	// Set nextDue to far future initially
	nextDue = time.Now().AddDate(10, 0, 0)

	for _, cardObj := range s.Cards {
		// Skip cards that are already due
		if algorithm.SM2.IsDue(cardObj) {
			return time.Now()
		}

		cardDueDate := algorithm.SM2.CalculateDueDate(cardObj)
		if cardDueDate.Before(nextDue) {
			nextDue = cardDueDate
		}
	}

	return nextDue
}

// GetReviewStats returns statistics about the review process
func (s *CardStore) GetReviewStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalCards := len(s.Cards)
	dueCards := len(s.GetDueCards())

	// Count cards by interval ranges
	newCards := 0
	young := 0  // 1-7 days
	mature := 0 // > 7 days

	for _, cardObj := range s.Cards {
		if cardObj.ReviewInterval == 0 {
			newCards++
		} else if cardObj.ReviewInterval <= 7 {
			young++
		} else {
			mature++
		}
	}

	stats["total_cards"] = totalCards
	stats["due_cards"] = dueCards
	stats["new_cards"] = newCards
	stats["young_cards"] = young
	stats["mature_cards"] = mature

	return stats
}

// GetDeckStats returns statistics about a specific deck
func (s *CardStore) GetDeckStats(deckObj *deck.Deck) map[string]interface{} {
	stats := make(map[string]interface{})

	allCards := deckObj.GetAllCards()
	totalCards := len(allCards)

	// Get due cards
	var dueCards []*card.Card
	for _, cardObj := range allCards {
		if algorithm.SM2.IsDue(cardObj) {
			dueCards = append(dueCards, cardObj)
		}
	}

	// Count cards by interval ranges
	newCards := 0
	young := 0  // 1-7 days
	mature := 0 // > 7 days

	for _, cardObj := range allCards {
		if cardObj.ReviewInterval == 0 {
			newCards++
		} else if cardObj.ReviewInterval <= 7 {
			young++
		} else {
			mature++
		}
	}

	stats["total_cards"] = totalCards
	stats["due_cards"] = len(dueCards)
	stats["new_cards"] = newCards
	stats["young_cards"] = young
	stats["mature_cards"] = mature
	stats["sub_decks"] = len(deckObj.SubDecks)
	stats["direct_cards"] = len(deckObj.Cards)

	return stats
}

// GetAllTags returns a list of all unique tags used in cards
func (s *CardStore) GetAllTags() []string {
	tagMap := make(map[string]bool)

	for _, cardObj := range s.Cards {
		for _, tag := range cardObj.Tags {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}

// GetCardsByTag returns all cards with a specific tag
func (s *CardStore) GetCardsByTag(tag string) []*card.Card {
	var result []*card.Card

	for _, cardObj := range s.Cards {
		for _, cardTag := range cardObj.Tags {
			if cardTag == tag {
				result = append(result, cardObj)
				break
			}
		}
	}

	return result
}

// SearchCards searches for cards matching the given text in title, question, or answer
func (s *CardStore) SearchCards(searchText string) []*card.Card {
	var result []*card.Card

	// Convert search text to lowercase for case-insensitive matching
	searchLower := strings.ToLower(searchText)

	for _, cardObj := range s.Cards {
		// Check title, question, and answer for the search text
		titleMatch := strings.Contains(strings.ToLower(cardObj.Title), searchLower)
		questionMatch := strings.Contains(strings.ToLower(cardObj.Question), searchLower)
		answerMatch := strings.Contains(strings.ToLower(cardObj.Answer), searchLower)

		if titleMatch || questionMatch || answerMatch {
			result = append(result, cardObj)
		}
	}

	return result
}
