// File: internal/storage/store.go
package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/card"
	"gopkg.in/yaml.v3"
)

// CardStore manages the file-based storage of flashcards
type CardStore struct {
	RootDir string
	Cards   map[string]*card.Card // Map of filepath to Card
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
		Cards:   make(map[string]*card.Card),
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
		cardObj, err := s.LoadCard(path)
		if err != nil {
			return fmt.Errorf("failed to load card %s: %w", path, err)
		}

		s.Cards[path] = cardObj
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
	return nil
}

// DeleteCard removes a card from the filesystem and from our map
func (s *CardStore) DeleteCard(cardObj *card.Card) error {
	if err := os.Remove(cardObj.FilePath); err != nil {
		return err
	}

	delete(s.Cards, cardObj.FilePath)
	return nil
}

// parseMarkdown parses a markdown file into a Card structure
func parseMarkdown(content []byte) (*card.Card, error) {
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
	if err := yaml.Unmarshal([]byte(frontmatter), &fmData); err != nil {
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

	// Extract title, question, and answer from markdown
	// This is a simplified implementation - in practice you'd use a proper markdown parser
	lines := strings.Split(markdown, "\n")
	var inQuestion, inAnswer bool
	var questionLines, answerLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			cardObj.Title = strings.TrimPrefix(line, "# ")
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

	cardObj.Question = strings.TrimSpace(strings.Join(questionLines, "\n"))
	cardObj.Answer = strings.TrimSpace(strings.Join(answerLines, "\n"))

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
	cardObj := &card.Card{
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
	cardObj.Created = time.Now()

	// Save the card to disk
	if err := s.SaveCard(cardObj); err != nil {
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
