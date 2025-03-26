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

func (m *DeckListModel) Init() tea.Cmd {
	return m.loadDecks
}

func (m *DeckListModel) loadDecks() tea.Msg {
	currentPath := m.RootDir
	if len(m.Breadcrumbs) > 1 {
		currentPath = m.Breadcrumbs[len(m.Breadcrumbs)-1]
	}

	subdecks, err := m.DeckService.GetSubdecks(currentPath)
	if err != nil {
		log.Printf("Error: %v", err)
		return errMsg{err}
	}

	var deckItems []DeckItem
	for _, deck := range subdecks {
		stats, err := m.DeckService.GetCardStats(deck.Path)
		if err != nil {
			log.Printf("Error: %v", err)
			continue // Skip decks with stat retrieval errors
		}

		deckItems = append(deckItems, DeckItem{
			Path:       deck.Path,
			Name:       deck.Name,
			TotalCards: stats["total"],
			DueCards:   stats["due"],
		})
	}

	m.Decks = deckItems
	return nil
}

func (m *DeckListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
				return m, m.loadDecks
			}

		case key.Matches(msg, m.Keys.Back):
			if len(m.Breadcrumbs) > 1 {
				m.Breadcrumbs = m.Breadcrumbs[:len(m.Breadcrumbs)-1]
				m.Cursor = 0
				return m, m.loadDecks
			}
		}

	case errMsg:
		// Handle errors
		log.Printf("Error: %v", msg.err)
		return m, nil
	}

	return m, nil
}

func (m *DeckListModel) View() string {
	s := strings.Builder{}

	// Breadcrumb
	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)
	s.WriteString(breadcrumbStyle.Render(strings.Join(m.Breadcrumbs, " > ")) + "\n\n")

	// Deck List
	if len(m.Decks) == 0 {
		s.WriteString("No decks found.\n")
		return s.String()
	}

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

	return s.String()
}
