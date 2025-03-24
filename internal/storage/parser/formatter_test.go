// File: internal/storage/parser/formatter_test.go

package parser

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
)

func TestFormatCardAsMarkdown(t *testing.T) {
	// Create a test card with predictable time values
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	testCard := &card.Card{
		Title:          "Test Card",
		Tags:           []string{"test", "markdown"},
		Created:        now,
		LastReviewed:   now.Add(-24 * time.Hour),
		ReviewInterval: 3,
		Difficulty:     4,
		Question:       "Is this a test?",
		Answer:         "Yes, it is.",
		FilePath:       "/path/to/test-card.md",
	}

	// Format the card
	content, err := FormatCardAsMarkdown(testCard)
	if err != nil {
		t.Fatalf("FormatCardAsMarkdown returned error: %v", err)
	}

	// Check for required components in the output
	expectedComponents := []string{
		"---",                // YAML frontmatter delimiter
		"tags:",              // Tags section
		"- test",             // First tag
		"- markdown",         // Second tag
		"created:",           // Created timestamp
		"last_reviewed:",     // Last reviewed timestamp
		"review_interval: 3", // Review interval
		"difficulty: 4",      // Difficulty
		"---",                // End of frontmatter
		"# Test Card",        // Title
		"## Question",        // Question section
		"Is this a test?",    // Question content
		"## Answer",          // Answer section
		"Yes, it is.",        // Answer content
	}

	for _, component := range expectedComponents {
		if !bytes.Contains(content, []byte(component)) {
			t.Errorf("Expected output to contain %q, but it doesn't", component)
		}
	}
}

func TestValidateCardContent(t *testing.T) {
	testCases := []struct {
		name        string
		card        *card.Card
		expectError bool
	}{
		{
			name: "Valid card",
			card: &card.Card{
				Title:    "Valid Card",
				Question: "Is this valid?",
				Answer:   "Yes, it is.",
			},
			expectError: false,
		},
		{
			name: "Missing title",
			card: &card.Card{
				Title:    "",
				Question: "Is this valid?",
				Answer:   "No, it's missing a title.",
			},
			expectError: true,
		},
		{
			name: "Missing question",
			card: &card.Card{
				Title:    "Invalid Card",
				Question: "",
				Answer:   "No question provided.",
			},
			expectError: true,
		},
		{
			name: "Missing answer",
			card: &card.Card{
				Title:    "Invalid Card",
				Question: "Is this valid?",
				Answer:   "",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCardContent(tc.card)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestExtractCardMetadata(t *testing.T) {
	// Create a test card with known values
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	testCard := &card.Card{
		Title:          "Metadata Test",
		Tags:           []string{"meta", "data"},
		Created:        now,
		LastReviewed:   now,
		ReviewInterval: 5,
		Difficulty:     3,
		Question:       "Question content",
		Answer:         "Answer content",
		FilePath:       "/path/to/metadata-test.md",
	}

	// Extract metadata
	metadata := ExtractCardMetadata(testCard)

	// Check metadata values
	expectedChecks := map[string]interface{}{
		"title":           "Metadata Test",
		"tags":            []string{"meta", "data"},
		"created":         now,
		"last_reviewed":   now,
		"review_interval": 5,
		"difficulty":      3,
		"file_path":       "/path/to/metadata-test.md",
	}

	for key, expected := range expectedChecks {
		value, exists := metadata[key]
		if !exists {
			t.Errorf("Expected metadata to contain key %q, but it doesn't", key)
			continue
		}

		if !reflect.DeepEqual(value, expected) {
			t.Errorf("For key %q, expected %v, got %v", key, expected, value)
		}
	}
}
