// File: internal/ui/browse_decks.go

package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

const (
	// Number of decks to display per page
	decksPerPage = 5
)

// Key mapping for browse screen
type browseKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Next  key.Binding
	Prev  key.Binding
	Quit  key.Binding
}

var browseKeys = browseKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"), // "k" for Vim users
		key.WithHelp("↑/k", "navigate"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"), // "j" for Vim users
		key.WithHelp("↓/j", "navigate"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "study"),
	),
	Back: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "back"),
	),
	Next: key.NewBinding(
		key.WithKeys("n", "right", "l"), // "l" for Vim users
		key.WithHelp("n/p", "next/prev page"),
	),
	Prev: key.NewBinding(
		key.WithKeys("p", "left", "h"), // "h" for Vim users
		key.WithHelp("n/p", "next/prev page"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	paginationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	browseHelpStyle = lipgloss.NewStyle(). // Changed from helpStyle to browseHelpStyle
			Foreground(lipgloss.Color("#999999"))
)

// BrowseScreen represents the browse decks screen
type BrowseScreen struct {
	store        *data.Store
	decks        []model.Deck
	cursor       int
	page         int
	totalPages   int
	width        int
	height       int
	selectedDeck string
}

// NewBrowseScreen creates a new browse screen
func NewBrowseScreen(store *data.Store) *BrowseScreen {
	decks := store.GetDecks()
	totalPages := (len(decks) + decksPerPage - 1) / decksPerPage // Ceiling division

	return &BrowseScreen{
		store:      store,
		decks:      decks,
		cursor:     0,
		page:       0,
		totalPages: totalPages,
	}
}

// Init initializes the browse screen
func (b BrowseScreen) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model
func (b BrowseScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, browseKeys.Quit):
			return b, tea.Quit

		case key.Matches(msg, browseKeys.Up):
			if b.cursor > 0 {
				b.cursor--
			}

		case key.Matches(msg, browseKeys.Down):
			// Calculate the maximum cursor position for the current page
			maxCursor := min(decksPerPage, len(b.decks)-b.page*decksPerPage) - 1
			if b.cursor < maxCursor {
				b.cursor++
			}

		case key.Matches(msg, browseKeys.Next):
			if b.page < b.totalPages-1 {
				b.page++
				b.cursor = 0
			}

		case key.Matches(msg, browseKeys.Prev):
			if b.page > 0 {
				b.page--
				b.cursor = 0
			}

		case key.Matches(msg, browseKeys.Back):
			// Return to main menu
			return NewMainMenu(), nil

		case key.Matches(msg, browseKeys.Enter):
			// Get the selected deck
			deckIndex := b.page*decksPerPage + b.cursor
			if deckIndex < len(b.decks) {
				b.selectedDeck = b.decks[deckIndex].ID
				// TODO: Navigate to study screen with the selected deck
				// return NewStudyScreen(b.store, b.selectedDeck), nil
			}
		}

	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
	}

	return b, nil
}

// View renders the browse screen
func (b BrowseScreen) View() string {
	// Title
	s := headerStyle.Render("Browse Decks")
	s += "\n\n"

	// Header row
	headerRow := fmt.Sprintf("%-20s %-10s %-10s %-15s", "DECK NAME", "CARDS", "DUE", "LAST STUDIED")
	s += headerStyle.Render(headerRow)
	s += "\n"

	// Calculate the range of decks to display on the current page
	startIdx := b.page * decksPerPage
	endIdx := min(startIdx+decksPerPage, len(b.decks))
	displayDecks := b.decks[startIdx:endIdx]

	// Display each deck
	for i, deck := range displayDecks {
		// Count due cards
		dueCards := 0
		for _, card := range deck.Cards {
			if card.NextReview.Before(time.Now()) {
				dueCards++
			}
		}

		// Format the last studied date
		lastStudied := "Never"
		if !deck.LastStudied.IsZero() {
			if isToday(deck.LastStudied) {
				lastStudied = "Today"
			} else if isYesterday(deck.LastStudied) {
				lastStudied = "Yesterday"
			} else if isWithinDays(deck.LastStudied, 7) {
				days := daysBetween(deck.LastStudied, time.Now())
				lastStudied = fmt.Sprintf("%d days ago", days)
			} else {
				lastStudied = deck.LastStudied.Format("Jan 2")
			}
		}

		// Format the row
		row := fmt.Sprintf("%-20s %-10d %-10d %-15s",
			truncate(deck.Name, 20),
			len(deck.Cards),
			dueCards,
			lastStudied)

		// Highlight the selected row
		if i == b.cursor {
			s += selectedRowStyle.Render("> " + row)
		} else {
			s += normalRowStyle.Render("  " + row)
		}
		s += "\n"
	}

	// Pagination
	s += "\n"
	pagination := fmt.Sprintf("Page %d of %d", b.page+1, b.totalPages)
	s += paginationStyle.Render(pagination)
	s += "\n\n"

	// Help text
	help := "↑/↓: Navigate" + "\tEnter: Study\t" + "b: Back" + "\tn/p: Next/Prev Page" + "\tq: Quit"
	s += browseHelpStyle.Render(help)

	return s
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func isYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

func isWithinDays(t time.Time, days int) bool {
	return time.Now().Sub(t) < time.Duration(days)*24*time.Hour
}

func daysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}
