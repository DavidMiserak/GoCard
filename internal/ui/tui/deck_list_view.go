// internal/ui/tui/deck_list_view.go

package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// ReturnToDeckListMsg is sent to return to the deck list after a review
type ReturnToDeckListMsg struct{}

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
			m.navigateDown()
			return m, nil

		case key.Matches(msg, m.Keys.Up):
			m.navigateUp()
			return m, nil

		case key.Matches(msg, m.Keys.Enter):
			return m, m.enterDeck()

		case key.Matches(msg, m.Keys.Back):
			return m, m.navigateBack()

		case key.Matches(msg, m.Keys.Study):
			return m, m.startReview()

		case key.Matches(msg, m.Keys.Refresh):
			return m, m.loadDecks()
		}

	case NoDueCardsMsg:
		// Show a message that no cards are due
		log.Printf("No due cards in deck: %s", msg.DeckName)
		return m, nil

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
		Height(height - 2) // Leave space for help

	// Header styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Width(width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	// Breadcrumb styles
	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	breadcrumbArrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Join breadcrumbs with arrows
	breadcrumbText := ""
	for i, crumb := range m.Breadcrumbs {
		if i > 0 {
			breadcrumbText += breadcrumbArrowStyle.Render(" > ")
		}
		breadcrumbText += breadcrumbStyle.Render(crumb)
	}

	// Header with title and breadcrumbs
	s.WriteString(headerStyle.Render(
		fmt.Sprintf("GoCard - %s", breadcrumbText)))

	// No Decks View
	if len(m.Decks) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Width(width).
			Align(lipgloss.Center).
			PaddingTop(1).
			PaddingBottom(1)

		emptyMsg := "No decks found in this location"

		// If we're at the root and there are no decks, show a different message
		if len(m.Breadcrumbs) == 1 {
			emptyMsg = "Welcome to GoCard!\n\n" +
				"We'll create some default decks to help you get started.\n" +
				"Press 'r' to refresh the view after they're created."
		}

		s.WriteString(emptyStyle.Render(emptyMsg))
	} else {
		// Deck List
		listStyle := lipgloss.NewStyle().Width(width)

		// Table header style
		tableHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")).
			PaddingBottom(1)

		// Table header
		tableHeader := fmt.Sprintf(
			"%-25s %-10s %-10s %-10s",
			"Deck Name", "Total", "Due", "New")

		s.WriteString(tableHeaderStyle.Render(tableHeader))
		s.WriteString("\n")

		for i, deck := range m.Decks {
			// Cursor styling
			cursor := " "
			if m.Cursor == i {
				cursor = "▶"
			}

			// Deck name styling
			nameStyle := lipgloss.NewStyle().Width(25)
			if m.Cursor == i {
				nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("39"))
			} else {
				nameStyle = nameStyle.Foreground(lipgloss.Color("252"))
			}

			// Stats styling
			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Width(10)

			// Due cards styling - highlight if due > 0
			dueStyle := statsStyle
			if deck.DueCards > 0 {
				dueStyle = dueStyle.Bold(true).Foreground(lipgloss.Color("205"))
			}

			// Format the deck row
			deckRow := fmt.Sprintf("%s %s %s %s %s",
				cursor,
				nameStyle.Render(deck.Name),
				statsStyle.Render(fmt.Sprintf("%d", deck.TotalCards)),
				dueStyle.Render(fmt.Sprintf("%d", deck.DueCards)),
				statsStyle.Render(fmt.Sprintf("%d", deck.NewCards)),
			)
			s.WriteString(listStyle.Render(deckRow))
			s.WriteString("\n")
		}
	}

	// Help text
	helpModel := help.New()
	helpModel.Width = width

	// Help section styles
	helpStyle := lipgloss.NewStyle().
		PaddingTop(1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpText := helpStyle.Render(helpModel.View(m.Keys))

	// Return the complete view
	return baseStyle.Render(s.String()) + "\n" + helpText
}
