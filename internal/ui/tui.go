// File: internal/ui/tui.go
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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
	ChangeDeck key.Binding
	CreateDeck key.Binding
	RenameDeck key.Binding
	DeleteDeck key.Binding
	MoveToDeck key.Binding
	Quit       key.Binding
	Help       key.Binding
	Back       key.Binding
}

// ShortHelp returns keybinding help
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ShowAnswer, k.ChangeDeck, k.Edit, k.New, k.Quit, k.Help}
}

// FullHelp returns the full set of keybindings
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ShowAnswer, k.Rate0, k.Rate1, k.Rate2, k.Rate3, k.Rate4, k.Rate5},
		{k.ChangeDeck, k.CreateDeck, k.RenameDeck, k.DeleteDeck, k.MoveToDeck},
		{k.Edit, k.New, k.Delete, k.Tags, k.Search},
		{k.Back, k.Quit, k.Help},
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
		ChangeDeck: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "change deck"),
		),
		CreateDeck: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "create deck"),
		),
		RenameDeck: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename deck"),
		),
		DeleteDeck: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete deck"),
		),
		MoveToDeck: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move to deck"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "go back"),
		),
	}
}

// viewState represents the current view in the TUI
type viewState int

const (
	viewReview viewState = iota
	viewDeckList
	viewDeckBrowser
	viewDeckStats
	viewCreateDeck
	viewRenameDeck
	viewDeleteDeck
	viewMoveToDeck
	viewCreateCard
	viewEditCard
	viewSearch
	viewSearchResults
)

// model is the Bubble Tea model for our TUI
type model struct {
	keys           keyMap
	help           help.Model
	viewport       viewport.Model
	store          *storage.CardStore
	currentDeck    *deck.Deck
	currentCard    *card.Card
	dueCards       []*card.Card
	reviewState    reviewState
	viewState      viewState
	previousState  viewState
	renderer       *glamour.TermRenderer
	reviewCount    int
	showHelp       bool
	width          int
	height         int
	error          string
	textInput      textinput.Model
	inputLabel     string
	searchResults  []*card.Card
	deckListOffset int
	selectedDeck   *deck.Deck
}

// reviewState represents the current state of the review session
type reviewState int

const (
	stateShowingQuestion reviewState = iota
	stateShowingAnswer
	stateCompleted
)

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

	// Initialize text input for various input operations
	ti := textinput.New()
	ti.Placeholder = "Enter text here..."
	ti.CharLimit = 100
	ti.Width = 40

	// Initialize the model
	m := model{
		keys:          keys,
		help:          helpModel,
		viewport:      vp,
		store:         store,
		currentDeck:   store.RootDeck,
		reviewState:   stateShowingQuestion,
		viewState:     viewDeckBrowser,
		previousState: viewDeckBrowser,
		renderer:      renderer,
		width:         width,
		height:        height,
		showHelp:      false,
		textInput:     ti,
	}

	// Load due cards for the current deck
	m.dueCards = store.GetDueCardsInDeck(m.currentDeck)

	// Move to the first card if available
	if len(m.dueCards) > 0 {
		m.currentCard = m.dueCards[0]
		m.updateViewport()
	} else {
		m.updateDeckBrowser()
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
	if m.viewState == viewReview {
		m.updateReviewViewport()
	} else if m.viewState == viewDeckBrowser {
		m.updateDeckBrowser()
	} else if m.viewState == viewDeckList {
		m.updateDeckList()
	} else if m.viewState == viewDeckStats {
		m.updateDeckStats()
	} else if m.viewState == viewSearchResults {
		m.updateSearchResults()
	}
}

// updateReviewViewport updates the viewport for the review view
func (m *model) updateReviewViewport() {
	if m.currentCard == nil {
		if len(m.dueCards) == 0 {
			m.viewport.SetContent("No cards due for review in this deck. Great job!")
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
	if m.reviewState == stateShowingAnswer {
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

// updateDeckBrowser updates the viewport for the deck browser view
func (m *model) updateDeckBrowser() {
	var content strings.Builder

	// Render deck information
	content.WriteString(fmt.Sprintf("# Deck: %s\n\n", m.currentDeck.FullName()))

	// Render subdeck list
	if len(m.currentDeck.SubDecks) > 0 {
		content.WriteString("## Subdecks\n\n")
		for _, subDeck := range m.currentDeck.SubDecks {
			stats := m.store.GetDeckStats(subDeck)
			content.WriteString(fmt.Sprintf("- %s (%d cards, %d due)\n",
				subDeck.Name,
				stats["total_cards"],
				stats["due_cards"]))
		}
		content.WriteString("\n")
	}

	// Get deck statistics
	stats := m.store.GetDeckStats(m.currentDeck)
	content.WriteString("## Statistics\n\n")
	content.WriteString(fmt.Sprintf("- Total cards: %d\n", stats["total_cards"]))
	content.WriteString(fmt.Sprintf("- Due cards: %d\n", stats["due_cards"]))
	content.WriteString(fmt.Sprintf("- New cards: %d\n", stats["new_cards"]))
	content.WriteString(fmt.Sprintf("- Young cards (1-7 days): %d\n", stats["young_cards"]))
	content.WriteString(fmt.Sprintf("- Mature cards (>7 days): %d\n", stats["mature_cards"]))
	content.WriteString("\n")

	// Show actions available
	content.WriteString("## Actions\n\n")
	if stats["due_cards"].(int) > 0 {
		content.WriteString("- Press space to start review session\n")
	}
	content.WriteString("- Press 'c' to change deck\n")
	content.WriteString("- Press 'C' to create a new deck\n")
	content.WriteString("- Press 'n' to create a new card\n")
	content.WriteString("- Press 's' to search cards\n")

	// Add cards list if there are cards in this deck
	if len(m.currentDeck.Cards) > 0 {
		content.WriteString("\n## Cards in this deck\n\n")
		for _, c := range m.currentDeck.Cards {
			content.WriteString(fmt.Sprintf("- %s\n", c.Title))
		}
	}

	contentMd, _ := m.renderer.Render(content.String())
	m.viewport.SetContent(contentMd)
	m.viewport.GotoTop()
}

// updateDeckList updates the viewport for the deck list view
func (m *model) updateDeckList() {
	var content strings.Builder

	content.WriteString("# Select a Deck\n\n")

	// Get all decks
	allDecks := m.store.RootDeck.AllDecks()

	// Display decks with indentation based on depth
	for i, d := range allDecks {
		// Skip decks before the offset
		if i < m.deckListOffset {
			continue
		}

		// Limit the display to fit the viewport
		if i >= m.deckListOffset+m.height-10 {
			content.WriteString("\n(More decks below...)\n")
			break
		}

		// Calculate depth for indentation
		depth := 0
		parent := d.ParentDeck
		for parent != nil {
			depth++
			parent = parent.ParentDeck
		}

		// Add selection indicator
		prefix := "  "
		if d == m.selectedDeck {
			prefix = "> "
		}

		// Add indentation based on depth
		indent := strings.Repeat("  ", depth)
		stats := m.store.GetDeckStats(d)
		content.WriteString(fmt.Sprintf("%s%s%s (%d cards, %d due)\n",
			prefix, indent, d.Name,
			stats["total_cards"],
			stats["due_cards"]))
	}

	// Add navigation instructions
	content.WriteString("\nUse arrow keys to navigate, Enter to select, Esc to cancel\n")

	m.viewport.SetContent(content.String())
	m.viewport.GotoTop()
}

// updateDeckStats updates the viewport for the deck statistics view
func (m *model) updateDeckStats() {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Deck Statistics: %s\n\n", m.currentDeck.FullName()))

	// Get deck statistics
	stats := m.store.GetDeckStats(m.currentDeck)

	content.WriteString("## Overview\n\n")
	content.WriteString(fmt.Sprintf("- Total cards: %d\n", stats["total_cards"]))
	content.WriteString(fmt.Sprintf("- Cards in this deck: %d\n", stats["direct_cards"]))
	content.WriteString(fmt.Sprintf("- Subdecks: %d\n", stats["sub_decks"]))
	content.WriteString("\n")

	content.WriteString("## Review Status\n\n")
	content.WriteString(fmt.Sprintf("- Due cards: %d\n", stats["due_cards"]))
	content.WriteString(fmt.Sprintf("- New cards: %d\n", stats["new_cards"]))
	content.WriteString(fmt.Sprintf("- Young cards (1-7 days): %d\n", stats["young_cards"]))
	content.WriteString(fmt.Sprintf("- Mature cards (>7 days): %d\n", stats["mature_cards"]))
	content.WriteString("\n")

	// Add tag information if there are cards
	if stats["total_cards"].(int) > 0 {
		// Count tags in this deck
		tagCount := make(map[string]int)
		allCards := m.currentDeck.GetAllCards()
		for _, c := range allCards {
			for _, tag := range c.Tags {
				tagCount[tag]++
			}
		}

		if len(tagCount) > 0 {
			content.WriteString("## Tags\n\n")
			for tag, count := range tagCount {
				content.WriteString(fmt.Sprintf("- %s: %d cards\n", tag, count))
			}
		}
	}

	contentMd, _ := m.renderer.Render(content.String())
	m.viewport.SetContent(contentMd)
	m.viewport.GotoTop()
}

// updateSearchResults updates the viewport for search results
func (m *model) updateSearchResults() {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Search Results: %d found\n\n", len(m.searchResults)))

	if len(m.searchResults) == 0 {
		content.WriteString("No cards match your search criteria.\n")
	} else {
		for i, card := range m.searchResults {
			// Get the relative path to show which deck the card is in
			dirPath := filepath.Dir(card.FilePath)
			var deckPath string
			for path, d := range m.store.Decks {
				if path == dirPath {
					deckPath = d.FullName()
					break
				}
			}

			content.WriteString(fmt.Sprintf("## %d. %s\n", i+1, card.Title))
			content.WriteString(fmt.Sprintf("Deck: %s\n", deckPath))
			content.WriteString(fmt.Sprintf("Tags: %s\n\n", strings.Join(card.Tags, ", ")))

			// Show a preview of the question
			preview := card.Question
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			content.WriteString(fmt.Sprintf("Question: %s\n\n", preview))
		}
	}

	contentMd, _ := m.renderer.Render(content.String())
	m.viewport.SetContent(contentMd)
	m.viewport.GotoTop()
}

// renderSessionSummary generates a summary of the review session
func (m *model) renderSessionSummary() {
	var content strings.Builder

	content.WriteString("# Review Session Completed\n\n")
	content.WriteString(fmt.Sprintf("Cards reviewed: %d\n", m.reviewCount))

	// Get stats from the card store for the current deck
	stats := m.store.GetDeckStats(m.currentDeck)
	content.WriteString("\nDeck Statistics:\n")
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
		m.reviewState = stateShowingQuestion
		m.updateViewport()
	} else {
		m.reviewState = stateCompleted
		m.renderSessionSummary()
	}
}

// startReviewSession starts a review session for the current deck
func (m *model) startReviewSession() {
	// Load due cards for the current deck
	m.dueCards = m.store.GetDueCardsInDeck(m.currentDeck)

	if len(m.dueCards) > 0 {
		m.viewState = viewReview
		m.reviewState = stateShowingQuestion
		m.currentCard = m.dueCards[0]
		m.reviewCount = 0
		m.updateViewport()
	} else {
		m.error = "No cards due for review in this deck"
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

		case key.Matches(msg, m.keys.Back):
			if m.viewState != viewDeckBrowser {
				m.viewState = viewDeckBrowser
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.ShowAnswer):
			if m.viewState == viewReview && m.reviewState == stateShowingQuestion && m.currentCard != nil {
				m.reviewState = stateShowingAnswer
				m.updateViewport()
			} else if m.viewState == viewDeckBrowser {
				// Start a review session when pressing space in deck browser
				m.startReviewSession()
			}

		case key.Matches(msg, m.keys.Rate0),
			key.Matches(msg, m.keys.Rate1),
			key.Matches(msg, m.keys.Rate2),
			key.Matches(msg, m.keys.Rate3),
			key.Matches(msg, m.keys.Rate4),
			key.Matches(msg, m.keys.Rate5):
			// Only process ratings when showing an answer
			if m.viewState == viewReview && m.reviewState == stateShowingAnswer && m.currentCard != nil {
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

		case key.Matches(msg, m.keys.ChangeDeck):
			if m.viewState != viewDeckList {
				m.previousState = m.viewState
				m.viewState = viewDeckList
				m.selectedDeck = m.currentDeck
				m.deckListOffset = 0
				m.updateDeckList()
			}

		case key.Matches(msg, m.keys.CreateDeck):
			if m.viewState == viewDeckBrowser {
				m.viewState = viewCreateDeck
				m.textInput.Focus()
				m.textInput.SetValue("")
				m.inputLabel = "Enter new deck name:"
			}

		case key.Matches(msg, m.keys.RenameDeck):
			if m.viewState == viewDeckBrowser && m.currentDeck != m.store.RootDeck {
				m.viewState = viewRenameDeck
				m.textInput.Focus()
				m.textInput.SetValue(m.currentDeck.Name)
				m.inputLabel = "Enter new name for deck:"
			}

		case key.Matches(msg, m.keys.DeleteDeck):
			if m.viewState == viewDeckBrowser && m.currentDeck != m.store.RootDeck {
				m.viewState = viewDeleteDeck
				m.inputLabel = fmt.Sprintf("Delete deck '%s' and all its cards? (y/n)", m.currentDeck.Name)
			}

		case key.Matches(msg, m.keys.New):
			if m.viewState == viewDeckBrowser {
				m.viewState = viewCreateCard
				m.textInput.Focus()
				m.textInput.SetValue("")
				m.inputLabel = "Enter card title:"
			}

		case key.Matches(msg, m.keys.Search):
			if m.viewState == viewDeckBrowser {
				m.viewState = viewSearch
				m.textInput.Focus()
				m.textInput.SetValue("")
				m.inputLabel = "Enter search text:"
			}

		case msg.String() == "y" || msg.String() == "Y":
			if m.viewState == viewDeleteDeck {
				if err := m.store.DeleteDeck(m.currentDeck); err != nil {
					m.error = fmt.Sprintf("Error deleting deck: %v", err)
				} else {
					// Move to parent deck
					m.currentDeck = m.currentDeck.ParentDeck
				}
				m.viewState = viewDeckBrowser
				m.updateViewport()
			}

		case msg.String() == "n" || msg.String() == "N":
			if m.viewState == viewDeleteDeck {
				m.viewState = viewDeckBrowser
				m.updateViewport()
			}

		case msg.String() == "enter":
			switch m.viewState {
			case viewDeckList:
				if m.selectedDeck != nil {
					m.currentDeck = m.selectedDeck
					m.viewState = viewDeckBrowser
					m.updateViewport()
				}
			case viewCreateDeck:
				// Create a new deck
				deckName := m.textInput.Value()
				if deckName != "" {
					newDeck, err := m.store.CreateDeck(deckName, m.currentDeck)
					if err != nil {
						m.error = fmt.Sprintf("Error creating deck: %v", err)
					} else {
						m.currentDeck = newDeck
					}
				}
				m.textInput.Blur()
				m.viewState = viewDeckBrowser
				m.updateViewport()
			case viewRenameDeck:
				// Rename the current deck
				newName := m.textInput.Value()
				if newName != "" && newName != m.currentDeck.Name {
					if err := m.store.RenameDeck(m.currentDeck, newName); err != nil {
						m.error = fmt.Sprintf("Error renaming deck: %v", err)
					}
				}
				m.textInput.Blur()
				m.viewState = viewDeckBrowser
				m.updateViewport()
			case viewCreateCard:
				// Create a new card (this is just the title step)
				cardTitle := m.textInput.Value()
				if cardTitle != "" {
					// For now, create a simple card with placeholder content
					_, err := m.store.CreateCardInDeck(
						cardTitle,
						"Enter your question here",
						"Enter your answer here",
						[]string{"new"},
						m.currentDeck,
					)
					if err != nil {
						m.error = fmt.Sprintf("Error creating card: %v", err)
					} else {
						m.error = "Card created with default content. Use an external editor to complete it."
					}
				}
				m.textInput.Blur()
				m.viewState = viewDeckBrowser
				m.updateViewport()
			case viewSearch:
				// Perform search
				searchText := m.textInput.Value()
				if searchText != "" {
					m.searchResults = m.store.SearchCards(searchText)
					m.viewState = viewSearchResults
					m.updateSearchResults()
				} else {
					m.viewState = viewDeckBrowser
					m.updateViewport()
				}
				m.textInput.Blur()
			}

		case msg.String() == "up":
			if m.viewState == viewDeckList {
				// Find the index of the currently selected deck
				allDecks := m.store.RootDeck.AllDecks()
				currentIdx := -1
				for i, d := range allDecks {
					if d == m.selectedDeck {
						currentIdx = i
						break
					}
				}

				// Move selection up if possible
				if currentIdx > 0 {
					m.selectedDeck = allDecks[currentIdx-1]
					// Adjust offset if necessary
					if currentIdx-1 < m.deckListOffset {
						m.deckListOffset = currentIdx - 1
					}
					m.updateDeckList()
				}
			}

		case msg.String() == "down":
			if m.viewState == viewDeckList {
				// Find the index of the currently selected deck
				allDecks := m.store.RootDeck.AllDecks()
				currentIdx := -1
				for i, d := range allDecks {
					if d == m.selectedDeck {
						currentIdx = i
						break
					}
				}

				// Move selection down if possible
				if currentIdx < len(allDecks)-1 {
					m.selectedDeck = allDecks[currentIdx+1]
					// Adjust offset if necessary
					if currentIdx+1 >= m.deckListOffset+m.height-10 {
						m.deckListOffset = currentIdx - m.height + 11
					}
					m.updateDeckList()
				}
			}

		case key.Matches(msg, m.keys.Edit):
			// Placeholder for edit functionality
			m.error = "Editing cards is not implemented in this version"

		case key.Matches(msg, m.keys.Delete):
			// Placeholder for delete functionality
			m.error = "Deleting cards is not implemented in this version"

		case key.Matches(msg, m.keys.Tags):
			// Placeholder for tag editing functionality
			m.error = "Editing tags is not implemented in this version"

		case key.Matches(msg, m.keys.MoveToDeck):
			// Placeholder for move to deck functionality
			m.error = "Moving cards is not implemented in this version"
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

	// Handle text input updates
	if m.viewState == viewCreateDeck || m.viewState == viewRenameDeck ||
		m.viewState == viewCreateCard || m.viewState == viewSearch {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
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

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Width(m.width).
		Padding(0, 1)

	// Render header based on current view
	var headerText string
	switch m.viewState {
	case viewReview:
		if m.reviewState == stateCompleted {
			headerText = "GoCard - Review Session Complete"
		} else {
			progress := fmt.Sprintf("%d/%d", m.reviewCount+1, m.reviewCount+len(m.dueCards))
			headerText = fmt.Sprintf("GoCard - Review Session - %s - Deck: %s", progress, m.currentDeck.FullName())
		}
	case viewDeckBrowser:
		headerText = fmt.Sprintf("GoCard - Deck Browser - %s", m.currentDeck.FullName())
	case viewDeckList:
		headerText = "GoCard - Select Deck"
	case viewDeckStats:
		headerText = fmt.Sprintf("GoCard - Deck Statistics - %s", m.currentDeck.FullName())
	case viewCreateDeck:
		headerText = "GoCard - Create New Deck"
	case viewRenameDeck:
		headerText = "GoCard - Rename Deck"
	case viewDeleteDeck:
		headerText = "GoCard - Delete Deck"
	case viewCreateCard:
		headerText = "GoCard - Create New Card"
	case viewSearch:
		headerText = "GoCard - Search Cards"
	case viewSearchResults:
		headerText = "GoCard - Search Results"
	default:
		headerText = "GoCard"
	}
	sb.WriteString(headerStyle.Render(headerText))
	sb.WriteString("\n")

	// Render error if present
	if m.error != "" {
		sb.WriteString(errorStyle.Render(m.error))
		sb.WriteString("\n")
	}

	// Render text input for input views
	if m.viewState == viewCreateDeck || m.viewState == viewRenameDeck ||
		m.viewState == viewCreateCard || m.viewState == viewSearch {
		sb.WriteString(inputStyle.Render(m.inputLabel))
		sb.WriteString("\n")
		sb.WriteString(m.textInput.View())
		sb.WriteString("\n")
	} else if m.viewState == viewDeleteDeck {
		sb.WriteString(inputStyle.Render(m.inputLabel))
		sb.WriteString("\n")
	} else {
		// Render main content area
		sb.WriteString(m.viewport.View())
		sb.WriteString("\n")
	}

	// Render help or footer
	helpView := m.help.View(m.keys)
	if m.showHelp {
		sb.WriteString(helpView)
	} else {
		// Show a simple footer with basic instructions
		var footerText string
		switch m.viewState {
		case viewReview:
			if m.reviewState == stateShowingQuestion {
				footerText = "Press space to show answer, ? for help"
			} else if m.reviewState == stateShowingAnswer {
				footerText = "Rate 0-5, ? for help"
			} else {
				footerText = "Press esc to return to deck browser, q to quit, ? for help"
			}
		case viewDeckBrowser:
			footerText = "Press space to review, c to change deck, n for new card, ? for help"
		case viewDeckList:
			footerText = "Use arrow keys to select, Enter to confirm, Esc to cancel"
		case viewCreateDeck, viewRenameDeck, viewCreateCard, viewSearch:
			footerText = "Enter text and press Enter to confirm, Esc to cancel"
		case viewDeleteDeck:
			footerText = "Press y to confirm, n to cancel"
		default:
			footerText = "Press ? for help, q to quit"
		}
		sb.WriteString(footerStyle.Render(footerText))
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
