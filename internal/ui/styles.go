// File: internal/ui/styles.go

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Common color palette
var (
	colorWhite     = lipgloss.Color("#FFFFFF")
	colorLightGray = lipgloss.Color("#888888")
	colorDarkGray  = lipgloss.Color("#444444")
	colorGreen     = lipgloss.Color("#00FF00")
	colorBlue      = lipgloss.Color("#2196F3")
)

// Title and Header Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true).
			Align(lipgloss.Center).
			Padding(1, 0, 0, 0)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			Align(lipgloss.Center).
			Padding(0, 0, 1, 0)
)

// Menu and Navigation Styles
var (
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)
)

// Statistics Screen Styles
var (
	statTitleStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorLightGray)

	activeTabStyle = tabStyle.
			Foreground(colorWhite).
			Underline(true)
)

// Browse Decks Styles
var (
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	paginationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	browseHelpStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)
)

// Study Screen Styles
var (
	studyTitleStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	cardCountStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)

	questionStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			PaddingLeft(4).
			PaddingRight(4).
			PaddingTop(2).
			PaddingBottom(2).
			Width(50).
			Align(lipgloss.Left)

	answerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			PaddingLeft(4).
			PaddingRight(4).
			PaddingTop(2).
			PaddingBottom(2).
			Width(50).
			Align(lipgloss.Left)

	revealPromptStyle = lipgloss.NewStyle().
				Foreground(colorLightGray).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorDarkGray).
				PaddingLeft(2).
				PaddingRight(2).
				PaddingTop(1).
				PaddingBottom(1).
				Align(lipgloss.Center)

	studyHelpStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)

	// Rating Colors
	ratingBlackoutColor = lipgloss.Color("#9C27B0")
	ratingWrongColor    = lipgloss.Color("#F44336")
	ratingHardColor     = lipgloss.Color("#FF9800")
	ratingGoodColor     = lipgloss.Color("#FFC107")
	ratingEasyColor     = lipgloss.Color("#4CAF50")

	// Rating Styles
	ratingBlackoutStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(ratingBlackoutColor).
				PaddingLeft(1).
				PaddingRight(1)

	ratingWrongStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(ratingWrongColor).
				PaddingLeft(1).
				PaddingRight(1)

	ratingHardStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(ratingHardColor).
			PaddingLeft(1).
			PaddingRight(1)

	ratingGoodStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(ratingGoodColor).
			PaddingLeft(1).
			PaddingRight(1)

	ratingEasyStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(ratingEasyColor).
			PaddingLeft(1).
			PaddingRight(1)

	// Progress Bar Styles
	progressBarEmptyStyle = lipgloss.NewStyle().
				Background(colorDarkGray)

	progressBarFilledStyle = lipgloss.NewStyle().
				Background(colorBlue)
)

// ViewPort Styles
var (
	viewportStyle = lipgloss.NewStyle().Padding(1, 2)
)
