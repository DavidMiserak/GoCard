// File: internal/data/markdown_parser_test.go

package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestParseMarkdownFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "markdown-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create a test markdown file
	testFile := filepath.Join(tempDir, "test-card.md")
	content := `---
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
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test parsing
	card, err := ParseMarkdownFile(testFile)
	if err != nil {
		t.Fatalf("ParseMarkdownFile error: %v", err)
	}

	// Validate parsed data
	if card.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, card.Path)
	}

	if len(card.FrontMatter.Tags) != 2 || card.FrontMatter.Tags[0] != "go" || card.FrontMatter.Tags[1] != "test" {
		t.Errorf("Expected tags [go test], got %v", card.FrontMatter.Tags)
	}

	expectedDate := time.Date(2025, 3, 22, 0, 0, 0, 0, time.UTC)
	if !card.FrontMatter.Created.Equal(expectedDate) {
		t.Errorf("Expected created %v, got %v", expectedDate, card.FrontMatter.Created)
	}

	if card.FrontMatter.ReviewInterval != 3 {
		t.Errorf("Expected interval 3, got %d", card.FrontMatter.ReviewInterval)
	}

	if card.FrontMatter.Difficulty != 2.0 {
		t.Errorf("Expected difficulty 2.0, got %f", card.FrontMatter.Difficulty)
	}

	expectedQuestion := "What is Go testing?"
	if card.Question != expectedQuestion {
		t.Errorf("Expected question %q, got %q", expectedQuestion, card.Question)
	}

	expectedAnswer := "Go testing is a framework for writing automated tests in Go."
	if card.Answer != expectedAnswer {
		t.Errorf("Expected answer %q, got %q", expectedAnswer, card.Answer)
	}
}

func TestToModelCard(t *testing.T) {
	// Create a test markdown card
	created := time.Date(2025, 3, 22, 0, 0, 0, 0, time.UTC)
	lastReviewed := time.Date(2025, 3, 23, 0, 0, 0, 0, time.UTC)

	mc := &MarkdownCard{
		Path: "/path/to/card.md",
		FrontMatter: FrontMatter{
			Tags:           []string{"go", "test"},
			Created:        created,
			LastReviewed:   lastReviewed,
			ReviewInterval: 3,
			Difficulty:     2.1,
		},
		Question: "Test Question",
		Answer:   "Test Answer",
	}

	// Convert to model card
	deckID := "/path/to/deck"
	card := mc.ToModelCard(deckID)

	// Validate conversion
	if card.ID != mc.Path {
		t.Errorf("Expected ID %s, got %s", mc.Path, card.ID)
	}

	if card.Question != mc.Question {
		t.Errorf("Expected question %s, got %s", mc.Question, card.Question)
	}

	if card.Answer != mc.Answer {
		t.Errorf("Expected answer %s, got %s", mc.Answer, card.Answer)
	}

	if card.DeckID != deckID {
		t.Errorf("Expected deckID %s, got %s", deckID, card.DeckID)
	}

	if !card.LastReviewed.Equal(lastReviewed) {
		t.Errorf("Expected lastReviewed %v, got %v", lastReviewed, card.LastReviewed)
	}

	expectedNextReview := lastReviewed.AddDate(0, 0, 3)
	if !card.NextReview.Equal(expectedNextReview) {
		t.Errorf("Expected nextReview %v, got %v", expectedNextReview, card.NextReview)
	}

	if card.Ease != mc.FrontMatter.Difficulty {
		t.Errorf("Expected ease %f, got %f", mc.FrontMatter.Difficulty, card.Ease)
	}

	if card.Interval != mc.FrontMatter.ReviewInterval {
		t.Errorf("Expected interval %d, got %d", mc.FrontMatter.ReviewInterval, card.Interval)
	}

	if card.Rating != 0 {
		t.Errorf("Expected rating 0, got %d", card.Rating)
	}
}

func TestScanDirForMarkdown(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "markdown-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create test files
	files := []string{
		"test1.md",
		"test2.md",
		"not-markdown.txt",
	}

	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", file, err)
		}
	}

	// Create a subdirectory with a markdown file (should be skipped)
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "sub.md")
	if err := os.WriteFile(subFile, []byte("subdir content"), 0644); err != nil {
		t.Fatalf("Failed to write file in subdirectory: %v", err)
	}

	// Test scanning
	mdFiles, err := ScanDirForMarkdown(tempDir)
	if err != nil {
		t.Fatalf("ScanDirForMarkdown error: %v", err)
	}

	// Validate results
	if len(mdFiles) != 2 {
		t.Errorf("Expected 2 markdown files, got %d", len(mdFiles))
	}

	// Check that only markdown files are included
	for _, file := range mdFiles {
		if filepath.Ext(file) != ".md" {
			t.Errorf("Non-markdown file included: %s", file)
		}

		// Check that subdirectory files are not included
		if filepath.Dir(file) != tempDir {
			t.Errorf("File from subdirectory included: %s", file)
		}
	}
}

func TestImportMarkdownToDeck(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "markdown-import")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create test markdown files
	for i := 1; i <= 2; i++ {
		filename := filepath.Join(tempDir, "test"+string(rune('0'+i))+".md")
		content := `---
tags: [go,test` + string(rune('0'+i)) + `]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: ` + string(rune('0'+i)) + `
difficulty: 2.` + string(rune('0'+i)) + `
---

# Question

Test question ` + string(rune('0'+i)) + `?

## Answer

Test answer ` + string(rune('0'+i)) + `.
`
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Create a deck
	deck := &model.Deck{
		ID:   tempDir,
		Name: "Test Deck",
	}

	// Test importing
	if err := ImportMarkdownToDeck(tempDir, deck); err != nil {
		t.Fatalf("ImportMarkdownToDeck error: %v", err)
	}

	// Validate deck
	if len(deck.Cards) != 2 {
		t.Errorf("Expected 2 cards in deck, got %d", len(deck.Cards))
	}

	// Check card content
	for _, card := range deck.Cards {
		if card.DeckID != tempDir {
			t.Errorf("Expected deckID %s, got %s", tempDir, card.DeckID)
		}

		// Check that files were properly parsed
		if card.Question == "" || card.Answer == "" {
			t.Errorf("Card missing question or answer: %+v", card)
		}
	}
}

func TestCreateDeckFromDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "deck-create")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Create test markdown files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(tempDir, "test"+string(rune('0'+i))+".md")
		content := `---
tags: [go,test` + string(rune('0'+i)) + `]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: ` + string(rune('0'+i)) + `
difficulty: 2.` + string(rune('0'+i)) + `
---

# Question

Test question ` + string(rune('0'+i)) + `?

## Answer

Test answer ` + string(rune('0'+i)) + `.
`
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Test creating deck
	deck, err := CreateDeckFromDir(tempDir)
	if err != nil {
		t.Fatalf("CreateDeckFromDir error: %v", err)
	}

	// Validate deck
	if deck.ID != tempDir {
		t.Errorf("Expected ID %s, got %s", tempDir, deck.ID)
	}

	expectedName := filepath.Base(tempDir)
	if deck.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, deck.Name)
	}

	if len(deck.Cards) != 3 {
		t.Errorf("Expected 3 cards in deck, got %d", len(deck.Cards))
	}

	// Check non-directory
	nonDir := filepath.Join(tempDir, "test1.md")
	_, err = CreateDeckFromDir(nonDir)
	if err == nil {
		t.Error("Expected error when creating deck from non-directory, got nil")
	}
}
