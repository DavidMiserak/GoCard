// Package views contains the different UI views for GoCard.
package views

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/ui/input"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	ViewReview      ViewState = iota
	ViewDeckList              // Deck list navigation view
	ViewDeckBrowser           // Deck browser (current deck details)
	ViewDeckStats
	ViewCreateDeck
	ViewRenameDeck
	ViewDeleteDeck
	ViewMoveToDeck
	ViewCreateCard
	ViewEditCard
	ViewSearch
	ViewSearchResults
)

// ReviewState represents the current state of the review session
type ReviewState int

const (
	StateShowingQuestion ReviewState = iota
	StateShowingAnswer
	StateCompleted
)

// View is the interface that all views must implement
type View interface {
	// Init initializes the view and returns any initial commands
	Init() tea.Cmd

	// Update handles messages and returns updated view and commands
	Update(msg tea.Msg, keys input.KeyMap) (View, tea.Cmd)

	// Render returns the view as a string
	Render(width, height int) string

	// State returns the current ViewState
	State() ViewState
}

// BaseView provides common functionality for all views
type BaseView struct {
	state    ViewState
	viewport viewport.Model
	width    int
	height   int
	error    string
}

// NewBaseView creates a new base view with common initialization
func NewBaseView(state ViewState, width, height int) BaseView {
	vp := viewport.New(width, height-6) // Leave room for header and footer
	vp.SetContent("")

	return BaseView{
		state:    state,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// State returns the current ViewState
func (v BaseView) State() ViewState {
	return v.state
}

// SetError sets an error message to be displayed
func (v *BaseView) SetError(msg string) {
	v.error = msg
}

// GetError returns the current error message
func (v BaseView) GetError() string {
	return v.error
}

// SetDimensions updates the view dimensions
func (v *BaseView) SetDimensions(width, height int) {
	v.width = width
	v.height = height
	v.viewport.Width = width
	v.viewport.Height = height - 6 // Leave room for header and footer
}

// UpdateViewport updates the viewport content
func (v *BaseView) UpdateViewport(content string) {
	v.viewport.SetContent(content)
	v.viewport.GotoTop()
}
