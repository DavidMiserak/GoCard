// internal/ui/tui/deck_list_view.go

package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *DeckListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Down):
			m.Cursor = min(m.Cursor+1, len(m.Decks)-1)

		case key.Matches(msg, m.Keys.Up):
			m.Cursor = max(m.Cursor-1, 0)

		case key.Matches(msg, m.Keys.Enter):
			if len(m.Decks) > 0 {
				selectedDeck := m.Decks[m.Cursor]
				m.Breadcrumbs = append(m.Breadcrumbs, selectedDeck.Path)
				return m, m.loadDecks()
			}

		case key.Matches(msg, m.Keys.Back):
			if len(m.Breadcrumbs) > 1 {
				m.Breadcrumbs = m.Breadcrumbs[:len(m.Breadcrumbs)-1]
				m.Cursor = 0
				return m, m.loadDecks()
			}
		}

	case errMsg:
		log.Printf("Error: %v", msg.err)
		return m, nil
	}

	return m, nil
}

func (m *DeckListModel) View() string {
	s := strings.Builder{}

	// Fallback terminal size if not set
	width := m.TerminalWidth
	if width == 0 {
		width = 80
	}
	height := m.TerminalHeight
	if height == 0 {
		height = 24
	}

	// Terminal size styles
	baseStyle := lipgloss.NewStyle().
		Width(width).
		Height(height)

	// Breadcrumb
	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		Width(width)
	s.WriteString(breadcrumbStyle.Render(strings.Join(m.Breadcrumbs, " > ")) + "\n\n")

	// No Decks View
	if len(m.Decks) == 0 {
		welcomeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Width(width).
			Align(lipgloss.Center)

		welcomeMessage := "Welcome to GoCard!\n\n" +
			"We've created some default decks to help you get started.\n" +
			"Use arrow keys to navigate, 'Enter' to select, and 'q' to quit."

		s.WriteString(welcomeStyle.Render(welcomeMessage))
		return baseStyle.Render(s.String())
	}

	// Deck List
	for i, deck := range m.Decks {
		// Cursor styling
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		// Deck name styling
		nameStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))
		if m.Cursor != i {
			nameStyle = nameStyle.Foreground(lipgloss.Color("245"))
		}

		// Deck info styling
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

		deckLine := fmt.Sprintf(
			"%s %s  %s\n",
			cursor,
			nameStyle.Render(deck.Name),
			infoStyle.Render(fmt.Sprintf("(%d cards, %d due)", deck.TotalCards, deck.DueCards)),
		)
		s.WriteString(deckLine)
	}

	return baseStyle.Render(s.String())
}
