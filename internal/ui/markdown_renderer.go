// File: internal/ui/markdown_renderer.go

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer handles rendering Markdown text to styled terminal output
type MarkdownRenderer struct {
	renderer      *glamour.TermRenderer
	renderedCache map[string]string
	defaultWidth  int
}

// NewMarkdownRenderer creates a new markdown renderer with specified width
func NewMarkdownRenderer(width int) *MarkdownRenderer {
	// Use default width if not specified
	if width <= 0 {
		width = 80
	}

	// Initialize glamour renderer with explicit style instead of auto-detection
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes(defaultMarkdownStyle()),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)

	return &MarkdownRenderer{
		renderer:      renderer,
		renderedCache: make(map[string]string),
		defaultWidth:  width,
	}
}

// UpdateWidth updates the renderer's width and clears the cache
func (r *MarkdownRenderer) UpdateWidth(width int) {
	if width <= 0 {
		return
	}

	// Create a new renderer with the updated width
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes(defaultMarkdownStyle()),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)

	r.renderer = renderer
	r.defaultWidth = width

	// Clear the cache because we need to re-render with new width
	r.renderedCache = make(map[string]string)
}

// Render renders markdown text to terminal output
func (r *MarkdownRenderer) Render(markdown string) string {
	// Create cache key using content and width to ensure proper rendering after resizes
	cacheKey := fmt.Sprintf("%s-%d", markdown, r.defaultWidth)

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

// defaultMarkdownStyle returns a JSON byte array with the default styling configuration
func defaultMarkdownStyle() []byte {
	return []byte(`{
		"document": {},
		"block_quote": {
			"indent": 1,
			"indent_token": "│ "
		},
		"paragraph": {},
		"list": {
			"level_indent": 2
		},
		"heading": {
			"level_1": {
				"prefix": "# ",
				"bold": true
			},
			"level_2": {
				"prefix": "## ",
				"bold": true
			},
			"level_3": {
				"prefix": "### ",
				"bold": true
			},
			"level_4": {
				"prefix": "#### ",
				"bold": true
			},
			"level_5": {
				"prefix": "##### ",
				"bold": true
			},
			"level_6": {
				"prefix": "###### ",
				"bold": true
			}
		},
		"code_block": {
			"theme": "monokai"
		},
		"html_block": {},
		"thematic_break": {
			"border": "─"
		},
		"text": {},
		"emphasis": {
			"italic": true
		},
		"strong": {
			"bold": true
		},
		"link": {
			"underline": true,
			"color": "blue"
		},
		"code": {
			"inline": true,
			"border": true
		}
	}`)
}
