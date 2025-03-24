// File: internal/storage/card_ops_metadata_test.go

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/storage/parser"
)

// TestCardOperationsCorrectPersistence tests that card operations correctly persist and load data
func TestCardOperationsCorrectPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-card-ops-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a card store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}
	defer store.Close()

	// Test Case 1: Create a card and verify it's saved to disk
	t.Run("CreateCardPersistence", func(t *testing.T) {
		title := "Test Card Creation"
		question := "Is this card created and persisted correctly?"
		answer := "Yes, it is."
		tags := []string{"test", "create", "persistence"}

		// Create the card
		cardObj, err := store.CreateCard(title, question, answer, tags)
		if err != nil {
			t.Fatalf("Failed to create card: %v", err)
		}

		// Verify the file exists on disk
		if _, err := os.Stat(cardObj.FilePath); os.IsNotExist(err) {
			t.Errorf("Card file was not created on disk at %s", cardObj.FilePath)
		}

		// Read the file content directly to verify it contains expected data
		content, err := os.ReadFile(cardObj.FilePath)
		if err != nil {
			t.Fatalf("Failed to read card file: %v", err)
		}

		// Check if content contains expected elements
		contentStr := string(content)
		expectedElements := []string{
			title,
			question,
			answer,
			"tags:",
		}
		for _, expected := range expectedElements {
			if !strings.Contains(contentStr, expected) {
				t.Errorf("Card file does not contain expected content: %q", expected)
			}
		}

		// Verify tags are saved correctly
		for _, tag := range tags {
			if !strings.Contains(contentStr, tag) {
				t.Errorf("Card file does not contain tag: %q", tag)
			}
		}
	})

	// Test Case 2: Modify a card and verify changes are persisted
	t.Run("ModifyCardPersistence", func(t *testing.T) {
		// Create a card to modify
		cardObj, err := store.CreateCard(
			"Initial Title",
			"Initial Question",
			"Initial Answer",
			[]string{"initial", "tag"},
		)
		if err != nil {
			t.Fatalf("Failed to create card for modification: %v", err)
		}

		// Modify the card
		cardObj.Title = "Modified Title"
		cardObj.Question = "Modified Question"
		cardObj.Answer = "Modified Answer"
		cardObj.Tags = []string{"modified", "tags"}

		// Save the modifications
		err = store.SaveCard(cardObj)
		if err != nil {
			t.Fatalf("Failed to save modified card: %v", err)
		}

		// Load the card from disk to verify changes persisted
		loadedCard, err := store.LoadCard(cardObj.FilePath)
		if err != nil {
			t.Fatalf("Failed to load modified card: %v", err)
		}

		// Verify the modified fields
		if loadedCard.Title != "Modified Title" {
			t.Errorf("Expected title %q, got %q", "Modified Title", loadedCard.Title)
		}
		if loadedCard.Question != "Modified Question" {
			t.Errorf("Expected question %q, got %q", "Modified Question", loadedCard.Question)
		}
		if loadedCard.Answer != "Modified Answer" {
			t.Errorf("Expected answer %q, got %q", "Modified Answer", loadedCard.Answer)
		}
		if !reflect.DeepEqual(loadedCard.Tags, []string{"modified", "tags"}) {
			t.Errorf("Expected tags %v, got %v", []string{"modified", "tags"}, loadedCard.Tags)
		}
	})

	// Test Case 3: Delete a card and verify it's removed from disk
	t.Run("DeleteCardPersistence", func(t *testing.T) {
		// Create a card to delete
		cardObj, err := store.CreateCard(
			"Card to Delete",
			"Will this card be deleted?",
			"Yes, it will.",
			[]string{"delete", "test"},
		)
		if err != nil {
			t.Fatalf("Failed to create card for deletion: %v", err)
		}

		filePath := cardObj.FilePath

		// Verify the file exists before deletion
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("Card file does not exist before deletion: %s", filePath)
		}

		// Delete the card
		err = store.DeleteCard(cardObj)
		if err != nil {
			t.Fatalf("Failed to delete card: %v", err)
		}

		// Verify the file no longer exists
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Card file still exists after deletion: %s", filePath)
		}

		// Verify the card is no longer in the store's map
		_, exists := store.GetCardByPath(filePath)
		if exists {
			t.Errorf("Card still exists in store's map after deletion")
		}
	})
}

// TestMetadataUpdatesInReviews tests that metadata updates during reviews correctly save to frontmatter
func TestMetadataUpdatesInReviews(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-metadata-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a card store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}
	defer store.Close()

	// Create a card for review
	cardObj, err := store.CreateCard(
		"Review Metadata Test",
		"Does review metadata update properly?",
		"We'll find out through testing!",
		[]string{"review", "metadata"},
	)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Save the original metadata for later comparison
	originalInterval := cardObj.ReviewInterval
	originalLastReviewed := cardObj.LastReviewed

	// Perform a review with rating 4 (easy)
	err = store.ReviewCard(cardObj, 4)
	if err != nil {
		t.Fatalf("Failed to review card: %v", err)
	}

	// Check that the in-memory card is updated
	if cardObj.ReviewInterval <= originalInterval {
		t.Errorf("Expected review interval to increase, got %d (was %d)", cardObj.ReviewInterval, originalInterval)
	}
	if !cardObj.LastReviewed.After(originalLastReviewed) {
		t.Errorf("Expected last reviewed date to update, but it didn't change")
	}

	// Load the card from disk to verify changes were saved to frontmatter
	loadedCard, err := store.LoadCard(cardObj.FilePath)
	if err != nil {
		t.Fatalf("Failed to load card after review: %v", err)
	}

	// Check that the loaded card has the same metadata
	if loadedCard.ReviewInterval != cardObj.ReviewInterval {
		t.Errorf("Expected loaded review interval %d, got %d", cardObj.ReviewInterval, loadedCard.ReviewInterval)
	}
	if !loadedCard.LastReviewed.Equal(cardObj.LastReviewed) {
		t.Errorf("Expected loaded last reviewed %v, got %v", cardObj.LastReviewed, loadedCard.LastReviewed)
	}
	if loadedCard.Difficulty != cardObj.Difficulty {
		t.Errorf("Expected loaded difficulty %d, got %d", cardObj.Difficulty, loadedCard.Difficulty)
	}

	// Read the file directly to verify frontmatter contains updated metadata
	content, err := os.ReadFile(cardObj.FilePath)
	if err != nil {
		t.Fatalf("Failed to read card file: %v", err)
	}

	// Verify the YAML frontmatter contains updated fields
	contentStr := string(content)
	if !strings.Contains(contentStr, "review_interval:") {
		t.Errorf("Card file does not contain review_interval field")
	}
	if !strings.Contains(contentStr, "last_reviewed:") {
		t.Errorf("Card file does not contain last_reviewed field")
	}
	if !strings.Contains(contentStr, "difficulty:") {
		t.Errorf("Card file does not contain difficulty field")
	}
}

// TestMalformedFrontmatterHandling tests how the application handles malformed frontmatter
func TestMalformedFrontmatterHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-malformed-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a card store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}
	defer store.Close()

	// Test Case 1: Invalid YAML syntax
	t.Run("InvalidYAMLSyntax", func(t *testing.T) {
		// Create a file with invalid YAML syntax
		invalidYAML := `---
tags: [test, this is invalid: [
created: 2025-03-24
---

# Card with Invalid YAML

## Question

Is this YAML valid?

## Answer

No, it has syntax errors.
`
		filePath := filepath.Join(tempDir, "invalid-yaml.md")
		err := os.WriteFile(filePath, []byte(invalidYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Try to load the card and expect an error
		_, err = store.LoadCard(filePath)
		if err == nil {
			t.Errorf("Expected error when loading card with invalid YAML, got nil")
		}
	})

	// Test Case 2: Missing frontmatter
	t.Run("MissingFrontmatter", func(t *testing.T) {
		// Create a file with no frontmatter
		noFrontmatter := `# Card with No Frontmatter

## Question

Does this card have frontmatter?

## Answer

No, it doesn't.
`
		filePath := filepath.Join(tempDir, "no-frontmatter.md")
		err := os.WriteFile(filePath, []byte(noFrontmatter), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Try to load the card and expect an error
		_, err = store.LoadCard(filePath)
		if err == nil {
			t.Errorf("Expected error when loading card without frontmatter, got nil")
		}
	})

	// Test Case 3: Malformed frontmatter delimiters
	t.Run("MalformedDelimiters", func(t *testing.T) {
		// Create a file with wrong frontmatter delimiters
		malformedDelimiters := `----
tags: [test]
----

# Card with Wrong Delimiters

## Question

Are these the correct frontmatter delimiters?

## Answer

No, they should be triple dashes (---).
`
		filePath := filepath.Join(tempDir, "wrong-delimiters.md")
		err := os.WriteFile(filePath, []byte(malformedDelimiters), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Try to load the card and expect an error
		_, err = store.LoadCard(filePath)
		if err == nil {
			t.Errorf("Expected error when loading card with malformed delimiters, got nil")
		}
	})

	// Test Case 4: Missing required sections
	t.Run("MissingSections", func(t *testing.T) {
		// Create a file with valid frontmatter but missing question/answer sections
		missingSections := `---
tags: [test]
created: 2025-03-24
---

# Card with Missing Sections

This card has no Question and Answer sections.
`
		filePath := filepath.Join(tempDir, "missing-sections.md")
		err := os.WriteFile(filePath, []byte(missingSections), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Try to load the card and expect either error or empty sections
		card, err := store.LoadCard(filePath)

		// We should either get an error or a card with empty question/answer
		if err == nil {
			if card.Question != "" || card.Answer != "" {
				t.Errorf("Expected empty question/answer for card with missing sections, got: %q/%q",
					card.Question, card.Answer)
			}
		}
	})

	// Test Case 5: Non-string fields in YAML
	t.Run("NonStringYAMLFields", func(t *testing.T) {
		// Create file with non-string fields
		nonStringFields := `---
tags: [test]
created: 2025-03-24
review_interval: "not a number"
difficulty: [1, 2, 3]
---

# Card with Non-String Fields

## Question

Are these fields valid?

## Answer

No, review_interval should be a number and difficulty should be a number.
`
		filePath := filepath.Join(tempDir, "non-string-fields.md")
		err := os.WriteFile(filePath, []byte(nonStringFields), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Try to load the card - this may or may not error depending on YAML parser behavior
		card, err := store.LoadCard(filePath)
		if err == nil {
			// If no error, ensure defaults are used for invalid fields
			if card.ReviewInterval != 0 {
				t.Errorf("Expected review_interval to default to 0 for invalid value, got %d",
					card.ReviewInterval)
			}
			if card.Difficulty != 0 {
				t.Errorf("Expected difficulty to default to 0 for invalid value, got %d",
					card.Difficulty)
			}
		}
	})
}

// TestConcurrentCardOperations tests data integrity during concurrent card operations
func TestConcurrentCardOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-concurrent-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a card store
	store, err := NewCardStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}
	defer store.Close()

	// Test Case 1: Concurrent card creation
	t.Run("ConcurrentCreation", func(t *testing.T) {
		var wg sync.WaitGroup
		numCards := 10
		cards := make([]*card.Card, numCards)

		// Create cards concurrently
		for i := 0; i < numCards; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				title := fmt.Sprintf("Concurrent Card %d", idx)
				question := fmt.Sprintf("Concurrent Question %d", idx)
				answer := fmt.Sprintf("Concurrent Answer %d", idx)
				tags := []string{"concurrent", fmt.Sprintf("card-%d", idx)}

				card, err := store.CreateCard(title, question, answer, tags)
				if err != nil {
					t.Errorf("Failed to create card %d: %v", idx, err)
					return
				}

				cards[idx] = card
			}(i)
		}

		wg.Wait()

		// Verify all cards were created
		for i, card := range cards {
			if card == nil {
				t.Errorf("Card %d was not created", i)
				continue
			}

			// Verify the file exists
			if _, err := os.Stat(card.FilePath); os.IsNotExist(err) {
				t.Errorf("Card file %d does not exist: %s", i, card.FilePath)
			}
		}

		// Count cards in the store
		storeCardCount := store.GetCardCount()
		if storeCardCount != numCards {
			t.Errorf("Expected %d cards in store, got %d", numCards, storeCardCount)
		}
	})

	// Test Case 2: Concurrent modification of the same card
	t.Run("ConcurrentModification", func(t *testing.T) {
		// Create a card to be modified concurrently
		sharedCard, err := store.CreateCard(
			"Shared Card",
			"Will this card handle concurrent modifications?",
			"Let's find out!",
			[]string{"shared", "concurrent"},
		)
		if err != nil {
			t.Fatalf("Failed to create shared card: %v", err)
		}

		var wg sync.WaitGroup
		numWorkers := 5

		// Record the original filepath to verify the same card is being modified
		originalPath := sharedCard.FilePath

		// Have multiple goroutines modify the same card
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Load the card
				card, err := store.LoadCard(originalPath)
				if err != nil {
					t.Errorf("Worker %d failed to load card: %v", idx, err)
					return
				}

				// Modify only the tags field to avoid conflicting changes
				card.Tags = append(card.Tags, fmt.Sprintf("worker-%d", idx))

				// Save the card
				err = store.SaveCard(card)
				if err != nil {
					t.Errorf("Worker %d failed to save card: %v", idx, err)
				}

				// Small delay to increase chance of concurrency issues
				time.Sleep(10 * time.Millisecond)
			}(i)
		}

		wg.Wait()

		// Load the final card
		finalCard, err := store.LoadCard(originalPath)
		if err != nil {
			t.Fatalf("Failed to load final card: %v", err)
		}

		// The card should still exist and have at least the original tags plus some worker tags
		if len(finalCard.Tags) < 2 {
			t.Errorf("Expected card to maintain its tags, got: %v", finalCard.Tags)
		}
	})

	// Test Case 3: Creating and deleting cards concurrently
	t.Run("ConcurrentCreateDelete", func(t *testing.T) {
		var wg sync.WaitGroup
		numCards := 5
		createdPaths := make([]string, numCards)

		// First create some cards
		for i := 0; i < numCards; i++ {
			title := fmt.Sprintf("ConcurrentCD Card %d", i)
			question := fmt.Sprintf("ConcurrentCD Question %d", i)
			answer := fmt.Sprintf("ConcurrentCD Answer %d", i)
			tags := []string{"concurrentCD", fmt.Sprintf("card-%d", i)}

			card, err := store.CreateCard(title, question, answer, tags)
			if err != nil {
				t.Fatalf("Failed to create card %d: %v", i, err)
			}

			createdPaths[i] = card.FilePath
		}

		// Now concurrently create some cards while deleting others
		for i := 0; i < numCards; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				if idx%2 == 0 {
					// Even indices: create a new card
					title := fmt.Sprintf("New ConcurrentCD Card %d", idx)
					question := fmt.Sprintf("New ConcurrentCD Question %d", idx)
					answer := fmt.Sprintf("New ConcurrentCD Answer %d", idx)
					tags := []string{"newConcurrentCD", fmt.Sprintf("new-card-%d", idx)}

					_, err := store.CreateCard(title, question, answer, tags)
					if err != nil {
						t.Errorf("Failed to create new card %d: %v", idx, err)
					}
				} else {
					// Odd indices: delete an existing card
					path := createdPaths[idx]
					cardObj, exists := store.GetCardByPath(path)
					if !exists {
						t.Errorf("Card %d not found for deletion: %s", idx, path)
						return
					}

					err := store.DeleteCard(cardObj)
					if err != nil {
						t.Errorf("Failed to delete card %d: %v", idx, err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify the odd-indexed cards were deleted
		for i := 0; i < numCards; i++ {
			if i%2 == 1 {
				// This card should have been deleted
				path := createdPaths[i]
				_, exists := store.GetCardByPath(path)
				if exists {
					t.Errorf("Card %d still exists after deletion: %s", i, path)
				}

				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("Card file %d still exists after deletion: %s", i, path)
				}
			}
		}
	})
}

// TestFrontmatterParsingEdgeCases tests handling of edge cases in frontmatter parsing
func TestFrontmatterParsingEdgeCases(t *testing.T) {
	// Test various edge cases directly with the parser
	t.Run("EmptyFrontmatter", func(t *testing.T) {
		content := []byte(`---
---

# Card with Empty Frontmatter

## Question

Is empty frontmatter valid?

## Answer

Yes, it should parse without error.
`)
		card, err := parser.ParseMarkdown(content)
		if err != nil {
			t.Errorf("Failed to parse empty frontmatter: %v", err)
		}
		if card == nil {
			t.Errorf("Expected non-nil card for empty frontmatter")
		}
	})

	t.Run("WhitespaceFrontmatter", func(t *testing.T) {
		content := []byte(`---


---

# Card with Whitespace Frontmatter

## Question

Is whitespace-only frontmatter valid?

## Answer

Yes, it should parse without error.
`)
		card, err := parser.ParseMarkdown(content)
		if err != nil {
			t.Errorf("Failed to parse whitespace frontmatter: %v", err)
		}
		if card == nil {
			t.Errorf("Expected non-nil card for whitespace frontmatter")
		}
	})

	t.Run("CommentsInFrontmatter", func(t *testing.T) {
		content := []byte(`---
tags: [test]
# This is a comment
created: 2025-03-24
---

# Card with Comments in Frontmatter

## Question

Are comments in frontmatter handled correctly?

## Answer

It depends on the YAML parser behavior.
`)
		// This may or may not error depending on YAML parser
		_, err := parser.ParseMarkdown(content)
		// We don't assert on the result since YAML parsers differ in comment handling
		t.Logf("Comments in frontmatter result: %v", err)
	})

	t.Run("InvalidDateFormat", func(t *testing.T) {
		content := []byte(`---
tags: [test]
created: not-a-date
last_reviewed: 2025/03/24
---

# Card with Invalid Dates

## Question

Are invalid dates handled gracefully?

## Answer

They should default to zero values.
`)
		card, err := parser.ParseMarkdown(content)
		if err != nil {
			t.Logf("Parser error for invalid dates: %v", err)
		}

		if card != nil && !card.Created.IsZero() {
			t.Errorf("Expected zero time for invalid created date, got %v", card.Created)
		}
	})

	t.Run("SpecialCharactersInFields", func(t *testing.T) {
		content := []byte(`---
tags: ["special!@#$%^&*()_+", "unicode: 你好, こんにちは"]
---

# Card with Special Characters

## Question

Are _special_ *characters* and [formatting] in frontmatter handled correctly?

## Answer

They should be preserved.
`)
		card, err := parser.ParseMarkdown(content)
		if err != nil {
			t.Errorf("Failed to parse special characters: %v", err)
		}

		if card == nil || len(card.Tags) != 2 {
			t.Errorf("Expected 2 tags with special characters")
			return
		}

		if card.Tags[0] != "special!@#$%^&*()_+" {
			t.Errorf("Special characters not preserved in tag, got: %s", card.Tags[0])
		}

		if card.Tags[1] != "unicode: 你好, こんにちは" {
			t.Errorf("Unicode characters not preserved in tag, got: %s", card.Tags[1])
		}
	})
}

// Test how metadata is encoded and decoded during formatting
func TestMetadataFormattingRoundTrip(t *testing.T) {
	// Create a card with known metadata values
	originalCard := &card.Card{
		Title:          "Metadata Test",
		Question:       "Is metadata preserved in a round trip?",
		Answer:         "Let's find out!",
		Tags:           []string{"metadata", "test", "round-trip"},
		Created:        time.Date(2025, 3, 24, 12, 34, 56, 0, time.UTC),
		LastReviewed:   time.Date(2025, 3, 23, 10, 11, 12, 0, time.UTC),
		ReviewInterval: 5,
		Difficulty:     4,
	}

	// Format the card to markdown
	content, err := parser.FormatCardAsMarkdown(originalCard)
	if err != nil {
		t.Fatalf("Failed to format card: %v", err)
	}

	// Parse the markdown back to a card
	parsedCard, err := parser.ParseMarkdown(content)
	if err != nil {
		t.Fatalf("Failed to parse formatted markdown: %v", err)
	}

	// Verify all metadata fields match
	if parsedCard.Title != originalCard.Title {
		t.Errorf("Title mismatch: expected %q, got %q", originalCard.Title, parsedCard.Title)
	}

	if parsedCard.Question != originalCard.Question {
		t.Errorf("Question mismatch: expected %q, got %q", originalCard.Question, parsedCard.Question)
	}

	if parsedCard.Answer != originalCard.Answer {
		t.Errorf("Answer mismatch: expected %q, got %q", originalCard.Answer, parsedCard.Answer)
	}

	if !reflect.DeepEqual(parsedCard.Tags, originalCard.Tags) {
		t.Errorf("Tags mismatch: expected %v, got %v", originalCard.Tags, parsedCard.Tags)
	}

	// Times should be the same when rounded to seconds (sub-second precision may be lost)
	originalCreated := originalCard.Created.Truncate(time.Second)
	parsedCreated := parsedCard.Created.Truncate(time.Second)
	if !parsedCreated.Equal(originalCreated) {
		t.Errorf("Created time mismatch: expected %v, got %v", originalCreated, parsedCreated)
	}

	originalLastReviewed := originalCard.LastReviewed.Truncate(time.Second)
	parsedLastReviewed := parsedCard.LastReviewed.Truncate(time.Second)
	if !parsedLastReviewed.Equal(originalLastReviewed) {
		t.Errorf("LastReviewed time mismatch: expected %v, got %v",
			originalLastReviewed, parsedLastReviewed)
	}

	if parsedCard.ReviewInterval != originalCard.ReviewInterval {
		t.Errorf("ReviewInterval mismatch: expected %d, got %d",
			originalCard.ReviewInterval, parsedCard.ReviewInterval)
	}

	if parsedCard.Difficulty != originalCard.Difficulty {
		t.Errorf("Difficulty mismatch: expected %d, got %d",
			originalCard.Difficulty, parsedCard.Difficulty)
	}
}
