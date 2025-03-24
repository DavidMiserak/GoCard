// File: internal/ui/render/style.go

package render

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all UI styling definitions
type Styles struct {
	Header     lipgloss.Style
	Footer     lipgloss.Style
	Error      lipgloss.Style
	Input      lipgloss.Style
	Highlight  lipgloss.Style
	Subtle     lipgloss.Style
	DimmedText lipgloss.Style
	Title      lipgloss.Style
	Question   lipgloss.Style
	Answer     lipgloss.Style
}

// DefaultStyles returns the default styling for the application
func DefaultStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("15")).
			Align(lipgloss.Center).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Align(lipgloss.Center),

		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		DimmedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			MarginBottom(1),

		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),

		Answer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("76")).
			Bold(true),
	}
}

// DarkTheme returns a dark theme for the application
func DarkTheme() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Align(lipgloss.Center).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Align(lipgloss.Center),

		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("50")).
			Bold(true),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		DimmedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("50")).
			Bold(true).
			MarginBottom(1),

		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color("50")).
			Bold(true),

		Answer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("48")).
			Bold(true),
	}
}

// LightTheme returns a light theme for the application
func LightTheme() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("232")).
			Background(lipgloss.Color("153")).
			Align(lipgloss.Center).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Align(lipgloss.Center),

		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(lipgloss.Color("253")).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("21")).
			Bold(true),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		DimmedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Italic(true),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("21")).
			Bold(true).
			MarginBottom(1),

		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color("21")).
			Bold(true),

		Answer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")).
			Bold(true),
	}
}
