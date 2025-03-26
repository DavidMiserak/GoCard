// internal/ui/tui/screen_manager.go

package tui

import (
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	tea "github.com/charmbracelet/bubbletea"
)

// ScreenType represents different screens in the application
type ScreenType int

const (
	ScreenWelcome ScreenType = iota
	ScreenDeckList
	ScreenReview
)

// SwitchScreenMsg is a message to switch screens
type SwitchScreenMsg struct {
	Screen ScreenType
	Data   interface{} // Optional data to pass to the next screen
}

// AppModel is the main application model that manages different screens
type AppModel struct {
	deckService    interfaces.DeckService
	cardService    interfaces.CardService
	reviewService  interfaces.ReviewService
	storageService interfaces.StorageService
	currentScreen  ScreenType
	cardsDir       string
	welcomeModel   *WelcomeModel
	deckListModel  *DeckListModel
	reviewModel    *ReviewModel
	width          int
	height         int
}

// NewAppModel creates a new application model
func NewAppModel(
	deckService interfaces.DeckService,
	cardService interfaces.CardService,
	reviewService interfaces.ReviewService,
	storageService interfaces.StorageService,
	cardsDir string,
) *AppModel {
	welcomeModel := NewWelcomeModel()
	deckListModel := NewDeckListModel(deckService, storageService, cardsDir)
	reviewModel := NewReviewModel(reviewService, cardService, deckService)

	return &AppModel{
		deckService:    deckService,
		cardService:    cardService,
		reviewService:  reviewService,
		storageService: storageService,
		currentScreen:  ScreenWelcome,
		cardsDir:       cardsDir,
		welcomeModel:   welcomeModel,
		deckListModel:  deckListModel,
		reviewModel:    reviewModel,
	}
}

// Init initializes the application model
func (m *AppModel) Init() tea.Cmd {
	// Collect all commands to run at startup
	var cmds []tea.Cmd

	cmds = append(cmds, m.welcomeModel.Init())
	cmds = append(cmds, m.deckListModel.Init())
	cmds = append(cmds, m.updateWelcomeStats())

	// Return commands as a batch
	return tea.Batch(cmds...)
}

// updateWelcomeStats prepares stats for the welcome screen
func (m *AppModel) updateWelcomeStats() tea.Cmd {
	return func() tea.Msg {
		// This would be populated from actual data in a real implementation
		deckCount := 0
		totalCards := 0
		dueCards := 0
		newCards := 0
		reviewedCards := 0

		// Get statistics from root directory
		subdecks, err := m.deckService.GetSubdecks(m.cardsDir)
		if err == nil {
			deckCount = len(subdecks)

			for _, deck := range subdecks {
				stats, err := m.deckService.GetCardStats(deck.Path)
				if err == nil {
					totalCards += stats["total"]
					dueCards += stats["due"]
					newCards += stats["new"]
					reviewedCards += stats["learned"]
				}
			}
		}

		// Update welcome model with stats
		m.welcomeModel.SetStats(deckCount, totalCards, dueCards, newCards, reviewedCards)
		return nil
	}
}

// Update handles events and messages for the application model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update window size for all sub-models
		m.width = msg.Width
		m.height = msg.Height

		var welcomeModel tea.Model
		welcomeModel, cmd = m.welcomeModel.Update(msg)
		m.welcomeModel, _ = welcomeModel.(*WelcomeModel)
		cmds = append(cmds, cmd)

		var deckListModel tea.Model
		deckListModel, cmd = m.deckListModel.Update(msg)
		m.deckListModel, _ = deckListModel.(*DeckListModel)
		cmds = append(cmds, cmd)

		var reviewModel tea.Model
		reviewModel, cmd = m.reviewModel.Update(msg)
		m.reviewModel, _ = reviewModel.(*ReviewModel)
		cmds = append(cmds, cmd)

	case SwitchScreenMsg:
		// Handle screen switching
		m.currentScreen = msg.Screen

		switch msg.Screen {
		case ScreenDeckList:
			cmds = append(cmds, m.deckListModel.loadDecks())

		case ScreenReview:
			// Start a review session with the selected deck
			if deckPath, ok := msg.Data.(string); ok {
				cmds = append(cmds, m.reviewModel.StartReview(deckPath))
			}
		}

	case StartReviewMsg:
		// Message to start a review with a specific deck
		m.currentScreen = ScreenReview
		cmds = append(cmds, m.reviewModel.StartReview(msg.DeckPath))

	case ReturnToDeckListMsg:
		// Return to deck list after review
		m.currentScreen = ScreenDeckList
		cmds = append(cmds, m.updateWelcomeStats())

	default:
		// Route messages to the appropriate sub-model based on current screen
		switch m.currentScreen {
		case ScreenWelcome:
			var welcomeModel tea.Model
			welcomeModel, cmd = m.welcomeModel.Update(msg)
			m.welcomeModel, _ = welcomeModel.(*WelcomeModel)
			cmds = append(cmds, cmd)

		case ScreenDeckList:
			var deckListModel tea.Model
			deckListModel, cmd = m.deckListModel.Update(msg)
			m.deckListModel, _ = deckListModel.(*DeckListModel)
			cmds = append(cmds, cmd)

		case ScreenReview:
			var reviewModel tea.Model
			reviewModel, cmd = m.reviewModel.Update(msg)
			m.reviewModel, _ = reviewModel.(*ReviewModel)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current screen
func (m *AppModel) View() string {
	switch m.currentScreen {
	case ScreenWelcome:
		return m.welcomeModel.View()
	case ScreenDeckList:
		return m.deckListModel.View()
	case ScreenReview:
		return m.reviewModel.View()
	default:
		return "Error: Unknown screen"
	}
}
