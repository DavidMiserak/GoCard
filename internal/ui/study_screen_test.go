// File: internal/ui/study_screen_test.go

package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestStudyScreenView(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Get the first deck ID for testing
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create the study screen
	study := NewStudyScreen(store, deckID)
	if study == nil {
		t.Fatal("Failed to create study screen")
	}

	// Verify the study screen was initialized correctly
	if study.totalCards <= 0 {
		t.Errorf("Expected study screen to have cards, but totalCards is %d", study.totalCards)
	}

	// Test the view rendering
	view := study.View()

	// Check that the view contains expected elements
	expectedElements := []string{
		"Studying:",
		"Card 1/",
		"Press SPACE to reveal answer",
	}

	for _, element := range expectedElements {
		if !strings.Contains(view, element) {
			t.Errorf("Expected view to contain '%s', but it didn't", element)
		}
	}
}

func TestStudyScreenAnswerReveal(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Get the first deck ID for testing
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create the study screen
	study := NewStudyScreen(store, deckID)
	if study == nil {
		t.Fatal("Failed to create study screen")
	}

	// Initially the answer should not be visible
	if study.state != ShowingQuestion {
		t.Errorf("Expected initial state to be ShowingQuestion, got %v", study.state)
	}

	// Simulate pressing SPACE to reveal the answer
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := study.Update(spaceMsg)
	updatedStudy, ok := updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Check that the state changed to showing the answer
	if updatedStudy.state != ShowingAnswer {
		t.Errorf("Expected state to be ShowingAnswer after pressing SPACE, got %v", updatedStudy.state)
	}

	// Check that the view now contains the answer and rating buttons
	view := updatedStudy.View()

	// The view should now contain rating buttons
	expectedElements := []string{
		"Blackout (1)",
		"Wrong (2)",
		"Hard (3)",
		"Good (4)",
		"Easy (5)",
	}

	for _, element := range expectedElements {
		if !strings.Contains(view, element) {
			t.Errorf("Expected view to contain '%s', but it didn't", element)
		}
	}
}

func TestStudyScreenNavigation(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Get the first deck ID for testing
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create the study screen
	study := NewStudyScreen(store, deckID)
	if study == nil {
		t.Fatal("Failed to create study screen")
	}

	// Record initial card index
	initialIndex := study.cardIndex

	// Test skipping to the next card
	skipMsg := tea.KeyMsg{Type: tea.KeyLeft}
	updatedModel, _ := study.Update(skipMsg)
	updatedStudy, ok := updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Check that we moved to the next card
	expectedIndex := (initialIndex + 1) % study.totalCards
	if updatedStudy.cardIndex != expectedIndex {
		t.Errorf("Expected cardIndex to be %d after skipping, got %d", expectedIndex, updatedStudy.cardIndex)
	}

	// Test going back to decks
	backMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	updatedModel, _ = updatedStudy.Update(backMsg)

	// Verify we got a BrowseScreen model back
	_, ok = updatedModel.(*BrowseScreen)
	if !ok {
		t.Fatalf("Expected *BrowseScreen after back key, got %T", updatedModel)
	}
}

func TestStudyScreenRating(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Get the first deck ID for testing
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create the study screen
	study := NewStudyScreen(store, deckID)
	if study == nil {
		t.Fatal("Failed to create study screen")
	}

	// Record initial card index
	initialIndex := study.cardIndex

	// First reveal the answer
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := study.Update(spaceMsg)
	updatedStudy, ok := updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Confirm we're showing the answer
	if updatedStudy.state != ShowingAnswer {
		t.Fatalf("Failed to show answer before rating. State: %v", updatedStudy.state)
	}

	// Rate the card as "Good" (4)
	rateMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}
	updatedModel, _ = updatedStudy.Update(rateMsg)
	updatedStudy, ok = updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Check that we moved to the next card and reset the state
	expectedIndex := (initialIndex + 1) % updatedStudy.totalCards
	if updatedStudy.cardIndex != expectedIndex {
		t.Errorf("Expected cardIndex to be %d after rating, got %d", expectedIndex, updatedStudy.cardIndex)
	}

	if updatedStudy.state != ShowingQuestion {
		t.Errorf("Expected state to be reset to ShowingQuestion after rating, got %v", updatedStudy.state)
	}

	// Check that the card was marked as studied
	if !updatedStudy.studiedCards[initialIndex] {
		t.Errorf("Expected card at index %d to be marked as studied, but it wasn't", initialIndex)
	}
}

// Test for edge case handling
func TestStudyScreenEmptyDeck(t *testing.T) {
	// Create a mock empty study screen to test edge case handling
	study := &StudyScreen{
		store:        nil,
		deckID:       "empty-deck",
		deck:         model.Deck{Name: "Empty Deck"},
		cards:        []model.Card{},
		cardIndex:    0,
		totalCards:   0,
		studiedCards: make(map[int]bool),
		state:        ShowingQuestion,
	}

	// Test the view rendering for empty deck
	view := study.View()

	// Check that the view contains a message about no cards
	if !strings.Contains(view, "No cards in this deck") {
		t.Errorf("Expected view to contain message about no cards, but it didn't")
	}

	// Test the progress bar rendering for empty deck
	progressBar := study.renderProgressBar()
	if len(progressBar) == 0 {
		t.Errorf("Expected progress bar to render something even with empty deck")
	}
}

// New test for the finish studying functionality
func TestStudyScreenFinishStudying(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Get the first deck ID for testing
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create the study screen
	study := NewStudyScreen(store, deckID)
	if study == nil {
		t.Fatal("Failed to create study screen")
	}

	// Mark all cards as studied except for the current one
	for i := 0; i < study.totalCards; i++ {
		if i != study.cardIndex {
			study.studiedCards[i] = true
		}
	}

	// Reveal the answer for the current card
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := study.Update(spaceMsg)
	updatedStudy, ok := updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Rate the last card to mark it as studied
	rateMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}
	updatedModel, _ = updatedStudy.Update(rateMsg)
	updatedStudy, ok = updatedModel.(*StudyScreen)
	if !ok {
		t.Fatalf("Expected *StudyScreen, got %T", updatedModel)
	}

	// Check that we're now in the FinishedStudying state
	if updatedStudy.state != FinishedStudying {
		t.Errorf("Expected state to be FinishedStudying after rating last card, got %v", updatedStudy.state)
	}

	// Check that the view shows the completion message
	view := updatedStudy.View()
	if !strings.Contains(view, "Study Session Complete") {
		t.Errorf("Expected view to contain completion message, but it didn't")
	}

	// Press any key to navigate to stats
	anyKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ = updatedStudy.Update(anyKeyMsg)

	// Verify we got a StatisticsScreen model back
	_, ok = updatedModel.(*StatisticsScreen)
	if !ok {
		t.Fatalf("Expected *StatisticsScreen after key press in FinishedStudying state, got %T", updatedModel)
	}
}

// Test for navigation when all cards have been studied
func TestStudyScreenAllCardsStudied(t *testing.T) {
	// Create a simple study screen with two cards
	cards := []model.Card{
		{ID: "1", Question: "Question 1", Answer: "Answer 1"},
		{ID: "2", Question: "Question 2", Answer: "Answer 2"},
	}

	study := &StudyScreen{
		store:        nil,
		deckID:       "test-deck",
		deck:         model.Deck{Name: "Test Deck", Cards: cards},
		cards:        cards,
		cardIndex:    0,
		totalCards:   2,
		studiedCards: make(map[int]bool),
		state:        ShowingQuestion,
	}

	// Mark the first card as studied
	study.studiedCards[0] = true

	// Navigate to the next card
	study.nextCard()

	// Check that we're now at the second card
	if study.cardIndex != 1 {
		t.Errorf("Expected to navigate to card index 1, got %d", study.cardIndex)
	}

	// Mark the second card as studied
	study.studiedCards[1] = true

	// Try to navigate to the next card
	study.nextCard()

	// Check that we're now in the FinishedStudying state
	if study.state != FinishedStudying {
		t.Errorf("Expected state to be FinishedStudying when all cards are studied, got %v", study.state)
	}
}
