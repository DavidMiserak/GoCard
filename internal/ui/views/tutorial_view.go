// internal/ui/views/tutorial_view.go
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/render"
)

// TutorialStep represents a single step in the tutorial
type TutorialStep struct {
	Title   string
	Content string
	Action  string
}

// TutorialView presents an interactive tutorial for new users
type TutorialView struct {
	BaseView
	store       storage.CardStoreInterface // Changed from pointer to interface
	renderer    *render.Renderer
	steps       []TutorialStep
	currentStep int
}

// NewTutorialView creates a new tutorial view
func NewTutorialView(store storage.CardStoreInterface, width, height int) (*TutorialView, error) {
	baseView := NewBaseView(ViewTutorial, width, height)

	renderer, err := render.NewRenderer(width)
	if err != nil {
		return nil, err
	}

	// Define tutorial steps
	steps := []TutorialStep{
		{
			Title:   "Welcome to GoCard!",
			Content: "GoCard is a file-based spaced repetition system designed for developers and text-oriented learners.\n\nThis tutorial will guide you through the basic features of GoCard.",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "File-Based Storage",
			Content: "All cards are stored as Markdown files in regular directories.\n\nYou can edit these files with any text editor, and GoCard will automatically detect changes.",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "Decks and Organization",
			Content: "Decks in GoCard are represented by directories.\n\n- Press `c` to browse and change decks\n- Press `C` to create a new deck\n- Nested directories = nested decks",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "Reviewing Cards",
			Content: "During review:\n\n1. The question is shown first\n2. Press `Space` to reveal the answer\n3. Rate your recall from 0-5:\n   - 0-2: Difficult/incorrect (short interval)\n   - 3: Correct but required effort (moderate interval)\n   - 4-5: Easy (longer interval)",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "Creating and Editing Cards",
			Content: "To manage your cards:\n\n- Press `n` to create a new card\n- Press `e` to edit the current card\n- Press `d` to delete a card\n- Press `t` to edit tags\n\nYou can also edit the Markdown files directly with your favorite text editor.",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "Navigation and Help",
			Content: "Basic navigation:\n\n- Arrow keys / j,k: Move up/down\n- Enter: Select\n- Esc: Go back\n- ?: Toggle help\n- q: Quit",
			Action:  "Press Space to continue...",
		},
		{
			Title:   "Sample Content",
			Content: "We've created some sample cards and decks to help you get started.\n\nFeel free to explore them, modify them, or delete them as you learn the system.",
			Action:  "Press Space to finish tutorial...",
		},
	}

	view := &TutorialView{
		BaseView:    baseView,
		store:       store,
		renderer:    renderer,
		steps:       steps,
		currentStep: 0,
	}

	view.updateViewport()

	return view, nil
}

// Init implements View.Init
func (v *TutorialView) Init() tea.Cmd {
	return nil
}

// Update implements View.Update
func (v *TutorialView) Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case input.KeyMatches(msg, keys.ShowAnswer) || input.KeyMatches(msg, keys.Back) || input.IsEnterKey(msg):
			// Advance to next step or finish tutorial
			v.currentStep++
			if v.currentStep >= len(v.steps) {
				// Tutorial complete, return to deck browser
				// Type assertion to convert from interface to concrete type
				concreteStore, ok := v.store.(*storage.CardStore)
				if !ok {
					v.SetError("Error: Unable to access card store")
					return v, nil
				}

				deckView, err := NewDeckBrowserView(concreteStore, "", v.width, v.height)
				if err != nil {
					v.SetError(fmt.Sprintf("Error returning to deck browser: %v", err))
					return v, nil
				}
				return deckView, deckView.Init()
			}
			v.updateViewport()

		case input.KeyMatches(msg, keys.Quit):
			return v, tea.Quit
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
func (v *TutorialView) Render(width, height int) string {
	var sb strings.Builder

	// Render header
	step := v.currentStep + 1
	total := len(v.steps)
	headerText := fmt.Sprintf("GoCard Tutorial - Step %d/%d", step, total)
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
	currentStep := v.steps[v.currentStep]
	footerText := currentStep.Action
	sb.WriteString(v.renderer.FooterStyle(footerText))

	return sb.String()
}

// updateViewport updates the viewport content based on the current step
func (v *TutorialView) updateViewport() {
	if v.currentStep >= len(v.steps) {
		v.viewport.SetContent("Tutorial complete!")
		return
	}

	currentStep := v.steps[v.currentStep]
	var content strings.Builder

	// Render step title
	content.WriteString("# " + currentStep.Title + "\n\n")

	// Render step content
	content.WriteString(currentStep.Content)

	// Add progress indicator
	content.WriteString("\n\n")
	for i := 0; i < len(v.steps); i++ {
		if i == v.currentStep {
			content.WriteString("●")
		} else if i < v.currentStep {
			content.WriteString("✓")
		} else {
			content.WriteString("○")
		}
		if i < len(v.steps)-1 {
			content.WriteString(" ")
		}
	}

	// Render as markdown
	contentMd, err := v.renderer.RenderMarkdown(content.String())
	if err != nil {
		v.SetError(fmt.Sprintf("Error rendering markdown: %v", err))
		v.viewport.SetContent(content.String())
	} else {
		v.viewport.SetContent(contentMd)
	}
	v.viewport.GotoTop()
}
