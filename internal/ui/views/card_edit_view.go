// File: internal/ui/views/card_edit_view.go
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/render"
)

// EditorField represents which field is currently being edited
type EditorField int

const (
	FieldTitle EditorField = iota
	FieldTags
	FieldQuestion
	FieldAnswer
	FieldNone
)

// EditorMode represents the current mode of the editor
type EditorMode int

const (
	ModeEdit EditorMode = iota
	ModePreview
)

// CardEditView handles the card editing interface
type CardEditView struct {
	BaseView
	store        *storage.CardStore
	renderer     *render.Renderer
	card         *card.Card
	originalDeck string
	isNewCard    bool

	mode        EditorMode
	activeField EditorField

	titleInput    textinput.Model
	tagsInput     textinput.Model
	questionInput textarea.Model
	answerInput   textarea.Model

	previewContent string
	showHelp       bool
	errorMsg       string
}

// NewCardEditView creates a new card editing view
func NewCardEditView(store *storage.CardStore, cardToEdit *card.Card, isNew bool, deckPath string, width, height int) (*CardEditView, error) {
	baseView := NewBaseView(ViewEditCard, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	view := &CardEditView{
		BaseView:     baseView,
		store:        store,
		renderer:     renderer,
		card:         cardToEdit,
		originalDeck: deckPath,
		isNewCard:    isNew,
		mode:         ModeEdit,
		activeField:  FieldTitle,
	}

	// Initialize text inputs
	view.initInputs()

	// If it's a new card, provide a template
	if isNew {
		view.questionInput.SetValue("Enter your question here...")
		view.answerInput.SetValue("Enter your answer here...")
	}

	return view, nil
}

// Init implements View.Init
func (v *CardEditView) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements View.Update
func (v *CardEditView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first
		if v.mode == ModeEdit && v.activeField != FieldNone {
			switch v.activeField {
			case FieldTitle:
				if msg.String() == "tab" {
					v.nextField()
					return v, nil
				} else if msg.String() == "shift+tab" {
					v.previousField()
					return v, nil
				}
			case FieldTags:
				if msg.String() == "tab" {
					v.nextField()
					return v, nil
				} else if msg.String() == "shift+tab" {
					v.previousField()
					return v, nil
				}
			// For textareas, use alt+arrows instead of ctrl+tab
			case FieldQuestion, FieldAnswer:
				// Let textarea handle tab internally for indentation
				// Use alt+up/down for field navigation
				if msg.String() == "alt+down" {
					v.nextField()
					return v, nil
				} else if msg.String() == "alt+up" {
					v.previousField()
					return v, nil
				}
			}
		}

		// Global key handlers
		switch {
		case input.KeyMatches(msg, keys.Quit):
			return v, tea.Quit

		case input.KeyMatches(msg, keys.Back):
			// Return to previous view without saving
			if v.isNewCard {
				// Return to deck browser for new cards
				deckView, err := NewDeckBrowserView(v.store, v.originalDeck, v.width, v.height)
				if err != nil {
					v.errorMsg = fmt.Sprintf("Error returning to deck browser: %v", err)
					return v, nil
				}
				return deckView, deckView.Init()
			} else {
				// Return to review view for existing cards
				reviewView, err := NewReviewView(v.store, v.originalDeck, v.width, v.height)
				if err != nil {
					v.errorMsg = fmt.Sprintf("Error returning to review: %v", err)
					return v, nil
				}
				return reviewView, reviewView.Init()
			}

		case msg.String() == "ctrl+p":
			// Toggle preview mode
			if v.mode == ModeEdit {
				v.mode = ModePreview
				v.updatePreview()
			} else {
				v.mode = ModeEdit
			}
			return v, nil

		case msg.String() == "ctrl+s":
			// Save card
			err := v.saveCard()
			if err != nil {
				v.errorMsg = fmt.Sprintf("Error saving card: %v", err)
				return v, nil
			}

			// Return to appropriate view
			if v.isNewCard {
				deckView, err := NewDeckBrowserView(v.store, v.originalDeck, v.width, v.height)
				if err != nil {
					v.errorMsg = fmt.Sprintf("Error returning to deck browser: %v", err)
					return v, nil
				}
				return deckView, deckView.Init()
			} else {
				reviewView, err := NewReviewView(v.store, v.originalDeck, v.width, v.height)
				if err != nil {
					v.errorMsg = fmt.Sprintf("Error returning to review: %v", err)
					return v, nil
				}
				return reviewView, reviewView.Init()
			}
		}

	case tea.WindowSizeMsg:
		v.SetDimensions(msg.Width, msg.Height)
		if err := v.renderer.UpdateWidth(msg.Width); err != nil {
			v.errorMsg = fmt.Sprintf("Error updating renderer: %v", err)
		}

		// Update textarea widths
		v.questionInput.SetWidth(msg.Width - 4)
		v.answerInput.SetWidth(msg.Width - 4)
	}

	// Handle input for active field
	switch v.activeField {
	case FieldTitle:
		newTitleModel, cmd := v.titleInput.Update(msg)
		v.titleInput = newTitleModel
		cmds = append(cmds, cmd)

	case FieldTags:
		newTagsModel, cmd := v.tagsInput.Update(msg)
		v.tagsInput = newTagsModel
		cmds = append(cmds, cmd)

	case FieldQuestion:
		newQuestionModel, cmd := v.questionInput.Update(msg)
		v.questionInput = newQuestionModel
		cmds = append(cmds, cmd)

	case FieldAnswer:
		newAnswerModel, cmd := v.answerInput.Update(msg)
		v.answerInput = newAnswerModel
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

// Render implements View.Render
func (v *CardEditView) Render(width, height int) string {
	var sb strings.Builder

	// Render header
	headerText := "GoCard - "
	if v.isNewCard {
		headerText += "Create New Card"
	} else {
		headerText += "Edit Card: " + v.card.Title
	}
	sb.WriteString(v.renderer.HeaderStyle(headerText))
	sb.WriteString("\n")

	// Render error if present
	if v.errorMsg != "" {
		sb.WriteString(v.renderer.ErrorStyle(v.errorMsg))
		sb.WriteString("\n")
	}

	// Main content
	if v.mode == ModeEdit {
		// Render edit interface
		sb.WriteString("\n")

		// Title field
		if v.activeField == FieldTitle {
			sb.WriteString(v.renderer.GetStyles().Highlight.Render("Title:") + "\n")
		} else {
			sb.WriteString("Title:\n")
		}
		sb.WriteString(v.titleInput.View() + "\n\n")

		// Tags field
		if v.activeField == FieldTags {
			sb.WriteString(v.renderer.GetStyles().Highlight.Render("Tags (comma separated):") + "\n")
		} else {
			sb.WriteString("Tags (comma separated):\n")
		}
		sb.WriteString(v.tagsInput.View() + "\n\n")

		// Question field
		if v.activeField == FieldQuestion {
			sb.WriteString(v.renderer.GetStyles().Highlight.Render("Question:") + "\n")
		} else {
			sb.WriteString("Question:\n")
		}
		sb.WriteString(v.questionInput.View() + "\n\n")

		// Answer field
		if v.activeField == FieldAnswer {
			sb.WriteString(v.renderer.GetStyles().Highlight.Render("Answer:") + "\n")
		} else {
			sb.WriteString("Answer:\n")
		}
		sb.WriteString(v.answerInput.View() + "\n")

	} else {
		// Render preview
		sb.WriteString(v.previewContent)
	}

	// Render footer with shortcuts
	var footerText string
	if v.activeField == FieldQuestion || v.activeField == FieldAnswer {
		footerText = "Alt+Down: Next Field • Alt+Up: Previous Field • Ctrl+p: Preview • Ctrl+s: Save • Esc: Cancel"
	} else {
		footerText = "Tab: Next Field • Shift+Tab: Previous Field • Ctrl+p: Preview • Ctrl+s: Save • Esc: Cancel"
	}
	sb.WriteString("\n" + v.renderer.FooterStyle(footerText))

	return sb.String()
}

// initInputs initializes all text inputs
func (v *CardEditView) initInputs() {
	// Initialize title input
	v.titleInput = textinput.New()
	v.titleInput.Placeholder = "Card Title"
	v.titleInput.Focus()
	v.titleInput.CharLimit = 100
	v.titleInput.Width = v.width - 4

	if v.card.Title != "" {
		v.titleInput.SetValue(v.card.Title)
	}

	// Initialize tags input
	v.tagsInput = textinput.New()
	v.tagsInput.Placeholder = "tag1, tag2, tag3"
	v.tagsInput.CharLimit = 200
	v.tagsInput.Width = v.width - 4

	if len(v.card.Tags) > 0 {
		v.tagsInput.SetValue(strings.Join(v.card.Tags, ", "))
	}

	// Initialize question input - now using textarea
	v.questionInput = textarea.New()
	v.questionInput.Placeholder = "Enter question..."
	v.questionInput.SetWidth(v.width - 4)
	v.questionInput.SetHeight(5) // Show 5 lines

	if v.card.Question != "" {
		v.questionInput.SetValue(v.card.Question)
	}

	// Initialize answer input - now using textarea
	v.answerInput = textarea.New()
	v.answerInput.Placeholder = "Enter answer..."
	v.answerInput.SetWidth(v.width - 4)
	v.answerInput.SetHeight(10) // Show 10 lines

	if v.card.Answer != "" {
		v.answerInput.SetValue(v.card.Answer)
	}
}

// updatePreview generates markdown preview from inputs
func (v *CardEditView) updatePreview() {
	var content strings.Builder

	content.WriteString("# " + v.titleInput.Value() + "\n\n")
	content.WriteString("## Question\n\n")
	content.WriteString(v.questionInput.Value() + "\n\n")
	content.WriteString("## Answer\n\n")
	content.WriteString(v.answerInput.Value())

	rendered, err := v.store.RenderMarkdown(content.String())
	if err != nil {
		v.errorMsg = fmt.Sprintf("Error rendering preview: %v", err)
		v.previewContent = content.String()
	} else {
		v.previewContent = rendered
	}
}

// saveCard saves the card to storage
func (v *CardEditView) saveCard() error {
	// Update card from inputs
	v.card.Title = v.titleInput.Value()
	v.card.Question = v.questionInput.Value()
	v.card.Answer = v.answerInput.Value()

	// Parse tags
	tagStr := v.tagsInput.Value()
	tags := strings.Split(tagStr, ",")
	cleanTags := []string{}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	v.card.Tags = cleanTags

	// Save card
	if v.isNewCard {
		// For new cards, get the current deck
		deckObj, err := v.store.GetDeckByRelativePath(v.originalDeck)
		if err != nil {
			return fmt.Errorf("failed to get deck: %w", err)
		}

		// Create new card in deck
		_, err = v.store.CreateCardInDeck(
			v.card.Title,
			v.card.Question,
			v.card.Answer,
			v.card.Tags,
			deckObj,
		)
		return err
	} else {
		// For existing cards, update
		return v.store.SaveCard(v.card)
	}
}

// nextField moves to the next input field
func (v *CardEditView) nextField() {
	switch v.activeField {
	case FieldTitle:
		v.activeField = FieldTags
		v.titleInput.Blur()
		v.tagsInput.Focus()
	case FieldTags:
		v.activeField = FieldQuestion
		v.tagsInput.Blur()
		v.questionInput.Focus()
	case FieldQuestion:
		v.activeField = FieldAnswer
		v.questionInput.Blur()
		v.answerInput.Focus()
	case FieldAnswer:
		v.activeField = FieldTitle
		v.answerInput.Blur()
		v.titleInput.Focus()
	}
}

// previousField moves to the previous input field
func (v *CardEditView) previousField() {
	switch v.activeField {
	case FieldTitle:
		v.activeField = FieldAnswer
		v.titleInput.Blur()
		v.answerInput.Focus()
	case FieldTags:
		v.activeField = FieldTitle
		v.tagsInput.Blur()
		v.titleInput.Focus()
	case FieldQuestion:
		v.activeField = FieldTags
		v.questionInput.Blur()
		v.tagsInput.Focus()
	case FieldAnswer:
		v.activeField = FieldQuestion
		v.answerInput.Blur()
		v.questionInput.Focus()
	}
}
