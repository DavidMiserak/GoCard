// File: internal/storage/parser/markdown_test.go

package parser

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseMarkdown(t *testing.T) {
	// Create a valid markdown document with YAML frontmatter
	// Note: Using inline array syntax for tags that matches what the parser expects
	validMarkdown := `---
tags: [test, markdown]
created: 2025-01-01T12:00:00Z
last_reviewed: 2024-12-31T12:00:00Z
review_interval: 3
difficulty: 4
---

# Test Card Title

## Question

Is this a test question?

## Answer

Yes, this is a test answer.
With multiple lines.
`

	// Test case for invalid markdown (missing frontmatter)
	invalidMarkdown := `# No Frontmatter

## Question

This card has no frontmatter.

## Answer

It should fail to parse.
`

	// Test case for invalid frontmatter format
	invalidFrontmatter := `---
invalid: yaml: [
---

# Invalid Frontmatter

## Question

This has invalid YAML in the frontmatter.

## Answer

It should fail to parse.
`

	testCases := []struct {
		name        string
		content     []byte
		expectError bool
		checkFields bool // whether to check the parsed fields
	}{
		{
			name:        "Valid markdown",
			content:     []byte(validMarkdown),
			expectError: false,
			checkFields: true,
		},
		{
			name:        "Invalid markdown - no frontmatter",
			content:     []byte(invalidMarkdown),
			expectError: true,
			checkFields: false,
		},
		{
			name:        "Invalid frontmatter format",
			content:     []byte(invalidFrontmatter),
			expectError: true,
			checkFields: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cardObj, err := ParseMarkdown(tc.content)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			if !tc.checkFields {
				return
			}

			// Check extracted fields
			expectedTitle := "Test Card Title"
			if cardObj.Title != expectedTitle {
				t.Errorf("Expected title %q, got %q", expectedTitle, cardObj.Title)
			}

			expectedQuestion := "Is this a test question?"
			if cardObj.Question != expectedQuestion {
				t.Errorf("Expected question %q, got %q", expectedQuestion, cardObj.Question)
			}

			expectedAnswer := "Yes, this is a test answer.\nWith multiple lines."
			if cardObj.Answer != expectedAnswer {
				t.Errorf("Expected answer %q, got %q", expectedAnswer, cardObj.Answer)
			}

			expectedTags := []string{"test", "markdown"}
			if len(cardObj.Tags) != len(expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(expectedTags), len(cardObj.Tags))
			} else {
				for i, tag := range expectedTags {
					if cardObj.Tags[i] != tag {
						t.Errorf("Expected tag %q, got %q", tag, cardObj.Tags[i])
					}
				}
			}

			expectedInterval := 3
			if cardObj.ReviewInterval != expectedInterval {
				t.Errorf("Expected review interval %d, got %d", expectedInterval, cardObj.ReviewInterval)
			}

			expectedDifficulty := 4
			if cardObj.Difficulty != expectedDifficulty {
				t.Errorf("Expected difficulty %d, got %d", expectedDifficulty, cardObj.Difficulty)
			}

			// Check timestamps (created and last_reviewed)
			expectedCreated := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
			if !cardObj.Created.Equal(expectedCreated) {
				t.Errorf("Expected created time %v, got %v", expectedCreated, cardObj.Created)
			}

			expectedLastReviewed := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)
			if !cardObj.LastReviewed.Equal(expectedLastReviewed) {
				t.Errorf("Expected last reviewed time %v, got %v", expectedLastReviewed, cardObj.LastReviewed)
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	testCases := []struct {
		name        string
		markdown    string
		expectError bool
		checkHTML   bool
	}{
		{
			name:        "Basic markdown",
			markdown:    "# Header\n\nThis is a paragraph.",
			expectError: false,
			checkHTML:   true,
		},
		{
			name:        "Markdown with code block",
			markdown:    "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```",
			expectError: false,
			checkHTML:   true,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			html, err := RenderMarkdown(tc.markdown)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			if !tc.checkHTML {
				return
			}

			// Basic checks for HTML output
			if len(html) == 0 {
				t.Error("Expected non-empty HTML output")
			}

			// Check for specific HTML elements based on input markdown
			if tc.name == "Basic markdown" {
				if !strings.Contains(html, "Header") {
					t.Errorf("Expected HTML to contain 'Header', got: %s", html)
				}
				if !strings.Contains(html, "paragraph") {
					t.Errorf("Expected HTML to contain 'paragraph', got: %s", html)
				}
			}

			if tc.name == "Markdown with code block" {
				// Just check for code content
				if !strings.Contains(html, "Hello") {
					t.Errorf("Expected HTML to contain code content 'Hello', got: %s", html)
				}
			}
		})
	}
}

func TestCreateGoldmarkParser(t *testing.T) {
	parser := createGoldmarkParser()

	// Simple test to ensure function doesn't panic
	if parser == nil {
		t.Error("Expected non-nil markdown parser")
	}

	// Test with a simple markdown document
	doc := "# Test\n\nThis is a test."
	var buf bytes.Buffer

	err := parser.Convert([]byte(doc), &buf)
	if err != nil {
		t.Errorf("Parser conversion failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output from parser")
	}
}
