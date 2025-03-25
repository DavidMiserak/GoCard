// internal/service/render/markdown_renderer.go
package render

import (
	"fmt"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
)

// MarkdownRenderer implements the RenderService interface
type MarkdownRenderer struct {
	codeTheme       string
	showLineNumbers bool
}

// NewMarkdownRenderer creates a new renderer with default settings
func NewMarkdownRenderer() interfaces.RenderService {
	return &MarkdownRenderer{
		codeTheme:       "monokai",
		showLineNumbers: true,
	}
}

// RenderMarkdown converts markdown text to terminal-friendly formatted text
// This is a minimal implementation that doesn't actually render markdown
func (r *MarkdownRenderer) RenderMarkdown(content string) (string, error) {
	// For now, we just return the content unmodified
	return content, nil
}

// RenderMarkdownWithTheme renders markdown with a specific theme
func (r *MarkdownRenderer) RenderMarkdownWithTheme(content string, theme string) (string, error) {
	// Save the current theme
	originalTheme := r.codeTheme

	// Set the new theme
	r.codeTheme = theme

	// Render the markdown
	result, err := r.RenderMarkdown(content)

	// Restore the original theme
	r.codeTheme = originalTheme

	return result, err
}

// GetAvailableCodeThemes returns a list of available syntax highlighting themes
func (r *MarkdownRenderer) GetAvailableCodeThemes() []string {
	// This is a subset of available themes in Chroma
	return []string{
		"monokai", "github", "vs", "solarized-dark", "solarized-light",
	}
}

// SetCodeTheme sets the syntax highlighting theme
func (r *MarkdownRenderer) SetCodeTheme(theme string) {
	r.codeTheme = theme
}

// EnableLineNumbers toggles line numbers in code blocks
func (r *MarkdownRenderer) EnableLineNumbers(enabled bool) {
	r.showLineNumbers = enabled
}

// StyleHeading applies heading styles
func (r *MarkdownRenderer) StyleHeading(text string, level int) string {
	// Simple implementation that just adds # characters
	prefix := strings.Repeat("#", level) + " "
	return prefix + text
}

// StyleInfo applies info style
func (r *MarkdownRenderer) StyleInfo(text string) string {
	return fmt.Sprintf("INFO: %s", text)
}

// StyleWarning applies warning style
func (r *MarkdownRenderer) StyleWarning(text string) string {
	return fmt.Sprintf("WARNING: %s", text)
}

// StyleError applies error style
func (r *MarkdownRenderer) StyleError(text string) string {
	return fmt.Sprintf("ERROR: %s", text)
}

// Ensure MarkdownRenderer implements RenderService
var _ interfaces.RenderService = (*MarkdownRenderer)(nil)
