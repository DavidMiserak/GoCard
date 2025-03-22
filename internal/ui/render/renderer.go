// Package render handles UI rendering and styling.
package render

import (
	"github.com/charmbracelet/glamour"
)

// Renderer handles markdown rendering and styling
type Renderer struct {
	mdRenderer *glamour.TermRenderer
	styles     Styles
	width      int
}

// NewRenderer creates a new renderer with the given width
func NewRenderer(width int) (*Renderer, error) {
	mdRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	if err != nil {
		return nil, err
	}

	return &Renderer{
		mdRenderer: mdRenderer,
		styles:     DefaultStyles(),
		width:      width,
	}, nil
}

// RenderMarkdown renders markdown content to terminal-friendly output
func (r *Renderer) RenderMarkdown(content string) (string, error) {
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
