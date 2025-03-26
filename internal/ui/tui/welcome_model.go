// internal/ui/tui/welcome_model.go

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeModel represents the welcome screen
type WelcomeModel struct {
	ready          bool
	width          int
	height         int
	keys           WelcomeKeyMap
	stats          map[string]int
	deckCount      int
	totalCardCount int
	dueCardCount   int
	newCardCount   int
	reviewedCount  int
}

// NewWelcomeModel creates a new welcome model
func NewWelcomeModel() *WelcomeModel {
	return &WelcomeModel{
		keys: DefaultWelcomeKeyMap(),
		stats: map[string]int{
			"decks":    0,
			"cards":    0,
			"due":      0,
			"new":      0,
			"reviewed": 0,
		},
	}
}

// Init initializes the welcome model
func (m *WelcomeModel) Init() tea.Cmd {
	return nil
}

// Update handles events and messages for the welcome model
func (m *WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			// Signal to switch to deck list
			return m, func() tea.Msg {
				return SwitchScreenMsg{Screen: ScreenDeckList}
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the welcome model
func (m *WelcomeModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var s strings.Builder

	// Calculate available space
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Width(width).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Width(width).
		Align(lipgloss.Center)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(width).
		Align(lipgloss.Center)

	statsStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		PaddingTop(1).
		PaddingBottom(1)

	statBlockStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 2).
		Margin(0, 1).
		Align(lipgloss.Center)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(width).
		Align(lipgloss.Center).
		PaddingTop(1)

	// Title section
	s.WriteString(titleStyle.Render("Welcome to GoCard"))
	s.WriteString("\n")
	s.WriteString(subtitleStyle.Render("Your Terminal-Based Spaced Repetition System"))
	s.WriteString("\n\n")
	s.WriteString(infoStyle.Render("Efficient learning with flashcards in your terminal"))
	s.WriteString("\n\n")

	// Stats section
	if m.deckCount > 0 {
		// Display statistics in a row using flexbox-style layout
		var statBlocks []string

		statBlocks = append(statBlocks, statBlockStyle.Render(fmt.Sprintf(
			"Decks\n%d", m.deckCount)))

		statBlocks = append(statBlocks, statBlockStyle.Render(fmt.Sprintf(
			"Cards\n%d", m.totalCardCount)))

		statBlocks = append(statBlocks, statBlockStyle.Render(fmt.Sprintf(
			"Due\n%d", m.dueCardCount)))

		statBlocks = append(statBlocks, statBlockStyle.Render(fmt.Sprintf(
			"New\n%d", m.newCardCount)))

		statBlocks = append(statBlocks, statBlockStyle.Render(fmt.Sprintf(
			"Reviewed\n%d", m.reviewedCount)))

		s.WriteString(statsStyle.Render(lipgloss.JoinHorizontal(
			lipgloss.Center, statBlocks...)))
	} else {
		// No decks yet - welcome message
		welcomeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Width(width).
			Align(lipgloss.Center)

		s.WriteString(welcomeStyle.Render(
			"\nGet started by creating some flashcards.\n" +
				"Basic decks will be created automatically on first run.\n"))
	}

	// Instructions
	s.WriteString("\n\n")
	s.WriteString(instructionStyle.Render("Press Enter to browse decks"))
	s.WriteString("\n")
	s.WriteString(instructionStyle.Render("Press q to quit"))

	// Center the entire view
	fullViewStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return fullViewStyle.Render(s.String())
}

// SetStats updates the statistics in the welcome model
func (m *WelcomeModel) SetStats(deckCount, totalCards, dueCards, newCards, reviewedCards int) {
	m.deckCount = deckCount
	m.totalCardCount = totalCards
	m.dueCardCount = dueCards
	m.newCardCount = newCards
	m.reviewedCount = reviewedCards
}
