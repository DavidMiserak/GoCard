package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/deck"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/render"
)

// DeckBrowserView handles the deck browsing state and UI
type DeckBrowserView struct {
	BaseView
	store       *storage.CardStore
	renderer    *render.Renderer
	currentDeck *deck.Deck
}

// NewDeckBrowserView creates a new deck browser view
func NewDeckBrowserView(store *storage.CardStore, deckPath string, width, height int) (*DeckBrowserView, error) {
	baseView := NewBaseView(ViewDeckBrowser, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	// Get deck from store
	var deckObj *deck.Deck
	if deckPath == "" {
		deckObj = store.RootDeck
	} else {
		deckObj, err = store.GetDeckByRelativePath(deckPath)
		if err != nil {
			return nil, err
		}
	}

	view := &DeckBrowserView{
		BaseView:    baseView,
		store:       store,
		renderer:    renderer,
		currentDeck: deckObj,
	}

	view.updateDeckBrowser()

	return view, nil
}

// Init implements View.Init
func (v *DeckBrowserView) Init() tea.Cmd {
	return nil
}

// Update implements View.Update
func (v *DeckBrowserView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case input.KeyMatches(msg, keys.ShowAnswer):
			// Start a review session
			reviewView, err := NewReviewView(v.store, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error starting review: %v", err))
				return v, nil
			}
			return reviewView, reviewView.Init()

		case input.KeyMatches(msg, keys.ChangeDeck):
			// Launch the deck list view for deck navigation
			deckListView, err := NewDeckListView(v.store, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error opening deck list: %v", err))
				return v, nil
			}
			return deckListView, deckListView.Init()

		case input.KeyMatches(msg, keys.New):
			// Launch card creation view
			newCard := card.NewCard("", "", "", []string{})
			editView, err := NewCardEditView(v.store, newCard, true, v.currentDeck.PathFromRoot(), v.width, v.height)
			if err != nil {
				v.SetError(fmt.Sprintf("Error creating card: %v", err))
				return v, nil
			}
			return editView, editView.Init()

		case input.KeyMatches(msg, keys.Search):
			// This would launch the search view
			v.SetError("Search view not yet implemented in refactored version")
		}

	case tea.WindowSizeMsg:
		v.SetDimensions(msg.Width, msg.Height)
		if err := v.renderer.UpdateWidth(msg.Width); err != nil {
			v.SetError(fmt.Sprintf("Error updating renderer: %v", err))
		}
		v.updateDeckBrowser()
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// Render implements View.Render
func (v *DeckBrowserView) Render(width, height int) string {
	var sb strings.Builder

	// Render header
	headerText := fmt.Sprintf("GoCard - Deck Browser - %s", v.currentDeck.FullName())
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

	// Render footer
	footerText := "Press space to review, ctrl+o to change deck, ctrl+n for new card, ctrl+h for help"
	sb.WriteString(v.renderer.FooterStyle(footerText))

	return sb.String()
}

// updateDeckBrowser updates the viewport for the deck browser view
func (v *DeckBrowserView) updateDeckBrowser() {
	var content strings.Builder

	// Render deck information
	content.WriteString(fmt.Sprintf("# Deck: %s\n\n", v.currentDeck.FullName()))

	// Render subdeck list
	if len(v.currentDeck.SubDecks) > 0 {
		content.WriteString("## Subdecks\n\n")
		for _, subDeck := range v.currentDeck.SubDecks {
			stats := v.store.GetDeckStats(subDeck)
			content.WriteString(fmt.Sprintf("- %s (%d cards, %d due)\n",
				subDeck.Name,
				stats["total_cards"],
				stats["due_cards"]))
		}
		content.WriteString("\n")
	}

	// Get deck statistics
	stats := v.store.GetDeckStats(v.currentDeck)
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
	if len(v.currentDeck.Cards) > 0 {
		content.WriteString("\n## Cards in this deck\n\n")
		for _, c := range v.currentDeck.Cards {
			content.WriteString(fmt.Sprintf("- %s\n", c.Title))
		}
	}

	contentMd, err := v.renderer.RenderMarkdown(content.String())
	if err != nil {
		v.SetError(fmt.Sprintf("Error rendering markdown: %v", err))
		v.viewport.SetContent(content.String())
	} else {
		v.viewport.SetContent(contentMd)
	}
	v.viewport.GotoTop()
}
