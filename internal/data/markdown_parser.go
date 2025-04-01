// File: internal/data/markdown_parser.go

package data

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
	"gopkg.in/yaml.v3"
)

// FrontMatter represents the YAML frontmatter in a markdown file
type FrontMatter struct {
	Tags           []string  `yaml:"tags"`
	Created        time.Time `yaml:"created"`
	LastReviewed   time.Time `yaml:"last_reviewed"`
	ReviewInterval int       `yaml:"review_interval"`
	Difficulty     float64   `yaml:"difficulty"`
}

// MarkdownCard represents a card in markdown format
type MarkdownCard struct {
	Path        string
	FrontMatter FrontMatter
	Question    string
	Answer      string
}

// ParseMarkdownFile parses a markdown file into a MarkdownCard
func ParseMarkdownFile(path string) (*MarkdownCard, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	card := &MarkdownCard{
		Path: path,
	}

	scanner := bufio.NewScanner(file)

	// Parse frontmatter
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("missing frontmatter start")
	}

	var frontmatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	frontmatter := strings.Join(frontmatterLines, "\n")
	if err := yaml.Unmarshal([]byte(frontmatter), &card.FrontMatter); err != nil {
		return nil, fmt.Errorf("error parsing frontmatter: %w", err)
	}

	// Parse question and answer
	section := ""
	var questionLines, answerLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "# Question") || strings.HasPrefix(line, "## Question") {
			section = "question"
			continue
		} else if strings.HasPrefix(line, "# Answer") || strings.HasPrefix(line, "## Answer") {
			section = "answer"
			continue
		}

		switch section {
		case "question":
			questionLines = append(questionLines, line)
		case "answer":
			answerLines = append(answerLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	card.Question = strings.TrimSpace(strings.Join(questionLines, "\n"))
	card.Answer = strings.TrimSpace(strings.Join(answerLines, "\n"))

	return card, nil
}

// ToModelCard converts a MarkdownCard to a model.Card
func (mc *MarkdownCard) ToModelCard(deckID string) model.Card {
	// Set sensible defaults
	now := time.Now()
	lastReviewed := mc.FrontMatter.LastReviewed
	if lastReviewed.IsZero() {
		lastReviewed = now
	}

	// Calculate next review based on interval
	interval := mc.FrontMatter.ReviewInterval
	nextReview := lastReviewed.AddDate(0, 0, interval)

	// Default ease value if not specified
	ease := mc.FrontMatter.Difficulty
	if ease == 0 {
		ease = 2.5 // Default difficulty value
	}

	return model.Card{
		ID:           mc.Path,
		Question:     mc.Question,
		Answer:       mc.Answer,
		DeckID:       deckID,
		LastReviewed: lastReviewed,
		NextReview:   nextReview,
		Ease:         ease,
		Interval:     interval,
		Rating:       0, // Default to 0 for new cards
	}
}

// ScanDirForMarkdown scans a directory for markdown files
func ScanDirForMarkdown(dirPath string) ([]string, error) {
	var mdFiles []string

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != dirPath {
			return filepath.SkipDir // Skip subdirectories
		}

		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			mdFiles = append(mdFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return mdFiles, nil
}

// ImportMarkdownToDeck imports markdown files into an existing deck
func ImportMarkdownToDeck(dirPath string, deck *model.Deck) error {
	mdFiles, err := ScanDirForMarkdown(dirPath)
	if err != nil {
		return err
	}

	for _, path := range mdFiles {
		card, err := ParseMarkdownFile(path)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", path, err)
		}

		modelCard := card.ToModelCard(deck.ID)
		deck.Cards = append(deck.Cards, modelCard)
	}

	return nil
}

// CreateDeckFromDir creates a new deck from a directory of markdown files
func CreateDeckFromDir(dirPath string) (*model.Deck, error) {
	// Create a new deck
	deckInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing deck directory: %w", err)
	}

	if !deckInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	deck := &model.Deck{
		ID:          dirPath,
		Name:        filepath.Base(dirPath),
		CreatedAt:   time.Now(),
		LastStudied: time.Now(),
		Cards:       []model.Card{},
	}

	// Import markdown files
	if err := ImportMarkdownToDeck(dirPath, deck); err != nil {
		return nil, err
	}

	return deck, nil
}
