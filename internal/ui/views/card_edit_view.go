// File: internal/ui/views/card_edit_view.go

package views

import (
	"fmt"
	"strings"
	"time"

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

// AutoSaveMsg is sent when it's time to auto-save
type AutoSaveMsg struct{}

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

	previewContent   string
	errorMsg         string
	statusMsg        string
	statusMsgTimer   int
	lastModified     time.Time
	autoSaveInterval time.Duration
	unsavedChanges   bool
	warnedAboutExit  bool // New flag to track if we've warned about exiting
}

// NewCardEditView creates a new card editing view with enhanced functionality
func NewCardEditView(store *storage.CardStore, cardToEdit *card.Card, isNew bool, deckPath string, width, height int) (*CardEditView, error) {
	baseView := NewBaseView(ViewEditCard, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	// Set auto-save interval
	autoSaveInterval := 60 * time.Second

	view := &CardEditView{
		BaseView:         baseView,
		store:            store,
		renderer:         renderer,
		card:             cardToEdit,
		originalDeck:     deckPath,
		isNewCard:        isNew,
		mode:             ModeEdit,
		activeField:      FieldTitle,
		autoSaveInterval: autoSaveInterval,
		unsavedChanges:   false,
		warnedAboutExit:  false,
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
	return tea.Batch(
		textinput.Blink,
		v.autoSaveCmd(),
	)
}

// autoSaveCmd returns a command that will trigger auto-save after the configured interval
func (v *CardEditView) autoSaveCmd() tea.Cmd {
	return tea.Tick(v.autoSaveInterval, func(time.Time) tea.Msg {
		return AutoSaveMsg{}
	})
}

// Update implements View.Update
func (v *CardEditView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AutoSaveMsg:
		// Handle auto-save timer tick
		if v.unsavedChanges {
			// Auto-save the card
			err := v.saveCard(false) // Save without exiting
			if err != nil {
				v.errorMsg = fmt.Sprintf("Auto-save failed: %v", err)
			} else {
				v.statusMsg = "Auto-saved"
				v.statusMsgTimer = 5 // Show for 5 updates
				v.unsavedChanges = false
			}
		}
		// Reset the timer for next auto-save
		cmds = append(cmds, v.autoSaveCmd())

	case tea.KeyMsg:
		// Only track content changes for non-navigation keys
		// This is the key fix to prevent navigation keys from triggering unsaved changes
		if !input.KeyMatches(msg, keys.Back) &&
			!input.KeyMatches(msg, keys.Quit) &&
			!input.IsNavKey(msg) &&
			msg.String() != "ctrl+p" &&
			msg.String() != "ctrl+s" &&
			msg.String() != "ctrl+shift+s" &&
			msg.String() != "ctrl+q" {
			v.unsavedChanges = true
			v.lastModified = time.Now()
		}

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
			// Check for unsaved changes
			if v.unsavedChanges && !v.warnedAboutExit {
				v.errorMsg = "You have unsaved changes. Press Ctrl+s to save or press Quit again to discard."
				v.warnedAboutExit = true
				return v, nil
			}
			return v, tea.Quit

		case input.KeyMatches(msg, keys.Back):
			// Check for unsaved changes and show warning first time
			if v.unsavedChanges && !v.warnedAboutExit {
				v.errorMsg = "You have unsaved changes. Press Esc again to discard changes."
				v.warnedAboutExit = true
				return v, nil
			}

			// Exit to the appropriate view (either after warning or no unsaved changes)
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
			// Save card and go to deck browser
			err := v.saveCard(false)
			if err != nil {
				v.errorMsg = fmt.Sprintf("Error saving card: %v", err)
				return v, nil
			}

			// Navigate to deck browser after saving
			deckView, err := NewDeckBrowserView(v.store, v.originalDeck, v.width, v.height)
			if err != nil {
				v.errorMsg = fmt.Sprintf("Error returning to deck browser: %v", err)
				return v, nil
			}
			return deckView, deckView.Init()

		case msg.String() == "ctrl+shift+s":
			// Save card and exit
			err := v.saveCard(true)
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

		case msg.String() == "ctrl+q":
			// Force exit without saving, regardless of unsaved changes
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

	// Decrement status message timer if active
	if v.statusMsgTimer > 0 {
		v.statusMsgTimer--
		if v.statusMsgTimer == 0 {
			v.statusMsg = ""
		}
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

	// Add mode to header
	switch v.mode {
	case ModeEdit:
		headerText += " (Edit Mode)"
	case ModePreview:
		headerText += " (Preview Mode)"
	}

	sb.WriteString(v.renderer.HeaderStyle(headerText))
	sb.WriteString("\n")

	// Render error or status messages if present
	if v.errorMsg != "" {
		sb.WriteString(v.renderer.ErrorStyle(v.errorMsg))
		sb.WriteString("\n")
	} else if v.statusMsg != "" && v.statusMsgTimer > 0 {
		sb.WriteString(v.renderer.GetStyles().Highlight.Render(v.statusMsg))
		sb.WriteString("\n")
	}

	// Main content
	switch v.mode {
	case ModeEdit:
		// Render edit interface
		sb.WriteString(v.renderEditMode())
	case ModePreview:
		// Render preview
		sb.WriteString(v.previewContent)
	}

	// Render footer with shortcuts
	var footerText string
	switch v.mode {
	case ModeEdit:
		if v.activeField == FieldQuestion || v.activeField == FieldAnswer {
			footerText = "Alt+↑/↓: Navigate • Ctrl+p: Preview • Ctrl+s: Save • Esc: Cancel • Ctrl+q: Force Exit"
		} else {
			footerText = "Tab/Shift+Tab: Navigate • Ctrl+p: Preview • Ctrl+s: Save • Esc: Cancel • Ctrl+q: Force Exit"
		}
	case ModePreview:
		footerText = "Ctrl+p: Return to Edit • Ctrl+s: Save • Esc: Cancel • Ctrl+q: Force Exit"
	}

	// Add unsaved changes indicator
	if v.unsavedChanges {
		footerText = "* " + footerText
	}

	sb.WriteString("\n" + v.renderer.FooterStyle(footerText))

	return sb.String()
}

// renderEditMode renders the editor fields in full edit mode
func (v *CardEditView) renderEditMode() string {
	var sb strings.Builder

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

	return sb.String()
}

// initInputs initializes all text inputs with enhanced functionality
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

	// Initialize tags input with improved visuals
	v.tagsInput = textinput.New()
	v.tagsInput.Placeholder = "tag1, tag2, tag3"
	v.tagsInput.CharLimit = 200
	v.tagsInput.Width = v.width - 4

	if len(v.card.Tags) > 0 {
		v.tagsInput.SetValue(strings.Join(v.card.Tags, ", "))
	}

	// Initialize question input as textarea
	v.questionInput = textarea.New()
	v.questionInput.Placeholder = "Enter question..."
	v.questionInput.SetWidth(v.width - 4)
	v.questionInput.SetHeight(5) // Show 5 lines
	v.questionInput.ShowLineNumbers = true

	if v.card.Question != "" {
		v.questionInput.SetValue(v.card.Question)
	}

	// Initialize answer input as textarea
	v.answerInput = textarea.New()
	v.answerInput.Placeholder = "Enter answer..."
	v.answerInput.SetWidth(v.width - 4)
	v.answerInput.SetHeight(10) // Show 10 lines
	v.answerInput.ShowLineNumbers = true

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

	// Use the renderer which properly formats for terminal
	rendered, err := v.renderer.RenderMarkdown(content.String())
	if err != nil {
		v.errorMsg = fmt.Sprintf("Error rendering preview: %v", err)
		v.previewContent = content.String() // Fallback to plain text
	} else {
		v.previewContent = rendered
	}
}

// saveCard saves the card to storage
func (v *CardEditView) saveCard(exit bool) error {
	// Update card from inputs
	v.card.Title = v.titleInput.Value()
	v.card.Question = v.questionInput.Value()
	v.card.Answer = v.answerInput.Value()

	// Parse tags with improved handling
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
		newCard, err := v.store.CreateCardInDeck(
			v.card.Title,
			v.card.Question,
			v.card.Answer,
			v.card.Tags,
			deckObj,
		)

		if err != nil {
			return err
		}

		// Update reference to the saved card
		v.card = newCard
		v.isNewCard = false
	} else {
		// For existing cards, update
		err := v.store.SaveCard(v.card)
		if err != nil {
			return err
		}
	}

	// Reset unsaved changes flag and exit warning
	v.unsavedChanges = false
	v.warnedAboutExit = false

	return nil
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
