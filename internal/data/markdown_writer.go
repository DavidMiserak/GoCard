// File: internal/data/markdown_writer.go

package data

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
	"gopkg.in/yaml.v3"
)

// CardToMarkdown converts a model.Card to a MarkdownCard
func CardToMarkdown(card model.Card) *MarkdownCard {
	// Extract tags (if stored in the card)
	tags := []string{}

	// Create MarkdownCard
	mc := &MarkdownCard{
		Path: card.ID,
		FrontMatter: FrontMatter{
			Tags:           tags,
			Created:        time.Now(), // Default to now if not available
			LastReviewed:   card.LastReviewed,
			ReviewInterval: card.Interval,
			Difficulty:     card.Ease,
		},
		Question: card.Question,
		Answer:   card.Answer,
	}

	return mc
}

// SanitizeFilename converts a string to a Unix-friendly filename
func SanitizeFilename(name string) string {
	// First trim spaces from start and end
	name = strings.TrimSpace(name)

	// If string is empty or just dots, return default
	if name == "" || strings.Trim(name, ".") == "" {
		return "card"
	}

	// Replace all whitespace with underscores
	re := regexp.MustCompile(`\s+`)
	name = re.ReplaceAllString(name, "_")

	// Replace other problematic characters
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")

	// Trim leading/trailing dots
	name = strings.Trim(name, ".")

	// Final check for empty string
	if name == "" {
		return "card"
	}

	return name
}

// WriteMarkdownCard writes a MarkdownCard to a file
func WriteMarkdownCard(mc *MarkdownCard, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Marshall frontmatter to YAML
	frontmatterBytes, err := yaml.Marshal(mc.FrontMatter)
	if err != nil {
		return fmt.Errorf("error marshalling frontmatter: %w", err)
	}

	// Construct file content
	content := fmt.Sprintf("---\n%s---\n\n# Question\n\n%s\n\n## Answer\n\n%s\n",
		string(frontmatterBytes),
		mc.Question,
		mc.Answer)

	// Write to file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// WriteCard writes a model.Card to a markdown file
func WriteCard(card model.Card, path string) error {
	mc := CardToMarkdown(card)
	return WriteMarkdownCard(mc, path)
}

// WriteDeckToMarkdown writes all cards in a deck to markdown files
func WriteDeckToMarkdown(deck *model.Deck, dirPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write each card
	for i, card := range deck.Cards {
		// Generate filename if not available
		var filename string
		if card.ID == "" {
			// Simple numeric filename
			filename = filepath.Join(dirPath, fmt.Sprintf("card_%d.md", i+1))
		} else {
			// Use card ID but sanitize it first
			baseName := filepath.Base(card.ID)
			sanitizedName := SanitizeFilename(baseName)

			// Ensure it ends with .md
			if !strings.HasSuffix(strings.ToLower(sanitizedName), ".md") {
				sanitizedName += ".md"
			}

			filename = filepath.Join(dirPath, sanitizedName)
		}

		// Update card ID to match filename
		card.ID = filename

		// Write card
		if err := WriteCard(card, filename); err != nil {
			return fmt.Errorf("error writing card %d: %w", i, err)
		}
	}

	return nil
}

// WriteNewDeck creates a new deck directory and writes all cards as markdown files
func WriteNewDeck(deck *model.Deck) error {
	// Use deck ID as directory path
	dirPath := deck.ID
	if dirPath == "" {
		return fmt.Errorf("deck ID (directory path) is required")
	}

	return WriteDeckToMarkdown(deck, dirPath)
}

// UpdateCardFile updates an existing markdown file with modified card data
func UpdateCardFile(card model.Card) error {
	// Check if file exists
	_, err := os.Stat(card.ID)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file if it doesn't exist
			return WriteCard(card, card.ID)
		}
		return fmt.Errorf("error checking file: %w", err)
	}

	// Read existing card to preserve metadata
	existingCard, err := ParseMarkdownFile(card.ID)
	if err != nil {
		return fmt.Errorf("error reading existing card: %w", err)
	}

	// Update with new data while preserving tags and created date
	mc := CardToMarkdown(card)
	mc.FrontMatter.Tags = existingCard.FrontMatter.Tags

	// Keep original creation date if it exists
	if !existingCard.FrontMatter.Created.IsZero() {
		mc.FrontMatter.Created = existingCard.FrontMatter.Created
	}

	// Write updated card
	return WriteMarkdownCard(mc, card.ID)
}
