package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/render"
)

// ReviewView handles the card review state and UI
type ReviewView struct {
	BaseView
	store       *storage.CardStore
	renderer    *render.Renderer
	currentCard *card.Card
	dueCards    []*card.Card
	reviewState ReviewState
	reviewCount int
}

// NewReviewView creates a new review view
func NewReviewView(store *storage.CardStore, currentDeck string, width, height int) (*ReviewView, error) {
	baseView := NewBaseView(ViewReview, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	// Get deck from store
	deck, err := store.GetDeckByRelativePath(currentDeck)
	if err != nil {
		return nil, err
	}

	// Get due cards for the deck
	dueCards := store.GetDueCardsInDeck(deck)

	view := &ReviewView{
		BaseView:    baseView,
		store:       store,
		renderer:    renderer,
		dueCards:    dueCards,
		reviewState: StateShowingQuestion,
		reviewCount: 0,
	}

	// Initialize with the first card if available
	if len(dueCards) > 0 {
		view.currentCard = dueCards[0]
		view.updateViewport()
	} else {
		view.renderSessionSummary()
		view.reviewState = StateCompleted
	}

	return view, nil
}

// Init implements View.Init
func (v *ReviewView) Init() tea.Cmd {
	return nil
}

// Update implements View.Update
func (v *ReviewView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case input.KeyMatches(msg, keys.ShowAnswer):
			if v.reviewState == StateShowingQuestion && v.currentCard != nil {
				v.reviewState = StateShowingAnswer
				v.updateViewport()
			}

		case input.KeyMatches(msg, keys.Rate0),
			input.KeyMatches(msg, keys.Rate1),
			input.KeyMatches(msg, keys.Rate2),
			input.KeyMatches(msg, keys.Rate3),
			input.KeyMatches(msg, keys.Rate4),
			input.KeyMatches(msg, keys.Rate5):

			// Only process ratings when showing an answer
			if v.reviewState == StateShowingAnswer && v.currentCard != nil {
				// Extract the rating from the key pressed (0-5)
				rating, ok := input.GetRatingFromKey(msg, keys)
				if !ok {
					// This shouldn't happen due to the case statement above, but handle it anyway
					rating = 3 // Default to "good" rating
				}

				// Apply the SM-2 algorithm to the card
				if err := v.store.ReviewCard(v.currentCard, rating); err != nil {
					v.SetError(fmt.Sprintf("Error reviewing card: %v", err))
				} else {
					v.reviewCount++
					v.moveToNextCard()
				}
			}
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
func (v *ReviewView) Render(width, height int) string {
	var sb strings.Builder

	// Render header
	var headerText string
	if v.reviewState == StateCompleted {
		headerText = "GoCard - Review Session Complete"
	} else {
		progress := fmt.Sprintf("%d/%d", v.reviewCount+1, v.reviewCount+len(v.dueCards))
		headerText = fmt.Sprintf("GoCard - Review Session - %s", progress)
	}
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
	var footerText string
	if v.reviewState == StateShowingQuestion {
		footerText = "Press space to show answer, ? for help"
	} else if v.reviewState == StateShowingAnswer {
		footerText = "Rate 0-5, ? for help"
	} else {
		footerText = "Press esc to return to deck browser, q to quit, ? for help"
	}
	sb.WriteString(v.renderer.FooterStyle(footerText))

	return sb.String()
}

// updateViewport updates the viewport content based on the current state
func (v *ReviewView) updateViewport() {
	if v.currentCard == nil {
		if len(v.dueCards) == 0 {
			v.viewport.SetContent("No cards due for review. Great job!")
		} else {
			v.viewport.SetContent("Error: Current card is nil but due cards exist.")
		}
		return
	}

	var content strings.Builder

	// Render card metadata
	tags := strings.Join(v.currentCard.Tags, ", ")
	content.WriteString(fmt.Sprintf("Card: %s\n", v.currentCard.Title))
	content.WriteString(fmt.Sprintf("Tags: %s\n\n", tags))

	// Render question
	content.WriteString("## Question\n\n")
	questionMd, err := v.renderer.RenderMarkdown(v.currentCard.Question)
	if err != nil {
		v.SetError(fmt.Sprintf("Error rendering question: %v", err))
		questionMd = v.currentCard.Question
	}
	content.WriteString(questionMd)

	// Render answer if we're in answer state
	if v.reviewState == StateShowingAnswer {
		content.WriteString("\n\n## Answer\n\n")
		answerMd, err := v.renderer.RenderMarkdown(v.currentCard.Answer)
		if err != nil {
			v.SetError(fmt.Sprintf("Error rendering answer: %v", err))
			answerMd = v.currentCard.Answer
		}
		content.WriteString(answerMd)

		// Add rating prompt
		content.WriteString("\n\nHow well did you remember? (0-5)\n")
		content.WriteString("0: Forgot completely | 3: Correct with effort | 5: Perfect recall\n")
	}

	v.viewport.SetContent(content.String())
	v.viewport.GotoTop()
}

// moveToNextCard moves to the next card in the due cards list
func (v *ReviewView) moveToNextCard() {
	// Remove the current card from the due cards list
	for i, c := range v.dueCards {
		if c == v.currentCard {
			v.dueCards = append(v.dueCards[:i], v.dueCards[i+1:]...)
			break
		}
	}

	// Move to the next card or show completion
	if len(v.dueCards) > 0 {
		v.currentCard = v.dueCards[0]
		v.reviewState = StateShowingQuestion
		v.updateViewport()
	} else {
		v.reviewState = StateCompleted
		v.renderSessionSummary()
	}
}

// renderSessionSummary generates a summary of the review session
func (v *ReviewView) renderSessionSummary() {
	var content strings.Builder

	content.WriteString("# Review Session Completed\n\n")
	content.WriteString(fmt.Sprintf("Cards reviewed: %d\n", v.reviewCount))

	// Get statistics from the store
	stats := v.store.GetReviewStats()
	content.WriteString("\nCard Statistics:\n")
	content.WriteString(fmt.Sprintf("- Total cards: %d\n", stats["total_cards"]))
	content.WriteString(fmt.Sprintf("- Due cards remaining: %d\n", stats["due_cards"]))
	content.WriteString(fmt.Sprintf("- New cards: %d\n", stats["new_cards"]))
	content.WriteString(fmt.Sprintf("- Young cards (1-7 days): %d\n", stats["young_cards"]))
	content.WriteString(fmt.Sprintf("- Mature cards (>7 days): %d\n", stats["mature_cards"]))

	// Show next due date
	nextDue := v.store.GetNextDueDate()
	content.WriteString(fmt.Sprintf("\nNext review session: %s\n", nextDue.Format(time.DateOnly)))

	// Render to the viewport
	summaryMd, err := v.renderer.RenderMarkdown(content.String())
	if err != nil {
		v.SetError(fmt.Sprintf("Error rendering summary: %v", err))
		v.viewport.SetContent(content.String())
	} else {
		v.viewport.SetContent(summaryMd)
	}
	v.viewport.GotoTop()
}
