// internal/ui/tui/review_model.go

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReviewState tracks the state of the card review
type ReviewState int

const (
	ReviewStateQuestion ReviewState = iota
	ReviewStateAnswer
	ReviewStateRating
	ReviewStateComplete
)

// ReviewModel represents the card review screen
type ReviewModel struct {
	ReviewService interfaces.ReviewService
	CardService   interfaces.CardService
	DeckService   interfaces.DeckService
	State         ReviewState
	Session       domain.ReviewSession
	CurrentCard   domain.Card
	Progress      progress.Model
	DeckPath      string
	DeckName      string
	Keys          ReviewKeyMap
	Width         int
	Height        int
	Error         string
	Stats         map[string]interface{}
	Help          help.Model
	ShowHelp      bool
}

// NewReviewModel creates a new review model
func NewReviewModel(
	reviewService interfaces.ReviewService,
	cardService interfaces.CardService,
	deckService interfaces.DeckService,
) *ReviewModel {
	progressBar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	helpModel := help.New()

	return &ReviewModel{
		ReviewService: reviewService,
		CardService:   cardService,
		DeckService:   deckService,
		State:         ReviewStateQuestion,
		Progress:      progressBar,
		Keys:          DefaultReviewKeyMap(),
		Help:          helpModel,
		ShowHelp:      false,
	}
}

// StartReview initializes a review session for a deck
func (m *ReviewModel) StartReview(deckPath string) tea.Cmd {
	m.DeckPath = deckPath

	// Extract deck name from path
	parts := strings.Split(deckPath, "/")
	if len(parts) > 0 {
		m.DeckName = parts[len(parts)-1]
	} else {
		m.DeckName = "Deck"
	}

	m.State = ReviewStateQuestion
	m.Error = ""

	return tea.Batch(
		m.initReviewSession(),
		m.loadNextCard(),
	)
}

// initReviewSession starts a new review session
func (m *ReviewModel) initReviewSession() tea.Cmd {
	return func() tea.Msg {
		session, err := m.ReviewService.StartSession(m.DeckPath)
		if err != nil {
			return errMsg{fmt.Errorf("failed to start review session: %w", err)}
		}

		m.Session = session

		// Get session stats
		stats, err := m.ReviewService.GetSessionStats()
		if err != nil {
			return errMsg{fmt.Errorf("failed to get session stats: %w", err)}
		}

		m.Stats = stats

		// If no cards due, end the session
		if len(session.CardPaths) == 0 {
			m.State = ReviewStateComplete
			return nil
		}

		return nil
	}
}

// loadNextCard loads the next card in the session
func (m *ReviewModel) loadNextCard() tea.Cmd {
	return func() tea.Msg {
		// Check if session is complete
		session, err := m.ReviewService.GetSession()
		if err != nil {
			return errMsg{fmt.Errorf("failed to get session: %w", err)}
		}

		m.Session = session

		if session.IsComplete() {
			m.State = ReviewStateComplete

			// End the session
			summary, err := m.ReviewService.EndSession()
			if err != nil {
				return errMsg{fmt.Errorf("failed to end session: %w", err)}
			}

			// Update final stats
			m.Stats["average_rating"] = summary.AverageRating
			m.Stats["duration"] = summary.Duration
			m.Stats["cards_reviewed"] = summary.CardsReviewed

			return nil
		}

		// Get the next card
		card, err := m.ReviewService.GetNextCard()
		if err != nil {
			return errMsg{fmt.Errorf("failed to get next card: %w", err)}
		}

		m.CurrentCard = card
		m.State = ReviewStateQuestion

		// Update progress stats
		stats, err := m.ReviewService.GetSessionStats()
		if err == nil {
			m.Stats = stats
		}

		return nil
	}
}

// showAnswer changes the state to show the answer
func (m *ReviewModel) showAnswer() tea.Cmd {
	m.State = ReviewStateAnswer
	return nil
}

// showRatingPrompt changes the state to prompt for a rating
func (m *ReviewModel) showRatingPrompt() tea.Cmd {
	m.State = ReviewStateRating
	return nil
}

// submitRating submits a rating for the current card
func (m *ReviewModel) submitRating(rating int) tea.Cmd {
	return func() tea.Msg {
		err := m.ReviewService.SubmitRating(rating)
		if err != nil {
			return errMsg{fmt.Errorf("failed to submit rating: %w", err)}
		}

		// Update stats
		stats, err := m.ReviewService.GetSessionStats()
		if err == nil {
			m.Stats = stats
		}

		// Load the next card
		return m.loadNextCard()()
	}
}

// Init initializes the review model
func (m *ReviewModel) Init() tea.Cmd {
	return nil
}

// Update handles events and messages for the review model
func (m *ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Progress.Width = min(m.Width-20, 60)
		m.Help.Width = m.Width
		return m, nil

	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, m.Keys.Quit):
			// End the session before quitting
			if m.State != ReviewStateComplete {
				_, _ = m.ReviewService.EndSession()
			}
			return m, func() tea.Msg {
				return ReturnToDeckListMsg{}
			}

		case key.Matches(msg, m.Keys.Help):
			m.ShowHelp = !m.ShowHelp
			return m, nil
		}

		// Handle state-specific keys
		switch m.State {
		case ReviewStateQuestion:
			switch {
			case key.Matches(msg, m.Keys.Space):
				return m, m.showAnswer()
			}

		case ReviewStateAnswer:
			switch {
			case key.Matches(msg, m.Keys.Space):
				return m, m.showRatingPrompt()
			}

		case ReviewStateRating:
			// Handle rating keys (0-5)
			if msg.String() >= "0" && msg.String() <= "5" {
				rating := int(msg.String()[0] - '0')
				return m, m.submitRating(rating)
			}

		case ReviewStateComplete:
			// Any key returns to deck list
			return m, func() tea.Msg {
				return ReturnToDeckListMsg{}
			}
		}

	case errMsg:
		m.Error = msg.Error()
		return m, nil
	}

	// Update progress bar
	progressModel, cmd := m.Progress.Update(msg)
	m.Progress = progressModel.(progress.Model)
	return m, cmd
}

// View renders the review model
func (m *ReviewModel) View() string {
	s := strings.Builder{}

	// Fallback terminal size if not set
	width := m.Width
	if width == 0 {
		width = 80
	}
	height := m.Height
	if height == 0 {
		height = 24
	}

	// Base style for the entire view
	baseStyle := lipgloss.NewStyle().
		Width(width).
		Height(height - 2) // Leave space for help

	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Width(width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	// Create header with deck name
	header := fmt.Sprintf("Reviewing: %s", m.DeckName)
	s.WriteString(headerStyle.Render(header))

	// Display error if any
	if m.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("197")).
			Bold(true).
			Width(width).
			Align(lipgloss.Center).
			MarginTop(1).
			MarginBottom(1)

		s.WriteString(errorStyle.Render("Error: " + m.Error))
	}

	// Different views based on review state
	switch m.State {
	case ReviewStateQuestion:
		s.WriteString(m.renderQuestion())

	case ReviewStateAnswer:
		s.WriteString(m.renderAnswer())

	case ReviewStateRating:
		s.WriteString(m.renderRating())

	case ReviewStateComplete:
		s.WriteString(m.renderComplete())
	}

	// Help view at the bottom
	helpModel := m.Help
	helpModel.Width = width

	helpStyle := lipgloss.NewStyle().
		PaddingTop(1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	var helpText string
	if m.ShowHelp {
		helpText = helpStyle.Render(helpModel.View(m.Keys))
	} else {
		shortHelpStyle := helpStyle.Foreground(lipgloss.Color("241"))
		helpText = shortHelpStyle.Render("Press ? for help")
	}

	return baseStyle.Render(s.String()) + "\n" + helpText
}

// renderQuestion renders the question view
func (m *ReviewModel) renderQuestion() string {
	s := strings.Builder{}

	// Add progress info at the top
	s.WriteString(m.renderProgress())

	// Question box style
	questionBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(min(m.Width-4, 80))

	// Question content
	question := ""
	if m.CurrentCard.Question != "" {
		question = m.CurrentCard.Question
	} else {
		question = "No question content"
	}

	// Card title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	titleText := fmt.Sprintf("Card: %s", m.CurrentCard.Title)

	questionContent := titleStyle.Render(titleText) + "\n" + question

	s.WriteString(lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Center).
		Render(questionBoxStyle.Render(questionContent)))

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(m.Width).
		PaddingTop(2)

	instruction := "Press Space to show answer"
	s.WriteString(instructionStyle.Render(instruction))

	return s.String()
}

// renderAnswer renders the answer view
func (m *ReviewModel) renderAnswer() string {
	s := strings.Builder{}

	// Add progress info at the top
	s.WriteString(m.renderProgress())

	// Question box style
	questionBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		MarginBottom(1).
		Width(min(m.Width-4, 80))

	// Answer box style
	answerBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(min(m.Width-4, 80))

	// Card title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	titleText := fmt.Sprintf("Card: %s", m.CurrentCard.Title)

	// Question content
	question := ""
	if m.CurrentCard.Question != "" {
		question = m.CurrentCard.Question
	} else {
		question = "No question content"
	}

	questionContent := titleStyle.Render(titleText) + "\n" + question

	// Answer content
	answer := ""
	if m.CurrentCard.Answer != "" {
		answer = m.CurrentCard.Answer
	} else {
		answer = "No answer content"
	}

	answerLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	answerContent := answerLabelStyle.Render("Answer:") + "\n" + answer

	// Render question and answer
	contentStyle := lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Center)

	s.WriteString(contentStyle.Render(questionBoxStyle.Render(questionContent)))
	s.WriteString(contentStyle.Render(answerBoxStyle.Render(answerContent)))

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(m.Width).
		PaddingTop(1)

	instruction := "Press Space to rate your recall"
	s.WriteString(instructionStyle.Render(instruction))

	return s.String()
}

// renderRating renders the rating prompt
func (m *ReviewModel) renderRating() string {
	s := strings.Builder{}

	// Add progress info at the top
	s.WriteString(m.renderProgress())

	// Rating instructions
	ratingBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(min(m.Width-4, 80))

	ratingTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	titleText := "How well did you recall this card?"

	ratingOptions := []struct {
		key   string
		label string
		desc  string
	}{
		{"0", "Again", "Complete blackout, need to relearn"},
		{"1", "Hard", "Significant effort, barely recalled"},
		{"2", "Okay", "Difficult but remembered"},
		{"3", "Good", "Some effort needed"},
		{"4", "Easy", "Clear recall, minor hesitation"},
		{"5", "Perfect", "Perfect recall, very easy"},
	}

	// Build rating options text
	var optionsText string
	for _, opt := range ratingOptions {
		optLine := fmt.Sprintf("%s - %s: %s",
			lipgloss.NewStyle().Bold(true).Render(opt.key),
			lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(opt.label),
			opt.desc)
		optionsText += optLine + "\n"
	}

	ratingContent := ratingTitleStyle.Render(titleText) + "\n\n" + optionsText

	// Render rating box
	contentStyle := lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Center)

	s.WriteString(contentStyle.Render(ratingBoxStyle.Render(ratingContent)))

	return s.String()
}

// renderComplete renders the session completion view
func (m *ReviewModel) renderComplete() string {
	s := strings.Builder{}

	// Session summary box
	summaryBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(min(m.Width-4, 80))

	summaryTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginBottom(1)

	titleText := "Review Session Complete!"

	// Get stats from the review session
	cardsReviewed := 0
	if val, ok := m.Stats["completed_cards"]; ok {
		if intVal, ok := val.(int); ok {
			cardsReviewed = intVal
		}
	}

	averageRating := 0.0
	if val, ok := m.Stats["average_rating"]; ok {
		if floatVal, ok := val.(float64); ok {
			averageRating = floatVal
		}
	}

	// Format duration
	durationText := "N/A"
	if val, ok := m.Stats["duration"]; ok {
		if duration, ok := val.(time.Duration); ok {
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			durationText = fmt.Sprintf("%dm %ds", minutes, seconds)
		}
	}

	// Build stats text
	statsText := fmt.Sprintf("Cards Reviewed: %d\n", cardsReviewed)
	statsText += fmt.Sprintf("Average Rating: %.1f\n", averageRating)
	statsText += fmt.Sprintf("Time Spent: %s\n", durationText)

	// Add message based on cards reviewed
	var messageText string
	if cardsReviewed == 0 {
		messageText = "\nNo cards were due for review in this deck."
	} else {
		messageText = "\nGreat job completing your review session!"
	}

	instructionText := "\nPress any key to return to deck list"

	summaryContent := summaryTitleStyle.Render(titleText) + "\n\n" +
		statsText +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(messageText) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(instructionText)

	// Render summary box
	contentStyle := lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Center).
		PaddingTop(2)

	s.WriteString(contentStyle.Render(summaryBoxStyle.Render(summaryContent)))

	return s.String()
}

// renderProgress renders the progress information
func (m *ReviewModel) renderProgress() string {
	s := strings.Builder{}

	// Progress statistics
	statsStyle := lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Center).
		MarginBottom(1)

	// Get progress percentage
	progress := 0.0
	if val, ok := m.Stats["progress"]; ok {
		if floatVal, ok := val.(float64); ok {
			progress = floatVal / 100.0 // Convert to 0-1 range
		}
	}

	// Get completed and total cards
	completed := 0
	if val, ok := m.Stats["completed_cards"]; ok {
		if intVal, ok := val.(int); ok {
			completed = intVal
		}
	}

	total := 0
	if val, ok := m.Stats["total_cards"]; ok {
		if intVal, ok := val.(int); ok {
			total = intVal
		}
	}

	// Progress text
	progressText := fmt.Sprintf("%d/%d Cards", completed, total)

	// Progress bar
	progressBar := m.Progress.ViewAs(progress)

	// Render progress stats and bar
	s.WriteString(statsStyle.Render(progressText))
	s.WriteString(statsStyle.Render(progressBar))

	return s.String()
}
