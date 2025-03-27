// internal/service/storage/new_format_test.go
package storage

import (
	"strings"
	"testing"
)

func TestLoadCardWithNewFormat(t *testing.T) {
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Test card with new ## Question and ## Answer format
	newFormatContent := `---
title: Spanish Greetings
tags: [spanish,vocabulary,language-learning]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 3
---
# Spanish Greetings and Introductions

## Question

What are the common Spanish greetings and introductions?

## Answer

### Formal Greetings
- Buenos días (Good morning)
- Buenas tardes (Good afternoon)
- Buenas noches (Good evening/night)

### Informal Greetings
- ¡Hola! (Hello!)
- ¿Qué tal? (How's it going?)
- ¿Cómo estás? (How are you?)
`

	// Create the card file
	cardPath, err := createSampleCardFile(tempDir, "new-format-test.md", newFormatContent)
	if err != nil {
		t.Fatalf("failed to create sample card file: %v", err)
	}

	// Load the card
	card, err := fs.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("LoadCard() error = %v", err)
	}

	// Verify card properties
	if card.Title != "Spanish Greetings" {
		t.Errorf("expected Title to be 'Spanish Greetings', got '%s'", card.Title)
	}

	// Verify tags were parsed correctly from array format
	expectedTags := []string{"spanish", "vocabulary", "language-learning"}
	if len(card.Tags) != len(expectedTags) {
		t.Errorf("expected %d tags, got %d: %v", len(expectedTags), len(card.Tags), card.Tags)
	} else {
		for i, tag := range expectedTags {
			if i < len(card.Tags) && card.Tags[i] != tag {
				t.Errorf("expected tag[%d] to be '%s', got '%s'", i, tag, card.Tags[i])
			}
		}
	}

	// Check question extraction with ## Question format
	if !contains(card.Question, "What are the common Spanish greetings and introductions?") {
		t.Errorf("Question not extracted correctly: %s", card.Question)
	}

	// Check answer extraction with ## Answer format
	if !contains(card.Answer, "Formal Greetings") || !contains(card.Answer, "Informal Greetings") {
		t.Errorf("Answer not extracted correctly: %s", card.Answer)
	}

	// Test card without title in frontmatter
	untitledContent := `---
tags: [test]
difficulty: 2
---
# This Is My Card Title

## Question
What happens when there's no title in frontmatter?

## Answer
The filename is used as the title.
`

	// Create the card file without title in frontmatter
	untitledPath, err := createSampleCardFile(tempDir, "untitled-test.md", untitledContent)
	if err != nil {
		t.Fatalf("failed to create untitled card file: %v", err)
	}

	// Load the card
	untitledCard, err := fs.LoadCard(untitledPath)
	if err != nil {
		t.Fatalf("LoadCard() error for untitled = %v", err)
	}

	// Filename should be used as title
	expectedTitle := "untitled-test"
	if untitledCard.Title != expectedTitle {
		t.Errorf("expected Title to be '%s' for untitled card, got '%s'", expectedTitle, untitledCard.Title)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
