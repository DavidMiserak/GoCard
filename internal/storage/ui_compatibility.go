// File: internal/storage/ui_compatibility.go (updated)

// Package storage implements the file-based storage system for GoCard.
package storage

import (
	"github.com/DavidMiserak/GoCard/internal/storage/parser"
)

// RenderMarkdown is a compatibility function for UI components
// that forwards to the parser package
func (s *CardStore) RenderMarkdown(content string) (string, error) {
	return parser.RenderMarkdown(content)
}

// RenderMarkdownWithTheme renders markdown with a specific syntax highlighting theme
func (s *CardStore) RenderMarkdownWithTheme(content string, theme string) (string, error) {
	config := parser.DefaultSyntaxConfig()
	config.Theme = theme
	return parser.RenderMarkdownWithHighlighting(content, config)
}

// GetAvailableSyntaxThemes returns a list of available syntax highlighting themes
func (s *CardStore) GetAvailableSyntaxThemes() []string {
	return parser.AvailableThemes()
}

// GetDefaultSyntaxTheme returns the default syntax highlighting theme
func (s *CardStore) GetDefaultSyntaxTheme() string {
	return parser.DefaultSyntaxConfig().Theme
}
