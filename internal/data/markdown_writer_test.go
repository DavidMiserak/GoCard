// File: internal/data/markdown_writer_test.go

package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal_file.md", "normal_file.md"},
		{"file with spaces.md", "file_with_spaces.md"},
		{"file/with/slashes.md", "file-with-slashes.md"},
		{"file:with:colons.md", "file-with-colons.md"},
		{"file*with*stars.md", "file-with-stars.md"},
		{"file?with?questions.md", "file-with-questions.md"},
		{"file\"with\"quotes.md", "file-with-quotes.md"},
		{"file<with>brackets.md", "file-with-brackets.md"},
		{"file|with|pipes.md", "file-with-pipes.md"},
		{".leading.dot.md", "leading.dot.md"},
		{"trailing.dot.", "trailing.dot"},
		{"", "card"},
		{" \t\n", "card"},
		{"...", "card"},
		{"complex file: with * many ? special / chars.md", "complex_file-_with_-_many_-_special_-_chars.md"},
	}

	for _, tc := range testCases {
		result := SanitizeFilename(tc.input)
		if result != tc.expected {
			t.Errorf("SanitizeFilename(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestWriteCard(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "card-write-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create a test card
	testPath := filepath.Join(tempDir, "test-write.md")
	now := time.Now()
	card := model.Card{
		ID:           testPath,
		Question:     "What is a unit test?",
		Answer:       "A test that verifies a small unit of functionality.",
		DeckID:       "test-deck",
		LastReviewed: now,
		NextReview:   now.AddDate(0, 0, 3),
		Ease:         2.5,
		Interval:     3,
		Rating:       4,
	}

	// Write card to file
	err = WriteCard(card, testPath)
	if err != nil {
		t.Fatalf("WriteCard error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("WriteCard did not create file at %s", testPath)
	}

	// Read card back from file
	readCard, err := ParseMarkdownFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read back card: %v", err)
	}

	// Verify content is preserved
	if readCard.Question != card.Question {
		t.Errorf("Expected question %q, got %q", card.Question, readCard.Question)
	}

	if readCard.Answer != card.Answer {
		t.Errorf("Expected answer %q, got %q", card.Answer, readCard.Answer)
	}

	if readCard.FrontMatter.ReviewInterval != card.Interval {
		t.Errorf("Expected interval %d, got %d", card.Interval, readCard.FrontMatter.ReviewInterval)
	}

	if readCard.FrontMatter.Difficulty != card.Ease {
		t.Errorf("Expected difficulty %f, got %f", card.Ease, readCard.FrontMatter.Difficulty)
	}
}

func TestWriteDeckToMarkdown(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "deck-write-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create a test deck with multiple cards
	now := time.Now()
	deck := &model.Deck{
		ID:          "test-deck",
		Name:        "Test Deck",
		Description: "A deck for testing",
		CreatedAt:   now,
		LastStudied: now,
		Cards: []model.Card{
			{
				Question: "Question 1",
				Answer:   "Answer 1",
				Interval: 1,
				Ease:     2.1,
			},
			{
				Question: "Question 2",
				Answer:   "Answer 2",
				Interval: 2,
				Ease:     2.2,
			},
		},
	}

	// Write deck to markdown files
	err = WriteDeckToMarkdown(deck, tempDir)
	if err != nil {
		t.Fatalf("WriteDeckToMarkdown error: %v", err)
	}

	// Check that files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check file content for one card
	cardPath := filepath.Join(tempDir, "card_1.md")
	readCard, err := ParseMarkdownFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read card: %v", err)
	}

	if readCard.Question != "Question 1" {
		t.Errorf("Expected question %q, got %q", "Question 1", readCard.Question)
	}
}

func TestWriteDeckWithProblematicFilenames(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "deck-filenames-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create a test deck with cards that have problematic IDs
	deck := &model.Deck{
		ID:   "test-problematic-names",
		Name: "Test Problematic Names",
		Cards: []model.Card{
			{
				ID:       "file with spaces.md",
				Question: "Spaces Test",
				Answer:   "This filename has spaces",
			},
			{
				ID:       "file/with/slashes.md",
				Question: "Slashes Test",
				Answer:   "This filename has slashes",
			},
			{
				ID:       "file:with:colons*and?special|chars.md",
				Question: "Special Chars Test",
				Answer:   "This filename has special characters",
			},
		},
	}

	// Write deck to markdown files
	err = WriteDeckToMarkdown(deck, tempDir)
	if err != nil {
		t.Fatalf("WriteDeckToMarkdown error: %v", err)
	}

	// List all files created for debugging
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	t.Logf("Created files: %v", fileNames)

	// Check that files were created with sanitized names
	expectedFiles := []string{
		"file_with_spaces.md",
		"slashes.md", // Only the base name is used
		"file-with-colons-and-special-chars.md",
	}

	for _, expectedFile := range expectedFiles {
		path := filepath.Join(tempDir, expectedFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected sanitized file %s was not created", expectedFile)
		}
	}

	// Check card content for one sanitized file
	cardPath := filepath.Join(tempDir, "file_with_spaces.md")
	readCard, err := ParseMarkdownFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read card: %v", err)
	}

	if readCard.Question != "Spaces Test" {
		t.Errorf("Expected question %q, got %q", "Spaces Test", readCard.Question)
	}
}

func TestUpdateCardFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "card-update-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create initial card file
	testPath := filepath.Join(tempDir, "update-test.md")
	initialContent := `---
tags: [go,test]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 3
difficulty: 2.0
---

# Question

What is Go testing?

## Answer

Go testing is a framework for writing automated tests in Go.
`
	if err := os.WriteFile(testPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Create updated card
	now := time.Now()
	updatedCard := model.Card{
		ID:           testPath,
		Question:     "What is Go testing? (Updated)",
		Answer:       "Go testing is a framework for writing automated tests in Go. It's easy to use!",
		LastReviewed: now,
		NextReview:   now.AddDate(0, 0, 5),
		Ease:         2.5,
		Interval:     5,
		Rating:       5,
	}

	// Update the card file
	err = UpdateCardFile(updatedCard)
	if err != nil {
		t.Fatalf("UpdateCardFile error: %v", err)
	}

	// Read back updated card
	readCard, err := ParseMarkdownFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read updated card: %v", err)
	}

	// Verify content is updated
	if readCard.Question != updatedCard.Question {
		t.Errorf("Expected updated question %q, got %q", updatedCard.Question, readCard.Question)
	}

	if readCard.Answer != updatedCard.Answer {
		t.Errorf("Expected updated answer %q, got %q", updatedCard.Answer, readCard.Answer)
	}

	// Verify original tags are preserved
	if len(readCard.FrontMatter.Tags) != 2 || readCard.FrontMatter.Tags[0] != "go" || readCard.FrontMatter.Tags[1] != "test" {
		t.Errorf("Expected tags [go test], got %v", readCard.FrontMatter.Tags)
	}

	// Verify creation date is preserved
	expectedDate := time.Date(2025, 3, 22, 0, 0, 0, 0, time.UTC)
	if !readCard.FrontMatter.Created.Equal(expectedDate) {
		t.Errorf("Expected created date preserved as %v, got %v", expectedDate, readCard.FrontMatter.Created)
	}

	// Verify review interval is updated
	if readCard.FrontMatter.ReviewInterval != updatedCard.Interval {
		t.Errorf("Expected updated interval %d, got %d", updatedCard.Interval, readCard.FrontMatter.ReviewInterval)
	}
}
