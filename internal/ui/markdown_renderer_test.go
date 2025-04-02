// File: internal/ui/markdown_renderer_test.go

package ui

import (
	"regexp"
	"strings"
	"testing"
)

// TestNewMarkdownRenderer tests the constructor function
func TestNewMarkdownRenderer(t *testing.T) {
	testCases := []struct {
		name          string
		width         int
		themeName     string
		expectedWidth int
		expectedTheme string
	}{
		{
			name:          "Default width and theme",
			width:         0,
			themeName:     "",
			expectedWidth: 80,
			expectedTheme: "monokai",
		},
		{
			name:          "Custom width and theme",
			width:         120,
			themeName:     "dracula",
			expectedWidth: 120,
			expectedTheme: "dracula",
		},
		{
			name:          "Negative width",
			width:         -10,
			themeName:     "github",
			expectedWidth: 80,
			expectedTheme: "github",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(tc.width, tc.themeName)

			if renderer.defaultWidth != tc.expectedWidth {
				t.Errorf("Expected width %d, got %d", tc.expectedWidth, renderer.defaultWidth)
			}

			if renderer.syntaxTheme != tc.expectedTheme {
				t.Errorf("Expected theme %s, got %s", tc.expectedTheme, renderer.syntaxTheme)
			}

			if renderer.renderer == nil {
				t.Error("Renderer should not be nil")
			}

			if renderer.renderedCache == nil {
				t.Error("Rendered cache should be initialized")
			}
		})
	}
}

// TestMarkdownRenderer_Render tests the Render method
func TestMarkdownRenderer_Render(t *testing.T) {
	testCases := []struct {
		name     string
		markdown string
		want     func(string) bool
	}{
		{
			name:     "Simple text",
			markdown: "Hello, world!",
			want: func(result string) bool {
				cleanedResult := cleanAnsiCodes(result)
				return cleanedResult != "" && strings.Contains(cleanedResult, "Hello, world!")
			},
		},
		{
			name:     "Markdown with headers",
			markdown: "# Header 1\n## Header 2\n\nSome text.",
			want: func(result string) bool {
				cleanedResult := cleanAnsiCodes(result)
				return strings.Contains(cleanedResult, "Header 1") &&
					strings.Contains(cleanedResult, "Header 2") &&
					strings.Contains(cleanedResult, "Some text.")
			},
		},
		{
			name:     "Code block",
			markdown: "```go\nfunc example() {\n\tfmt.Println(\"Hello\")\n}\n```",
			want: func(result string) bool {
				cleanedResult := cleanAnsiCodes(result)
				return strings.Contains(cleanedResult, "func") &&
					strings.Contains(cleanedResult, "example") &&
					strings.Contains(cleanedResult, "Println")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(80, "monokai")
			result := renderer.Render(tc.markdown)

			if !tc.want(result) {
				t.Errorf("Render did not produce expected output for: %s", tc.name)
			}
		})
	}
}

// TestMarkdownRenderer_UpdateWidth tests the UpdateWidth method
func TestMarkdownRenderer_UpdateWidth(t *testing.T) {
	renderer := NewMarkdownRenderer(80, "monokai")

	// Validate initial state
	if renderer.defaultWidth != 80 {
		t.Errorf("Expected initial width 80, got %d", renderer.defaultWidth)
	}

	// Update width
	renderer.UpdateWidth(120)
	if renderer.defaultWidth != 120 {
		t.Errorf("Expected updated width 120, got %d", renderer.defaultWidth)
	}

	// Test with invalid width
	initialRenderer := renderer.renderer
	renderer.UpdateWidth(-10)
	if renderer.defaultWidth != 120 {
		t.Errorf("Width should not change with invalid input")
	}
	if renderer.renderer != initialRenderer {
		t.Error("Renderer should not be recreated with invalid width")
	}
}

// TestMarkdownRenderer_SetSyntaxTheme tests the SetSyntaxTheme method
func TestMarkdownRenderer_SetSyntaxTheme(t *testing.T) {
	renderer := NewMarkdownRenderer(80, "monokai")

	// Validate initial state
	if renderer.syntaxTheme != "monokai" {
		t.Errorf("Expected initial theme monokai, got %s", renderer.syntaxTheme)
	}

	// Change theme
	renderer.SetSyntaxTheme("github")
	if renderer.syntaxTheme != "github" {
		t.Errorf("Expected updated theme github, got %s", renderer.syntaxTheme)
	}

	// Test empty theme
	renderer.SetSyntaxTheme("")
	if renderer.syntaxTheme != "monokai" {
		t.Errorf("Expected fallback to monokai, got %s", renderer.syntaxTheme)
	}
}

// TestMarkdownRenderer_ClearCache tests the ClearCache method
func TestMarkdownRenderer_ClearCache(t *testing.T) {
	renderer := NewMarkdownRenderer(80, "monokai")

	// Add something to cache
	renderer.Render("# Test")
	if len(renderer.renderedCache) == 0 {
		t.Error("Cache should not be empty after rendering")
	}

	// Clear cache
	renderer.ClearCache()
	if len(renderer.renderedCache) != 0 {
		t.Error("Cache should be empty after ClearCache")
	}
}

// TestRenderCodeBlock tests the renderCodeBlock function
func TestRenderCodeBlock(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		language string
		theme    string
	}{
		{
			name:     "Go code",
			code:     "func Hello() { fmt.Println(\"Hello\") }",
			language: "go",
			theme:    "monokai",
		},
		{
			name:     "Python code",
			code:     "def hello():\n    print('Hello')",
			language: "python",
			theme:    "github",
		},
		{
			name:     "Unknown language",
			code:     "some code",
			language: "unknown",
			theme:    "dracula",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderCodeBlock(tc.code, tc.language, tc.theme)

			// Basic checks
			if result == "" {
				t.Error("Rendered code block should not be empty")
			}

			// If the language is known, we expect some transformation
			if tc.language != "unknown" {
				if result == tc.code {
					t.Error("Code block should be syntax highlighted")
				}
			}
		})
	}
}

// TestMarkdownRenderer_FormattingPreservation checks that specific markdown
// formatting is not accidentally modified during rendering
func TestMarkdownRenderer_FormattingPreservation(t *testing.T) {
	testCases := []struct {
		name     string
		markdown string
		// validateFunc allows custom checks to ensure formatting is preserved
		validateFunc func(input, rendered string) bool
	}{
		{
			name:     "Numbered List Preservation",
			markdown: "1. First item\n2. Second item\n3. Third item",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				// Ensure numbered list markers are preserved exactly
				return strings.Contains(cleanedOutput, "1. First item") &&
					strings.Contains(cleanedOutput, "2. Second item") &&
					strings.Contains(cleanedOutput, "3. Third item")
			},
		},
		{
			name:     "Nested Numbered List Preservation",
			markdown: "1. Parent Item\n   1. Nested Item\n   2. Another Nested Item\n2. Another Parent Item",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				return strings.Contains(cleanedOutput, "1. Parent Item") &&
					strings.Contains(cleanedOutput, "1. Nested Item") &&
					strings.Contains(cleanedOutput, "2. Another Nested Item") &&
					strings.Contains(cleanedOutput, "2. Another Parent Item")
			},
		},
		{
			name:     "Inline Code Preservation",
			markdown: "This is `inline code` that should not change",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				return strings.Contains(cleanedOutput, "inline code")
			},
		},
		{
			name:     "Code Block Exact Spacing",
			markdown: "```go\nfunc Example() {\n    fmt.Println(\"Hello\")\n}\n```",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				// Ensure code block spacing is preserved
				return strings.Contains(cleanedOutput, "    fmt.Println(\"Hello\")") &&
					strings.Contains(cleanedOutput, "func Example() {")
			},
		},
		{
			name:     "Bullet List Preservation",
			markdown: "- First bullet point\n- Second bullet point\n  - Nested bullet point",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				return strings.Contains(cleanedOutput, "• First bullet point") &&
					strings.Contains(cleanedOutput, "• Second bullet point") &&
					strings.Contains(cleanedOutput, "  • Nested bullet point")
			},
		},
		{
			name:     "Block Quote Spacing",
			markdown: "> Block quote.\n> Another line of block quote.\n>> Nested block quote.",
			validateFunc: func(input, rendered string) bool {
				cleanedOutput := cleanAnsiCodes(rendered)
				return strings.Contains(cleanedOutput, "│ Block quote. Another line of block quote.") &&
					strings.Contains(cleanedOutput, "│ │ Nested block quote.")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(80, "monokai")
			rendered := renderer.Render(tc.markdown)

			if !tc.validateFunc(tc.markdown, rendered) {
				t.Errorf("Markdown formatting was not preserved for: %s\n\nOriginal:\n%s\n\nRendered:\n%s",
					tc.name, tc.markdown, rendered)
			}
		})
	}
}

// cleanAnsiCodes removes ANSI color codes from a string
func cleanAnsiCodes(s string) string {
	// This is a basic regex to remove ANSI escape sequences
	ansiRegex := `\x1b\[[0-9;]*m`
	re := regexp.MustCompile(ansiRegex)
	return re.ReplaceAllString(s, "")
}
