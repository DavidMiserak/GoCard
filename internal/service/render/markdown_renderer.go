// internal/service/render/markdown_renderer.go
package render

import (
	"fmt"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	"github.com/charmbracelet/lipgloss"
)

// MarkdownRenderer implements the RenderService interface
type MarkdownRenderer struct {
	codeTheme       string
	showLineNumbers bool
	styles          map[string]lipgloss.Style
}

// NewMarkdownRenderer creates a new renderer with default settings
func NewMarkdownRenderer() interfaces.RenderService {
	// Create styles for different elements
	styles := make(map[string]lipgloss.Style)

	// Basic styles
	styles["heading1"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E88388")).
		Bold(true)

	styles["heading2"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A8CC8C")).
		Bold(true)

	styles["heading3"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DBAB79")).
		Bold(true)

	styles["info"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#88C0D0")).
		Bold(true)

	styles["warning"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EBCB8B")).
		Bold(true)

	styles["error"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BF616A")).
		Bold(true)

	return &MarkdownRenderer{
		codeTheme:       "monokai",
		showLineNumbers: true,
		styles:          styles,
	}
}

// RenderMarkdown converts markdown text to terminal-friendly formatted text
func (r *MarkdownRenderer) RenderMarkdown(content string) (string, error) {
	// Very simple implementation that just formats headers and code blocks
	var result strings.Builder
	lines := strings.Split(content, "\n")

	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			result.WriteString(line + "\n")
			continue
		}

		if inCodeBlock {
			// Format code lines with line numbers if enabled
			if r.showLineNumbers {
				line = "    " + line
			}
			result.WriteString(line + "\n")
			continue
		}

		// Format headings
		if strings.HasPrefix(line, "# ") {
			text := strings.TrimPrefix(line, "# ")
			result.WriteString(r.styles["heading1"].Render(text) + "\n")
		} else if strings.HasPrefix(line, "## ") {
			text := strings.TrimPrefix(line, "## ")
			result.WriteString(r.styles["heading2"].Render(text) + "\n")
		} else if strings.HasPrefix(line, "### ") {
			text := strings.TrimPrefix(line, "### ")
			result.WriteString(r.styles["heading3"].Render(text) + "\n")
		} else {
			result.WriteString(line + "\n")
		}
	}

	return result.String(), nil
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
	// Return a static list of supported themes
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
	styleKey := fmt.Sprintf("heading%d", level)
	style, exists := r.styles[styleKey]
	if !exists {
		// Default to heading1 style
		style = r.styles["heading1"]
	}
	return style.Render(text)
}

// StyleInfo applies info style
func (r *MarkdownRenderer) StyleInfo(text string) string {
	return r.styles["info"].Render(fmt.Sprintf("INFO: %s", text))
}

// StyleWarning applies warning style
func (r *MarkdownRenderer) StyleWarning(text string) string {
	return r.styles["warning"].Render(fmt.Sprintf("WARNING: %s", text))
}

// StyleError applies error style
func (r *MarkdownRenderer) StyleError(text string) string {
	return r.styles["error"].Render(fmt.Sprintf("ERROR: %s", text))
}

// Ensure MarkdownRenderer implements RenderService
var _ interfaces.RenderService = (*MarkdownRenderer)(nil)
