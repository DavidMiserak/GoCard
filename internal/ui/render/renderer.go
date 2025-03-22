// File: internal/ui/render/renderer.go (updated)
package render

import (
	"github.com/charmbracelet/glamour"
)

// Renderer handles markdown rendering and styling
type Renderer struct {
	mdRenderer    *glamour.TermRenderer
	styles        Styles
	width         int
	syntaxEnabled bool // Whether syntax highlighting is enabled
}

// NewRenderer creates a new renderer with the given width
func NewRenderer(width int) (*Renderer, error) {
	// Create a glamour renderer with default settings
	mdRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
		// We don't need to specify additional options here as our syntax
		// highlighting is handled by the storage/parser package before
		// the content reaches the UI renderer
	)

	if err != nil {
		return nil, err
	}

	return &Renderer{
		mdRenderer:    mdRenderer,
		styles:        DefaultStyles(),
		width:         width,
		syntaxEnabled: true, // Enable syntax highlighting by default
	}, nil
}

// RenderMarkdown renders markdown content to terminal-friendly output
func (r *Renderer) RenderMarkdown(content string) (string, error) {
	// Let Glamour handle the rendering - it can process both regular markdown
	// and HTML produced by our syntax highlighter
	return r.mdRenderer.Render(content)
}

// UpdateWidth updates the renderer's width and recreates the markdown renderer
func (r *Renderer) UpdateWidth(width int) error {
	r.width = width

	mdRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	if err != nil {
		return err
	}

	r.mdRenderer = mdRenderer
	return nil
}

// HeaderStyle returns the style for headers
func (r *Renderer) HeaderStyle(text string) string {
	return r.styles.Header.Width(r.width).Render(text)
}

// FooterStyle returns the style for footers
func (r *Renderer) FooterStyle(text string) string {
	return r.styles.Footer.Width(r.width).Render(text)
}

// ErrorStyle returns the style for error messages
func (r *Renderer) ErrorStyle(text string) string {
	return r.styles.Error.Width(r.width).Render(text)
}

// InputStyle returns the style for input prompts
func (r *Renderer) InputStyle(text string) string {
	return r.styles.Input.Width(r.width).Render(text)
}

// GetStyles returns the current styles
func (r *Renderer) GetStyles() Styles {
	return r.styles
}

// SetStyles updates the renderer's styles
func (r *Renderer) SetStyles(styles Styles) {
	r.styles = styles
}

// SetSyntaxHighlighting enables or disables syntax highlighting
func (r *Renderer) SetSyntaxHighlighting(enabled bool) {
	r.syntaxEnabled = enabled
}

// IsSyntaxHighlightingEnabled returns whether syntax highlighting is enabled
func (r *Renderer) IsSyntaxHighlightingEnabled() bool {
	return r.syntaxEnabled
}
