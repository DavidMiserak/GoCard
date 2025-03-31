// File: internal/ui/menu.go

package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DavidMiserak/GoCard/internal/data"
)

// Define styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(1, 0, 0, 0)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Align(lipgloss.Center).
			Padding(0, 0, 1, 0)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

// Define key mappings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Quit  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"), // "k" for Vim users
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"), // "j" for Vim users
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// MainMenu represents the main menu model
type MainMenu struct {
	items    []string
	cursor   int
	selected int
	width    int
	height   int
}

// NewMainMenu creates a new main menu
func NewMainMenu() *MainMenu {
	return &MainMenu{
		items:    []string{"Study", "Browse Decks", "Statistics", "Quit"},
		cursor:   0,
		selected: -1,
	}
}

// Init initializes the main menu
func (m MainMenu) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model
func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Enter):
			m.selected = m.cursor

			// Handle menu selection
			switch m.cursor {
			case 0: // Study
				// Navigate to study screen
				return NewBrowseScreen(data.NewStore()), nil

			case 1: // Browse Decks
				// Navigate to browse decks screen
				return NewBrowseScreen(data.NewStore()), nil

			case 2: // Statistics
				// TODO: Navigate to statistics screen
				return m, nil

			case 3: // Quit
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the main menu
func (m MainMenu) View() string {
	// Title and subtitle
	s := titleStyle.Render("GoCard")
	s += "\n" + subtitleStyle.Render("Terminal Flashcards")
	s += "\n\n"

	// Menu items
	for i, item := range m.items {
		if i == m.cursor {
			s += selectedItemStyle.Render("> " + item)
		} else {
			s += itemStyle.Render("  " + item)
		}
		s += "\n"
	}

	// Help
	s += "\n" + helpStyle.Render("↑/↓: Navigate"+"\tEnter: Select"+"\tq: Quit")

	return s
}
