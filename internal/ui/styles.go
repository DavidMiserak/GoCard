// File: internal/ui/styles.go

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles for statistics screen
var (
	statTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#888888"))

	activeTabStyle = tabStyle.
			Foreground(lipgloss.Color("#FFFFFF")).
			Underline(true)
)
