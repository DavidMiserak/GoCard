// File: internal/ui/study_screen.go

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

// Key mapping for study screen
type studyKeyMap struct {
	ShowAnswer key.Binding
	Skip       key.Binding
	Back       key.Binding
	Quit       key.Binding
	Rate1      key.Binding // Blackout
	Rate2      key.Binding // Wrong
	Rate3      key.Binding // Hard
	Rate4      key.Binding // Good
	Rate5      key.Binding // Easy
}

var studyKeys = studyKeyMap{
	ShowAnswer: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("SPACE", "Show Answer"),
	),
	Skip: key.NewBinding(
		key.WithKeys("<", "left", "h"), // "h" for Vim users
		key.WithHelp("<", "Skip"),
	),
	Back: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "Back to Decks"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "Quit"),
	),
	Rate1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "Blackout"),
	),
	Rate2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "Wrong"),
	),
	Rate3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "Hard"),
	),
	Rate4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "Good"),
	),
	Rate5: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "Easy"),
	),
}

// StudyState represents the current state of the study screen
type StudyState int

const (
	ShowingQuestion StudyState = iota
	ShowingAnswer
	FinishedStudying
)

// StudyScreen represents the screen for studying flashcards
type StudyScreen struct {
	store        *data.Store
	deckID       string
	deck         model.Deck
	cards        []model.Card
	cardIndex    int
	totalCards   int
	studiedCards map[int]bool // Track which cards have been studied
	state        StudyState
	width        int
	height       int
}

// NewStudyScreen creates a new study screen for the specified deck
func NewStudyScreen(store *data.Store, deckID string) *StudyScreen {
	// Get the deck from the store
	deck, found := store.GetDeck(deckID)
	if !found {
		// If the deck is not found, return to the browse screen
		// This should not happen in normal operation but is a safeguard
		return nil
	}

	// Get the cards from the deck
	cards := deck.Cards

	return &StudyScreen{
		store:        store,
		deckID:       deckID,
		deck:         deck,
		cards:        cards,
		cardIndex:    0,
		totalCards:   len(cards),
		studiedCards: make(map[int]bool), // Initialize the map to track studied cards
		state:        ShowingQuestion,
	}
}

// Init initializes the study screen
func (s *StudyScreen) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model
func (s *StudyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in finished state, any key navigates to stats screen
		if s.state == FinishedStudying {
			return NewStatisticsScreenWithDeck(s.store, s.deckID), nil
		}

		// Handle space key explicitly since it's special in Bubble Tea
		if msg.Type == tea.KeySpace && s.state == ShowingQuestion {
			s.state = ShowingAnswer
			return s, nil
		}

		// Handle other keys
		switch {
		case key.Matches(msg, studyKeys.Quit):
			return s, tea.Quit

		case key.Matches(msg, studyKeys.Back):
			// Return to browse decks screen
			return NewBrowseScreen(s.store), nil

		case key.Matches(msg, studyKeys.Skip):
			// Skip this card and go to the next one
			s.nextCard()
		}

		// Handle rating keys when showing the answer
		if s.state == ShowingAnswer {
			// Check if the key pressed is a number between 1-5 for ratings
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				r := msg.Runes[0]
				if r >= '1' && r <= '5' {
					// Mark the current card as studied
					s.studiedCards[s.cardIndex] = true

					// Apply rating and move to next card
					s.nextCard()
				}
			}
		}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// nextCard advances to the next card or transitions to FinishedStudying state
// if all cards have been studied
func (s *StudyScreen) nextCard() {
	// Check if we've studied all cards
	if len(s.studiedCards) >= s.totalCards {
		s.state = FinishedStudying
		return
	}

	// Find the next unstudied card
	originalIndex := s.cardIndex
	for {
		s.cardIndex = (s.cardIndex + 1) % s.totalCards

		// If we've cycled through all cards and returned to our starting point,
		// it means there are no unstudied cards left
		if s.cardIndex == originalIndex {
			// Check if the current card is also studied
			if s.studiedCards[s.cardIndex] {
				s.state = FinishedStudying
				return
			}
			break
		}

		// If this card hasn't been studied yet, break the loop
		if !s.studiedCards[s.cardIndex] {
			break
		}
	}

	s.state = ShowingQuestion
}

// renderProgressBar renders a progress bar showing the current card position
func (s *StudyScreen) renderProgressBar() string {
	width := 50

	// Handle edge cases to prevent errors
	if s.totalCards <= 0 {
		return progressBarEmptyStyle.Render(strings.Repeat(" ", width))
	}

	// Calculate filled portion, ensuring it stays within bounds
	ratio := float64(s.cardIndex+1) / float64(s.totalCards)
	filled := int(ratio * float64(width))

	// Make sure filled is within valid range
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	empty := width - filled

	filledStr := strings.Repeat(" ", filled)
	emptyStr := strings.Repeat(" ", empty)

	return progressBarFilledStyle.Render(filledStr) + progressBarEmptyStyle.Render(emptyStr)
}

// View renders the study screen
func (s *StudyScreen) View() string {
	var sb strings.Builder

	// Handle edge case: no cards in the deck
	if s.totalCards <= 0 {
		return "No cards in this deck. Press 'b' to go back."
	}

	// Handle when user has finished studying all cards
	if s.state == FinishedStudying {
		sb.WriteString(studyTitleStyle.Render("Study Session Complete!"))
		sb.WriteString("\n\n")
		sb.WriteString("You've completed all cards in this deck!")
		sb.WriteString("\n\n")
		sb.WriteString("Press any key to view your statistics.")
		return sb.String()
	}

	// Title and card count
	title := fmt.Sprintf("Studying: %s", s.deck.Name)
	cardCount := fmt.Sprintf("Card %d/%d", s.cardIndex+1, s.totalCards)

	// Get the current card's question and answer
	currentCard := s.cards[s.cardIndex]

	sb.WriteString(studyTitleStyle.Render(title))
	sb.WriteString(strings.Repeat(" ", max(1, s.width-len(title)-len(cardCount))))
	sb.WriteString(cardCountStyle.Render(cardCount))
	sb.WriteString("\n")

	// Progress bar
	sb.WriteString(s.renderProgressBar())
	sb.WriteString("\n\n")

	// Question box
	sb.WriteString(questionStyle.Render(currentCard.Question))
	sb.WriteString("\n\n")

	// Answer or prompt to show answer
	if s.state == ShowingAnswer {
		sb.WriteString(answerStyle.Render(currentCard.Answer))
		sb.WriteString("\n\n")

		// Rating buttons
		blackoutBtn := ratingBlackoutStyle.Render("Blackout (1)")
		wrongBtn := ratingWrongStyle.Render("Wrong (2)")
		hardBtn := ratingHardStyle.Render("Hard (3)")
		goodBtn := ratingGoodStyle.Render("Good (4)")
		easyBtn := ratingEasyStyle.Render("Easy (5)")

		sb.WriteString(blackoutBtn + " " + wrongBtn + " " + hardBtn + " " + goodBtn + " " + easyBtn)
		sb.WriteString("\n\n")

		// Help text for rating state
		sb.WriteString(studyHelpStyle.Render("1-5: Rate Card    b: Back to Decks    q: Quit"))
	} else {
		// Show the prompt to reveal the answer
		sb.WriteString(revealPromptStyle.Render("Press SPACE to reveal answer"))
		sb.WriteString("\n\n")

		// Help text for question state
		sb.WriteString(studyHelpStyle.Render("SPACE: Show Answer    <: Skip    b: Back to Decks    q: Quit"))
	}

	return sb.String()
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
