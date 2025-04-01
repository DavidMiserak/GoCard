// File: internal/ui/markdown_renderer.go

package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer handles rendering Markdown text to styled terminal output
type MarkdownRenderer struct {
	renderer      *glamour.TermRenderer
	renderedCache map[string]string
	defaultWidth  int
	syntaxTheme   string
}

// NewMarkdownRenderer creates a new markdown renderer with specified width and theme
func NewMarkdownRenderer(width int, themeName string) *MarkdownRenderer {
	// Use default width if not specified
	if width <= 0 {
		width = 80
	}

	// Validate and set theme, fallback to "monokai" if invalid
	if themeName == "" {
		themeName = "monokai"
	}

	// Initialize glamour renderer with explicit style
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)

	return &MarkdownRenderer{
		renderer:      renderer,
		renderedCache: make(map[string]string),
		defaultWidth:  width,
		syntaxTheme:   themeName,
	}
}

// renderCodeBlock uses Chroma to syntax highlight code blocks
func renderCodeBlock(code, language, themeName string) string {
	// Determine the lexer based on the language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use specified theme
	style := styles.Get(themeName)
	if style == nil {
		style = styles.Fallback
	}

	// Create a terminal formatter
	formatter := formatters.Get("terminal")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code // Fallback to original code if tokenization fails
	}

	// Render the highlighted code
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code // Fallback to original code if formatting fails
	}

	return buf.String()
}

// UpdateWidth updates the renderer's width and clears the cache
func (r *MarkdownRenderer) UpdateWidth(width int) {
	if width <= 0 {
		return
	}

	// Create a new renderer with the updated width
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)

	r.renderer = renderer
	r.defaultWidth = width

	// Clear the cache because we need to re-render with new width
	r.renderedCache = make(map[string]string)
}

// SetSyntaxTheme allows changing the syntax highlighting theme
func (r *MarkdownRenderer) SetSyntaxTheme(themeName string) {
	// Validate theme
	if themeName == "" {
		themeName = "monokai"
	}

	// Update renderer with new theme
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(r.defaultWidth),
		glamour.WithEmoji(),
	)

	r.renderer = renderer
	r.syntaxTheme = themeName

	// Clear cache to force re-rendering with new theme
	r.renderedCache = make(map[string]string)
}

// Render renders markdown text to terminal output
func (r *MarkdownRenderer) Render(markdown string) string {
	// Create cache key using content, width, and theme to ensure proper rendering
	cacheKey := fmt.Sprintf("%s-%d-%s", markdown, r.defaultWidth, r.syntaxTheme)

	// Check if we already have this content rendered in the cache
	if rendered, exists := r.renderedCache[cacheKey]; exists {
		return rendered
	}

	// Not in cache, render it now
	if r.renderer == nil {
		// Fallback if renderer isn't initialized
		return markdown
	}

	rendered, err := r.renderer.Render(markdown)
	if err != nil {
		// Return the original text if rendering fails
		return markdown
	}

	// Trim extra whitespace that glamour might add
	rendered = strings.TrimSpace(rendered)

	// Store in cache for future use
	r.renderedCache[cacheKey] = rendered

	return rendered
}

// ClearCache clears the rendering cache
func (r *MarkdownRenderer) ClearCache() {
	r.renderedCache = make(map[string]string)
}
