// internal/service/interfaces/render_service.go
package interfaces

// RenderService handles rendering of content for display
type RenderService interface {
	// Markdown rendering
	RenderMarkdown(content string) (string, error)
	RenderMarkdownWithTheme(content string, theme string) (string, error)

	// Code syntax highlighting
	GetAvailableCodeThemes() []string
	SetCodeTheme(theme string)
	EnableLineNumbers(enabled bool)

	// UI styling
	StyleHeading(text string, level int) string
	StyleInfo(text string) string
	StyleWarning(text string) string
	StyleError(text string) string
}
