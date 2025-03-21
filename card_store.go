// Filename: card_store.go
// Version: 0.0.0
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Card represents a flashcard with its metadata and content
type Card struct {
	FilePath       string    `yaml:"-"`               // Not stored in YAML, just for reference
	Title          string    `yaml:"-"`               // Extracted from markdown content
	Tags           []string  `yaml:"tags"`            // Tags for categorization
	Created        time.Time `yaml:"created"`         // Creation date
	LastReviewed   time.Time `yaml:"last_reviewed"`   // Last review date
	ReviewInterval int       `yaml:"review_interval"` // Days until next review
	Difficulty     int       `yaml:"difficulty"`      // 0-5 difficulty rating
	Question       string    `yaml:"-"`               // Extracted from markdown content
	Answer         string    `yaml:"-"`               // Extracted from markdown content
}

// CardStore manages the file-based storage of flashcards
type CardStore struct {
	RootDir string
	Cards   map[string]*Card // Map of filepath to Card
}

// NewCardStore creates a new CardStore with the given root directory
func NewCardStore(rootDir string) (*CardStore, error) {
	// Ensure the directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		if err := os.MkdirAll(rootDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	store := &CardStore{
		RootDir: rootDir,
		Cards:   make(map[string]*Card),
	}

	// Load all cards from the directory
	if err := store.LoadAllCards(); err != nil {
		return nil, err
	}

	return store, nil
}

// LoadAllCards scans the root directory and loads all markdown files as cards
func (s *CardStore) LoadAllCards() error {
	return filepath.WalkDir(s.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Load the card
		card, err := s.LoadCard(path)
		if err != nil {
			return fmt.Errorf("failed to load card %s: %w", path, err)
		}

		s.Cards[path] = card
		return nil
	})
}

// LoadCard loads a single card from a markdown file
func (s *CardStore) LoadCard(path string) (*Card, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the markdown file
	card, err := parseMarkdown(content)
	if err != nil {
		return nil, err
	}

	card.FilePath = path
	return card, nil
}

// SaveCard writes a card to its file
func (s *CardStore) SaveCard(card *Card) error {
	// If the card is new and doesn't have a filepath, create one
	if card.FilePath == "" {
		// Create a filename from the title or use a timestamp if no title
		filename := "card_" + time.Now().Format("20060102_150405") + ".md"
		if card.Title != "" {
			// Convert title to a filename-friendly format
			filename = strings.ToLower(card.Title)
			filename = strings.ReplaceAll(filename, " ", "-")
			filename = strings.ReplaceAll(filename, "/", "-")
			filename += ".md"
		}

		// Create the filepath within the root directory
		card.FilePath = filepath.Join(s.RootDir, filename)
	}

	// Format the card as markdown
	content, err := formatCardAsMarkdown(card)
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(card.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(card.FilePath, content, 0644); err != nil {
		return err
	}

	// Update our map
	s.Cards[card.FilePath] = card
	return nil
}

// DeleteCard removes a card from the filesystem and from our map
func (s *CardStore) DeleteCard(card *Card) error {
	if err := os.Remove(card.FilePath); err != nil {
		return err
	}

	delete(s.Cards, card.FilePath)
	return nil
}

// parseMarkdown parses a markdown file into a Card structure
func parseMarkdown(content []byte) (*Card, error) {
	// Check if the file starts with YAML frontmatter
	if !strings.HasPrefix(string(content), "---\n") {
		return nil, fmt.Errorf("markdown file must start with YAML frontmatter")
	}

	// Split the content into frontmatter and markdown
	parts := strings.SplitN(string(content), "---\n", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format")
	}

	frontmatter := parts[1]
	markdown := parts[2]

	// Parse YAML frontmatter
	card := &Card{}
	if err := yaml.Unmarshal([]byte(frontmatter), card); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract title, question, and answer from markdown
	// This is a simplified implementation - in practice you'd use a proper markdown parser
	lines := strings.Split(markdown, "\n")
	var inQuestion, inAnswer bool
	var questionLines, answerLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			card.Title = strings.TrimPrefix(line, "# ")
		} else if strings.HasPrefix(line, "## Question") {
			inQuestion = true
			inAnswer = false
			continue
		} else if strings.HasPrefix(line, "## Answer") {
			inQuestion = false
			inAnswer = true
			continue
		} else if strings.HasPrefix(line, "## ") {
			// Another section, stop collecting
			inQuestion = false
			inAnswer = false
		}

		if inQuestion {
			questionLines = append(questionLines, line)
		} else if inAnswer {
			answerLines = append(answerLines, line)
		}
	}

	card.Question = strings.TrimSpace(strings.Join(questionLines, "\n"))
	card.Answer = strings.TrimSpace(strings.Join(answerLines, "\n"))

	return card, nil
}

// formatCardAsMarkdown converts a Card structure to markdown format with YAML frontmatter
func formatCardAsMarkdown(card *Card) ([]byte, error) {
	// Create a copy of the card to manipulate for YAML output
	yamlCard := struct {
		Tags           []string  `yaml:"tags,omitempty"`
		Created        time.Time `yaml:"created,omitempty"`
		LastReviewed   time.Time `yaml:"last_reviewed,omitempty"`
		ReviewInterval int       `yaml:"review_interval"`
		Difficulty     int       `yaml:"difficulty,omitempty"`
	}{
		Tags:           card.Tags,
		Created:        card.Created,
		LastReviewed:   card.LastReviewed,
		ReviewInterval: card.ReviewInterval,
		Difficulty:     card.Difficulty,
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
	if card.Title != "" {
		sb.WriteString("# " + card.Title + "\n\n")
	}

	// Add question and answer sections
	sb.WriteString("## Question\n\n")
	sb.WriteString(card.Question + "\n\n")
	sb.WriteString("## Answer\n\n")
	sb.WriteString(card.Answer + "\n")

	return []byte(sb.String()), nil
}

// CreateCard creates a new card with the given title, question, and answer
func (s *CardStore) CreateCard(title, question, answer string, tags []string) (*Card, error) {
	card := &Card{
		Title:          title,
		Tags:           tags,
		Created:        time.Time{},
		LastReviewed:   time.Time{},
		ReviewInterval: 0,
		Difficulty:     0,
		Question:       question,
		Answer:         answer,
	}

	// Set created time to now
	card.Created = time.Now()

	// Save the card to disk
	if err := s.SaveCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

// WatchForChanges monitors the file system for changes to cards
// This is a placeholder for a more sophisticated file watcher
func (s *CardStore) WatchForChanges() {
	// In a real implementation, you'd use something like fsnotify
	// to watch for file changes and reload cards as needed
	fmt.Println("File watching not implemented yet")
}

// SM2 is the spaced repetition algorithm used for scheduling reviews
var SM2 = NewSM2Algorithm()

// ReviewCard reviews a card with the given difficulty rating (0-5)
// and saves the updated card to disk
func (s *CardStore) ReviewCard(card *Card, rating int) error {
	// Apply the SM-2 algorithm to calculate the next review date
	SM2.CalculateNextReview(card, rating)

	// Save the updated card to disk
	return s.SaveCard(card)
}

// GetDueCards returns all cards that are due for review
// Update this method to use the SM2 algorithm for determining due cards
func (s *CardStore) GetDueCards() []*Card {
	var dueCards []*Card

	for _, card := range s.Cards {
		if SM2.IsDue(card) {
			dueCards = append(dueCards, card)
		}
	}

	return dueCards
}

// GetNextDueDate returns the date when the next card will be due
func (s *CardStore) GetNextDueDate() time.Time {
	var nextDue time.Time

	// Set nextDue to far future initially
	nextDue = time.Now().AddDate(10, 0, 0)

	for _, card := range s.Cards {
		// Skip cards that are already due
		if SM2.IsDue(card) {
			return time.Now()
		}

		cardDueDate := SM2.CalculateDueDate(card)
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

	for _, card := range s.Cards {
		if card.ReviewInterval == 0 {
			newCards++
		} else if card.ReviewInterval <= 7 {
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
