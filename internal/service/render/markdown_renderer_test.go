// internal/service/render/markdown_renderer_test.go
package render

import (
	"strings"
	"testing"
)

func TestNewMarkdownRenderer(t *testing.T) {
	renderer := NewMarkdownRenderer()

	// Check that the renderer was created with default values
	markdownRenderer, ok := renderer.(*MarkdownRenderer)
	if !ok {
		t.Fatal("NewMarkdownRenderer() did not return a *MarkdownRenderer")
	}

	if markdownRenderer.codeTheme != "monokai" {
		t.Errorf("expected default codeTheme to be 'monokai', got %s", markdownRenderer.codeTheme)
	}

	if len(markdownRenderer.styles) == 0 {
		t.Errorf("expected styles map to be populated")
	}
}

func TestRenderMarkdown(t *testing.T) {
	renderer := NewMarkdownRenderer()

	// Test simple markdown rendering
	markdown := "# Test Heading\nThis is a paragraph."
	result, err := renderer.RenderMarkdown(markdown)

	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v", err)
	}

	if !strings.Contains(result, "Test Heading") {
		t.Errorf("expected result to contain 'Test Heading'")
	}

	if !strings.Contains(result, "This is a paragraph.") {
		t.Errorf("expected result to contain 'This is a paragraph.'")
	}
}

func TestStyleMethods(t *testing.T) {
	renderer := NewMarkdownRenderer()

	// Test StyleHeading
	text := "Test Heading"
	styledHeading := renderer.StyleHeading(text, 1)
	if !strings.Contains(styledHeading, text) {
		t.Errorf("StyleHeading() result doesn't contain the original text")
	}

	// Test StyleInfo
	infoText := "Information"
	styledInfo := renderer.StyleInfo(infoText)
	if !strings.Contains(styledInfo, infoText) {
		t.Errorf("StyleInfo() result doesn't contain the original text")
	}

	// Test StyleWarning
	warningText := "Warning"
	styledWarning := renderer.StyleWarning(warningText)
	if !strings.Contains(styledWarning, warningText) {
		t.Errorf("StyleWarning() result doesn't contain the original text")
	}

	// Test StyleError
	errorText := "Error"
	styledError := renderer.StyleError(errorText)
	if !strings.Contains(styledError, errorText) {
		t.Errorf("StyleError() result doesn't contain the original text")
	}
}

func TestGetAvailableCodeThemes(t *testing.T) {
	renderer := NewMarkdownRenderer()
	themes := renderer.GetAvailableCodeThemes()

	if len(themes) == 0 {
		t.Errorf("GetAvailableCodeThemes() returned empty list")
	}

	// Check if "monokai" theme is present
	found := false
	for _, theme := range themes {
		if theme == "monokai" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected 'monokai' theme to be in available themes")
	}
}

func TestSetCodeTheme(t *testing.T) {
	renderer := NewMarkdownRenderer().(*MarkdownRenderer)

	// Change the theme
	newTheme := "github"
	renderer.SetCodeTheme(newTheme)

	if renderer.codeTheme != newTheme {
		t.Errorf("SetCodeTheme(%q) didn't change the theme, got %q", newTheme, renderer.codeTheme)
	}
}

func TestRenderMarkdownWithTheme(t *testing.T) {
	renderer := NewMarkdownRenderer().(*MarkdownRenderer)

	// Get original theme
	originalTheme := renderer.codeTheme

	// Test with a different theme
	differentTheme := "github"
	content := "# Test\n```\ncode\n```"

	_, err := renderer.RenderMarkdownWithTheme(content, differentTheme)
	if err != nil {
		t.Fatalf("RenderMarkdownWithTheme() error = %v", err)
	}

	// Verify theme was restored
	if renderer.codeTheme != originalTheme {
		t.Errorf("RenderMarkdownWithTheme() didn't restore original theme")
	}
}
