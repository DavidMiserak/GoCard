// Package views contains the different UI views for GoCard.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/render"
)

// DeckListView handles the deck selection interface
type DeckListView struct {
	BaseView
	store        *storage.CardStore
	renderer     *render.Renderer
	currentDeck  *deck.Deck             // Current directory/deck being viewed
	visibleDecks []*deck.Deck           // List of visible decks in current view
	cursor       int                    // Current cursor position in the list
	breadcrumbs  []*deck.Deck           // Breadcrumb trail for navigation
	stats        map[string]interface{} // Store statistics for the current deck
}

// NewDeckListView creates a new deck list view
func NewDeckListView(store *storage.CardStore, currentDeckPath string, width, height int) (*DeckListView, error) {
	baseView := NewBaseView(ViewDeckList, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	// Get the current deck from the path or use root deck
	var currentDeck *deck.Deck
	if currentDeckPath == "" {
		currentDeck = store.RootDeck
	} else {
		currentDeck, err = store.GetDeckByRelativePath(currentDeckPath)
		if err != nil {
			return nil, err
		}
	}

	// Initialize the deck list view
	view := &DeckListView{
		BaseView:    baseView,
		store:       store,
		renderer:    renderer,
		currentDeck: currentDeck,
		cursor:      0,
		breadcrumbs: []*deck.Deck{},
	}

	// Initialize the breadcrumbs
	view.initBreadcrumbs(currentDeck)

	// Update the visible decks and viewport content
	view.updateVisibleDecks()
	view.updateViewport()

	return view, nil
}

// initBreadcrumbs creates breadcrumb navigation path from root to current deck
func (v *DeckListView) initBreadcrumbs(deckObj *deck.Deck) {
	// Start with an empty breadcrumb trail
	v.breadcrumbs = []*deck.Deck{}

	// If we're at the root, no breadcrumbs needed
	if deckObj == v.store.RootDeck {
		return
	}

	// Build a list of decks from current to root
	var path []*deck.Deck
	current := deckObj
	for current != nil && current != v.store.RootDeck {
		path = append([]*deck.Deck{current}, path...)
		current = current.ParentDeck
	}

	// Set the breadcrumbs
	v.breadcrumbs = path
}

// updateVisibleDecks refreshes the list of decks shown in the current view
func (v *DeckListView) updateVisibleDecks() {
	// Start with the parent deck if it exists (for "Back" option)
	v.visibleDecks = []*deck.Deck{}

	// Add the parent deck as a navigation option if we're not at root
	if v.currentDeck != v.store.RootDeck && v.currentDeck.ParentDeck != nil {
		v.visibleDecks = append(v.visibleDecks, v.currentDeck.ParentDeck)
	}

	// Add all subdecks of the current deck
	for _, subDeck := range v.currentDeck.SubDecks {
		v.visibleDecks = append(v.visibleDecks, subDeck)
	}

	// Reset cursor position to prevent out-of-bounds errors
	if len(v.visibleDecks) > 0 {
		if v.cursor >= len(v.visibleDecks) {
			v.cursor = len(v.visibleDecks) - 1
		}
	} else {
		v.cursor = 0
	}

	// Fetch statistics for the current deck
	v.stats = v.store.GetDeckStats(v.currentDeck)
}

// Init implements View.Init
func (v *DeckListView) Init() tea.Cmd {
	return nil
}

// Update implements View.Update
func (v *DeckListView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case input.KeyMatches(msg, keys.Quit):
			return v, tea.Quit

		case input.KeyMatches(msg, keys.Back):
			// Go back to the deck browser view
			deckView, err := NewDeckBrowserView(v.store, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error returning to deck browser: %v", err))
				return v, nil
			}
			return deckView, deckView.Init()

		case input.KeyMatches(msg, keys.ShowAnswer): // Use space or enter to select
			fallthrough
		case msg.String() == "enter":
			fallthrough
		case msg.String() == "l": // vim right
			fallthrough
		case msg.String() == "ctrl+f": // emacs forward
			// Handle selection based on cursor position
			if len(v.visibleDecks) > 0 {
				selectedDeck := v.visibleDecks[v.cursor]

				// If first item is parent (Back option)
				if v.currentDeck != v.store.RootDeck && v.currentDeck.ParentDeck != nil && v.cursor == 0 {
					// Navigate up to parent deck
					v.currentDeck = selectedDeck
					v.initBreadcrumbs(v.currentDeck)
					v.updateVisibleDecks()
					v.updateViewport()
					return v, nil
				} else {
					// Navigate into selected subdeck
					v.currentDeck = selectedDeck
					v.initBreadcrumbs(v.currentDeck)
					v.updateVisibleDecks()
					v.updateViewport()
					return v, nil
				}
			}

		case msg.String() == "up":
			fallthrough
		case msg.String() == "k": // vim up
			fallthrough
		case msg.String() == "ctrl+p": // emacs previous
			// Move cursor up
			if v.cursor > 0 {
				v.cursor--
				v.updateViewport()
			}

		case msg.String() == "down":
			fallthrough
		case msg.String() == "j": // vim down
			fallthrough
		case msg.String() == "ctrl+n": // emacs next
			// Move cursor down
			if v.cursor < len(v.visibleDecks)-1 {
				v.cursor++
				v.updateViewport()
			}

		case msg.String() == "h": // vim left
			fallthrough
		case msg.String() == "ctrl+b": // emacs back
			// Go back - same as Esc
			deckView, err := NewDeckBrowserView(v.store, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error returning to deck browser: %v", err))
				return v, nil
			}
			return deckView, deckView.Init()

		case input.KeyMatches(msg, keys.ChangeDeck):
			// Same as Back - return to deck browser for current deck
			deckView, err := NewDeckBrowserView(v.store, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error returning to deck browser: %v", err))
				return v, nil
			}
			return deckView, deckView.Init()
		}

	case tea.WindowSizeMsg:
		v.SetDimensions(msg.Width, msg.Height)
		if err := v.renderer.UpdateWidth(msg.Width); err != nil {
			v.SetError(fmt.Sprintf("Error updating renderer: %v", err))
		}
		v.updateViewport()
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// Render implements View.Render
func (v *DeckListView) Render(width, height int) string {
	var sb strings.Builder

	// Render header
	headerText := fmt.Sprintf("GoCard - Select Deck - %s", v.currentDeck.FullName())
	sb.WriteString(v.renderer.HeaderStyle(headerText))
	sb.WriteString("\n")

	// Render error if present
	if v.GetError() != "" {
		sb.WriteString(v.renderer.ErrorStyle(v.GetError()))
		sb.WriteString("\n")
	}

	// Render main content
	sb.WriteString(v.viewport.View())
	sb.WriteString("\n")

	// Render footer - now with vim/emacs keys
	footerText := "↑/j/k: Navigate • Enter: Select • Esc: Back • ctrl+o: View Deck • ctrl+h: Help"
	sb.WriteString(v.renderer.FooterStyle(footerText))

	return sb.String()
}

// updateViewport updates the viewport content based on the deck list
func (v *DeckListView) updateViewport() {
	styles := v.renderer.GetStyles()

	var content strings.Builder

	// Breadcrumb navigation
	breadcrumbPath := "Root"
	if len(v.breadcrumbs) > 0 {
		paths := []string{"Root"}
		for _, d := range v.breadcrumbs {
			paths = append(paths, d.Name)
		}
		breadcrumbPath = strings.Join(paths, " > ")
	}

	content.WriteString(styles.Subtle.Render(breadcrumbPath))
	content.WriteString("\n\n")

	// Current deck info - more compact formatting
	content.WriteString(styles.Title.Render("# "+v.currentDeck.Name) + "\n")

	if totalCards, ok := v.stats["total_cards"].(int); ok && totalCards > 0 {
		content.WriteString(fmt.Sprintf("• Cards: %d total, %d due\n",
			totalCards, v.stats["due_cards"]))
	} else {
		content.WriteString("• No cards in this deck\n")
	}

	if len(v.currentDeck.SubDecks) > 0 {
		content.WriteString(fmt.Sprintf("• Subdecks: %d\n", len(v.currentDeck.SubDecks)))
	}

	content.WriteString("\n")

	// Available decks list with fixed column width layout
	content.WriteString(styles.Highlight.Render("Select a deck:") + "\n\n")

	if len(v.visibleDecks) == 0 {
		content.WriteString(styles.DimmedText.Render("No decks found. Press 'C' to create a new deck."))
		content.WriteString("\n")
	} else {
		// Use a table-like layout with fixed columns
		format := "%-2s %-14s %-18s\n"

		// Render parent deck as "Back" option if not at root
		startIdx := 0
		if v.currentDeck != v.store.RootDeck && v.currentDeck.ParentDeck != nil && len(v.visibleDecks) > 0 {
			isSelected := v.cursor == 0
			cursor := " "
			if isSelected {
				cursor = ">"
			}

			if isSelected {
				content.WriteString(fmt.Sprintf(format, cursor,
					styles.Highlight.Render("← Back"), ""))
			} else {
				content.WriteString(fmt.Sprintf(format, cursor, "← Back", ""))
			}
			startIdx = 1
		}

		// Render subdeck list with fixed column positioning
		for i := startIdx; i < len(v.visibleDecks); i++ {
			d := v.visibleDecks[i]
			deckStats := v.store.GetDeckStats(d)

			isSelected := v.cursor == i
			cursor := " "
			if isSelected {
				cursor = ">"
			}

			stats := fmt.Sprintf("(%d cards, %d due)",
				deckStats["total_cards"],
				deckStats["due_cards"])

			if isSelected {
				content.WriteString(fmt.Sprintf(format, cursor,
					styles.Highlight.Render(d.Name),
					styles.Highlight.Render(stats)))
			} else {
				content.WriteString(fmt.Sprintf(format, cursor, d.Name, stats))
			}
		}
	}

	// Add some empty lines for visual spacing (just one line to avoid excess space)
	content.WriteString("\n")

	// Single line help text with no wrapping
	content.WriteString(styles.Subtle.Render("Navigate with arrow keys, press Enter to select a deck, Esc to go back"))

	v.viewport.SetContent(content.String())
}
