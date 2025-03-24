// File: internal/storage/parser/formatter.go

// Package parser handles parsing and formatting of markdown files.
package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"gopkg.in/yaml.v3"
)

// FormatCardAsMarkdown converts a Card structure to markdown format with YAML frontmatter
func FormatCardAsMarkdown(cardObj *card.Card) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to marshal card to YAML: %w", err)
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

// ExtractCardMetadata extracts the metadata from a card for display
func ExtractCardMetadata(cardObj *card.Card) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["title"] = cardObj.Title
	metadata["tags"] = cardObj.Tags
	metadata["created"] = cardObj.Created
	metadata["last_reviewed"] = cardObj.LastReviewed
	metadata["review_interval"] = cardObj.ReviewInterval
	metadata["difficulty"] = cardObj.Difficulty
	metadata["file_path"] = cardObj.FilePath

	return metadata
}

// ValidateCardContent checks if a card has valid content
func ValidateCardContent(cardObj *card.Card) error {
	if cardObj.Title == "" {
		return fmt.Errorf("card must have a title")
	}

	if cardObj.Question == "" {
		return fmt.Errorf("card must have a question")
	}

	if cardObj.Answer == "" {
		return fmt.Errorf("card must have an answer")
	}

	return nil
}
