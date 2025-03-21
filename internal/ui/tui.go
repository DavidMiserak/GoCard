// File: internal/ui/tui.go
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/storage"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// keyMap defines the keybindings for the TUI
type keyMap struct {
	ShowAnswer key.Binding
	Rate0      key.Binding
	Rate1      key.Binding
	Rate2      key.Binding
	Rate3      key.Binding
	Rate4      key.Binding
	Rate5      key.Binding
	Edit       key.Binding
	New        key.Binding
	Delete     key.Binding
	Tags       key.Binding
	Search     key.Binding
	Quit       key.Binding
	Help       key.Binding
}

// ShortHelp returns keybinding help
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ShowAnswer, k.Edit, k.New, k.Quit, k.Help}
}

// FullHelp returns the full set of keybindings
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ShowAnswer, k.Rate0, k.Rate1, k.Rate2, k.Rate3, k.Rate4, k.Rate5},
		{k.Edit, k.New, k.Delete, k.Tags, k.Search},
		{k.Quit, k.Help},
	}
}

// defaultKeyMap creates the default keybindings
func defaultKeyMap() keyMap {
	return keyMap{
		ShowAnswer: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "show answer"),
		),
		Rate0: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "rate: again"),
		),
		Rate1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "rate: hard"),
		),
		Rate2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "rate: difficult"),
		),
		Rate3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "rate: good"),
		),
		Rate4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "rate: easy"),
		),
		Rate5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "rate: very easy"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit card"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new card"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete card"),
		),
		Tags: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "edit tags"),
		),
		Search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// reviewState represents the current state of the review session
type reviewState int

const (
	stateShowingQuestion reviewState = iota
	stateShowingAnswer
	stateCompleted
)

// model is the Bubble Tea model for our TUI
type model struct {
	keys        keyMap
	help        help.Model
	viewport    viewport.Model
	store       *storage.CardStore
	currentCard *card.Card
	dueCards    []*card.Card
	state       reviewState
	renderer    *glamour.TermRenderer
	reviewCount int
	showHelp    bool
	width       int
	height      int
	error       string
}

// initModel initializes the TUI model
func initModel(store *storage.CardStore) model {
	keys := defaultKeyMap()
	helpModel := help.New()
	helpModel.ShowAll = false

	// Get terminal width and height
	width, height, _ := getTerminalSize()

	// Initialize the markdown renderer
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	// Create the viewport for scrollable content
	vp := viewport.New(width, height-6) // Leave room for header and footer
	vp.SetContent("")

	// Initialize the model
	m := model{
		keys:     keys,
		help:     helpModel,
		viewport: vp,
		store:    store,
		state:    stateShowingQuestion,
		renderer: renderer,
		width:    width,
		height:   height,
		showHelp: false,
	}

	// Load due cards
	m.dueCards = store.GetDueCards()

	// Move to the first card if available
	if len(m.dueCards) > 0 {
		m.currentCard = m.dueCards[0]
		m.updateViewport()
	}

	return m
}

// getTerminalSize returns the terminal dimensions
func getTerminalSize() (width, height int, err error) {
	// Default fallback dimensions
	width, height = 80, 24

	// Try to get actual terminal size
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width, height = w, h
	}

	return width, height, nil
}

// updateViewport updates the viewport content based on the current state
func (m *model) updateViewport() {
	if m.currentCard == nil {
		if len(m.dueCards) == 0 {
			m.viewport.SetContent("No cards due for review. Great job!")
		} else {
			m.viewport.SetContent("Error: Current card is nil but due cards exist.")
		}
		return
	}

	var content strings.Builder

	// Render card metadata
	tags := strings.Join(m.currentCard.Tags, ", ")
	content.WriteString(fmt.Sprintf("Card: %s\n", m.currentCard.Title))
	content.WriteString(fmt.Sprintf("Tags: %s\n\n", tags))

	// Render question
	content.WriteString("## Question\n\n")
	questionMd, _ := m.renderer.Render(m.currentCard.Question)
	content.WriteString(questionMd)

	// Render answer if we're in answer state
	if m.state == stateShowingAnswer {
		content.WriteString("\n\n## Answer\n\n")
		answerMd, _ := m.renderer.Render(m.currentCard.Answer)
		content.WriteString(answerMd)

		// Add rating prompt
		content.WriteString("\n\nHow well did you remember? (0-5)\n")
		content.WriteString("0: Forgot completely | 3: Correct with effort | 5: Perfect recall\n")
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoTop()
}

// renderSessionSummary generates a summary of the review session
func (m *model) renderSessionSummary() {
	var content strings.Builder

	content.WriteString("# Review Session Completed\n\n")
	content.WriteString(fmt.Sprintf("Cards reviewed: %d\n", m.reviewCount))

	// Get stats from the card store
	stats := m.store.GetReviewStats()
	content.WriteString("\nCard Statistics:\n")
	content.WriteString(fmt.Sprintf("- Total cards: %d\n", stats["total_cards"]))
	content.WriteString(fmt.Sprintf("- Due cards remaining: %d\n", stats["due_cards"]))
	content.WriteString(fmt.Sprintf("- New cards: %d\n", stats["new_cards"]))
	content.WriteString(fmt.Sprintf("- Young cards (1-7 days): %d\n", stats["young_cards"]))
	content.WriteString(fmt.Sprintf("- Mature cards (>7 days): %d\n", stats["mature_cards"]))

	// Show next due date
	nextDue := m.store.GetNextDueDate()
	content.WriteString(fmt.Sprintf("\nNext review session: %s\n", nextDue.Format(time.DateOnly)))

	// Render to the viewport
	summaryMd, _ := m.renderer.Render(content.String())
	m.viewport.SetContent(summaryMd)
	m.viewport.GotoTop()
}

// moveToNextCard moves to the next card in the due cards list
func (m *model) moveToNextCard() {
	// Remove the current card from the due cards list
	for i, card := range m.dueCards {
		if card == m.currentCard {
			m.dueCards = append(m.dueCards[:i], m.dueCards[i+1:]...)
			break
		}
	}

	// Move to the next card or show completion
	if len(m.dueCards) > 0 {
		m.currentCard = m.dueCards[0]
		m.state = stateShowingQuestion
		m.updateViewport()
	} else {
		m.state = stateCompleted
		m.renderSessionSummary()
	}
}

// Init initializes the TUI
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles events and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key messages
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keys.ShowAnswer):
			if m.state == stateShowingQuestion && m.currentCard != nil {
				m.state = stateShowingAnswer
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.Rate0),
			key.Matches(msg, m.keys.Rate1),
			key.Matches(msg, m.keys.Rate2),
			key.Matches(msg, m.keys.Rate3),
			key.Matches(msg, m.keys.Rate4),
			key.Matches(msg, m.keys.Rate5):
			// Only process ratings when showing an answer
			if m.state == stateShowingAnswer && m.currentCard != nil {
				// Extract the rating from the key pressed (0-5)
				rating := 0
				switch {
				case key.Matches(msg, m.keys.Rate0):
					rating = 0
				case key.Matches(msg, m.keys.Rate1):
					rating = 1
				case key.Matches(msg, m.keys.Rate2):
					rating = 2
				case key.Matches(msg, m.keys.Rate3):
					rating = 3
				case key.Matches(msg, m.keys.Rate4):
					rating = 4
				case key.Matches(msg, m.keys.Rate5):
					rating = 5
				}

				// Apply the SM-2 algorithm to the card
				if err := m.store.ReviewCard(m.currentCard, rating); err != nil {
					m.error = fmt.Sprintf("Error reviewing card: %v", err)
				} else {
					m.reviewCount++
					m.moveToNextCard()
				}
			}

		case key.Matches(msg, m.keys.New):
			// Placeholder for new card functionality
			m.error = "Creating new cards is not implemented in this version"

		case key.Matches(msg, m.keys.Edit):
			// Placeholder for edit functionality
			m.error = "Editing cards is not implemented in this version"

		case key.Matches(msg, m.keys.Delete):
			// Placeholder for delete functionality
			m.error = "Deleting cards is not implemented in this version"

		case key.Matches(msg, m.keys.Search):
			// Placeholder for search functionality
			m.error = "Searching cards is not implemented in this version"

		case key.Matches(msg, m.keys.Tags):
			// Placeholder for tag editing functionality
			m.error = "Editing tags is not implemented in this version"
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6 // Leave room for header and footer

		// Update renderer word wrap
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(msg.Width),
		)

		// Re-render content with new dimensions
		m.updateViewport()
	}

	// Handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m model) View() string {
	var sb strings.Builder

	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("15")).
		Width(m.width).
		Align(lipgloss.Center).
		Padding(0, 1)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(m.width).
		Align(lipgloss.Center)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Width(m.width).
		Align(lipgloss.Center)

	// Render header
	if m.state == stateCompleted {
		sb.WriteString(headerStyle.Render("GoCard - Review Session Complete"))
	} else if len(m.dueCards) == 0 {
		sb.WriteString(headerStyle.Render("GoCard - No Cards Due"))
	} else {
		progress := fmt.Sprintf("%d/%d", m.reviewCount+1, m.reviewCount+len(m.dueCards))
		title := fmt.Sprintf("GoCard - Review Session - %s", progress)
		sb.WriteString(headerStyle.Render(title))
	}
	sb.WriteString("\n")

	// Render error if present
	if m.error != "" {
		sb.WriteString(errorStyle.Render(m.error))
		sb.WriteString("\n")
	}

	// Render main content area
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	// Render help or footer
	helpView := m.help.View(m.keys)
	if m.showHelp {
		sb.WriteString(helpView)
	} else {
		// Show a simple footer with basic instructions
		if m.state == stateShowingQuestion {
			sb.WriteString(footerStyle.Render("Press space to show answer, ? for help"))
		} else if m.state == stateShowingAnswer {
			sb.WriteString(footerStyle.Render("Rate 0-5, ? for help"))
		} else {
			sb.WriteString(footerStyle.Render("Press q to quit, ? for help"))
		}
	}

	return sb.String()
}

// RunTUI starts the terminal UI
func RunTUI(store *storage.CardStore) error {
	m := initModel(store)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
