// File: internal/storage/parser/syntax_test.go
package parser

import (
	"fmt"
	"strings"
	"testing"
)

// TestSyntaxHighlighting tests the syntax highlighting functionality
func TestSyntaxHighlighting(t *testing.T) {
	// Test various code samples
	testCases := []struct {
		name     string
		language string
		code     string
	}{
		{
			name:     "Go Code",
			language: "go",
			code: `package main

import "fmt"

func main() {
    // This is a comment
    greeting := "Hello, world!"
    fmt.Println(greeting)
}`,
		},
		{
			name:     "Python Code",
			language: "python",
			code: `def fibonacci(n):
    """Return the nth fibonacci number."""
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

# Print the first 10 fibonacci numbers
for i in range(10):
    print(f"Fibonacci({i}) = {fibonacci(i)}")`,
		},
		{
			name:     "JavaScript Code",
			language: "javascript",
			code: `// A simple function
function sayHello(name) {
    return "Hello, " + name;
}

// Sample usage
console.log(sayHello("Sara"));`,
		},
		{
			name:     "SQL Code",
			language: "sql",
			code: `-- Create a table
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert a user
INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com');

-- Query users
SELECT * FROM users WHERE name LIKE 'J%' ORDER BY created_at DESC;`,
		},
	}

	// Test different themes
	themes := []string{
		"monokai",
		"github",
		"dracula",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, theme := range themes {
				t.Run(theme, func(t *testing.T) {
					config := DefaultSyntaxConfig()
					config.Theme = theme

					// Create markdown with a code block
					markdown := fmt.Sprintf("# %s Example\n\n```%s\n%s\n```\n",
						tc.name, tc.language, tc.code)

					// Render the markdown with syntax highlighting
					html, err := RenderMarkdownWithHighlighting(markdown, config)
					if err != nil {
						t.Fatalf("Failed to render markdown: %v", err)
					}

					// Basic checks to ensure it worked
					if !strings.Contains(html, "<pre") || !strings.Contains(html, "<code") {
						t.Errorf("Generated HTML doesn't contain code block elements")
					}

					// Print the HTML for manual inspection if in verbose mode
					if testing.Verbose() {
						t.Logf("Generated HTML for %s with theme %s:\n%s", tc.name, theme, html)
					}
				})
			}
		})
	}

	// Test the available themes
	t.Run("AvailableThemes", func(t *testing.T) {
		themes := AvailableThemes()
		if len(themes) < 5 {
			t.Errorf("Expected at least 5 available themes, got %d", len(themes))
		}

		// Check for a few common themes
		foundMonokai := false
		foundGithub := false

		for _, theme := range themes {
			if theme == "monokai" {
				foundMonokai = true
			}
			if theme == "github" {
				foundGithub = true
			}
		}

		if !foundMonokai {
			t.Errorf("Monokai theme not found in available themes")
		}
		if !foundGithub {
			t.Errorf("Github theme not found in available themes")
		}
	})
}
