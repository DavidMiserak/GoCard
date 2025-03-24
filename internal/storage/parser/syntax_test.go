// File: internal/storage/parser/syntax_test.go
package parser

import (
	"strings"
	"testing"
)

func TestDefaultSyntaxConfig(t *testing.T) {
	config := DefaultSyntaxConfig()

	// Check default values
	if config.Theme != "monokai" {
		t.Errorf("Expected default theme to be 'monokai', got '%s'", config.Theme)
	}

	if !config.ShowLineNumbers {
		t.Error("Expected ShowLineNumbers to be true by default")
	}

	if config.DefaultLang != "text" {
		t.Errorf("Expected default language to be 'text', got '%s'", config.DefaultLang)
	}
}

func TestRenderMarkdownWithHighlighting(t *testing.T) {
	// Test code in different languages
	testCases := []struct {
		name     string
		markdown string
		language string
	}{
		{
			name:     "Go code",
			markdown: "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```",
			language: "go",
		},
		{
			name:     "Python code",
			markdown: "```python\ndef hello():\n    print(\"Hello\")\n```",
			language: "python",
		},
		{
			name:     "JavaScript code",
			markdown: "```javascript\nfunction hello() {\n    console.log(\"Hello\");\n}\n```",
			language: "javascript",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultSyntaxConfig()

			html, err := RenderMarkdownWithHighlighting(tc.markdown, config)
			if err != nil {
				t.Fatalf("Failed to render markdown with highlighting: %v", err)
			}

			// Check for code content only
			if !strings.Contains(html, "Hello") {
				t.Error("Expected output to contain the code content")
			}

			// Check for some kind of code formatting
			if !strings.Contains(html, "<pre") && !strings.Contains(html, "<code") && !strings.Contains(html, "code") {
				t.Error("Expected output to contain code formatting")
			}
		})
	}
}

func TestThemeFunctions(t *testing.T) {
	// Test AvailableThemes
	themes := AvailableThemes()
	if len(themes) == 0 {
		t.Error("Expected non-empty list of available themes")
	}

	// Look for at least one common theme
	commonThemes := []string{"monokai", "github", "dracula", "solarized-dark"}
	found := false
	for _, theme := range commonThemes {
		for _, availTheme := range themes {
			if availTheme == theme {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Errorf("Expected at least one common theme (monokai, github, etc.) to be available")
	}

	// Test IsDarkTheme with at least one dark and one light theme if available
	darkThemes := []string{"monokai", "dracula", "solarized-dark", "nord"}
	lightThemes := []string{"github", "friendly", "solarized-light"}

	// Just test that the function executes without errors
	for _, theme := range themes {
		isDark := IsDarkTheme(theme)
		// We're not asserting the result, just making sure it doesn't panic
		_ = isDark
	}

	// If we have known themes, test their expected dark/light values
	for _, theme := range darkThemes {
		// Only test if this theme is available
		for _, availTheme := range themes {
			if theme == availTheme {
				if !IsDarkTheme(theme) {
					t.Errorf("Expected '%s' to be identified as a dark theme", theme)
				}
				break
			}
		}
	}

	for _, theme := range lightThemes {
		// Only test if this theme is available
		for _, availTheme := range themes {
			if theme == availTheme {
				if IsDarkTheme(theme) {
					t.Errorf("Expected '%s' to be identified as a light theme", theme)
				}
				break
			}
		}
	}
}
