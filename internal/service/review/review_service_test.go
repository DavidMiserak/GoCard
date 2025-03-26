// internal/service/review/review_service_test.go
package review

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// Mock implementations for dependencies
type mockStorageService struct {
	cards      map[string]domain.Card
	loadCardFn func(string) (domain.Card, error)
}

func (m *mockStorageService) Initialize(rootDir string) error { return nil }
func (m *mockStorageService) Close() error                    { return nil }
func (m *mockStorageService) LoadCard(filePath string) (domain.Card, error) {
	if m.loadCardFn != nil {
		return m.loadCardFn(filePath)
	}
	card, ok := m.cards[filePath]
	if !ok {
		return domain.Card{}, errors.New("card not found")
	}
	return card, nil
}
func (m *mockStorageService) UpdateCardMetadata(card domain.Card) error { return nil }
func (m *mockStorageService) ListCardPaths(deckPath string) ([]string, error) {
	return []string{}, nil
}
func (m *mockStorageService) ParseFrontmatter(content []byte) (map[string]interface{}, []byte, error) {
	return nil, nil, nil
}
func (m *mockStorageService) UpdateFrontmatter(content []byte, updates map[string]interface{}) ([]byte, error) {
	return nil, nil
}
func (m *mockStorageService) LoadDeck(dirPath string) (domain.Deck, error) {
	return domain.Deck{Path: dirPath, Name: "Test Deck"}, nil
}
func (m *mockStorageService) ListDeckPaths(parentPath string) ([]string, error) {
	return []string{}, nil
}
func (m *mockStorageService) FindCardsByTag(tag string) ([]domain.Card, error) {
	return []domain.Card{}, nil
}
func (m *mockStorageService) SearchCards(query string) ([]domain.Card, error) {
	return []domain.Card{}, nil
}

type mockCardService struct {
	isDueFn      func(string) bool
	getDueDateFn func(string) time.Time
	reviewCardFn func(string, int) error
}

func (m *mockCardService) GetCard(cardPath string) (domain.Card, error) {
	return domain.Card{FilePath: cardPath, Title: "Test Card"}, nil
}
func (m *mockCardService) ReviewCard(cardPath string, rating int) error {
	if m.reviewCardFn != nil {
		return m.reviewCardFn(cardPath, rating)
	}
	return nil
}
func (m *mockCardService) IsDue(cardPath string) bool {
	if m.isDueFn != nil {
		return m.isDueFn(cardPath)
	}
	return true
}
func (m *mockCardService) GetDueDate(cardPath string) time.Time {
	if m.getDueDateFn != nil {
		return m.getDueDateFn(cardPath)
	}
	return time.Now()
}

type mockDeckService struct {
	getDueCardsFn func(string) ([]domain.Card, error)
}

func (m *mockDeckService) GetDeck(deckPath string) (domain.Deck, error) {
	return domain.Deck{Path: deckPath, Name: "Test Deck"}, nil
}
func (m *mockDeckService) GetSubdecks(deckPath string) ([]domain.Deck, error) {
	return []domain.Deck{}, nil
}
func (m *mockDeckService) GetParentDeck(deckPath string) (domain.Deck, error) {
	return domain.Deck{}, nil
}
func (m *mockDeckService) GetCards(deckPath string) ([]domain.Card, error) {
	return []domain.Card{}, nil
}
func (m *mockDeckService) GetDueCards(deckPath string) ([]domain.Card, error) {
	if m.getDueCardsFn != nil {
		return m.getDueCardsFn(deckPath)
	}
	cards := []domain.Card{
		{FilePath: "card1.md", Title: "Card 1"},
		{FilePath: "card2.md", Title: "Card 2"},
		{FilePath: "card3.md", Title: "Card 3"},
	}
	return cards, nil
}
func (m *mockDeckService) GetCardStats(deckPath string) (map[string]int, error) {
	return map[string]int{}, nil
}

func TestNewReviewService(t *testing.T) {
	// Create mocks
	storage := &mockStorageService{}
	cardSvc := &mockCardService{}
	deckSvc := &mockDeckService{}
	alg := algorithm.NewSM2Algorithm()

	// Create the service
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)
	if rs == nil {
		t.Fatal("NewReviewService returned nil")
	}

	// Verify it's the correct type
	_, ok := rs.(*DefaultReviewService)
	if !ok {
		t.Fatal("NewReviewService did not return a *DefaultReviewService")
	}
}

func TestStartSession(t *testing.T) {
	// Test cases
	testCases := []struct {
		name            string
		deckPath        string
		getDueCardsFn   func(string) ([]domain.Card, error)
		expectError     bool
		expectCardCount int
	}{
		{
			name:            "successful session start",
			deckPath:        "/test/deck",
			getDueCardsFn:   nil, // Use default implementation
			expectError:     false,
			expectCardCount: 3,
		},
		{
			name:            "no due cards",
			deckPath:        "/test/deck",
			getDueCardsFn:   func(string) ([]domain.Card, error) { return []domain.Card{}, nil },
			expectError:     false,
			expectCardCount: 0,
		},
		{
			name:            "error getting due cards",
			deckPath:        "/test/deck",
			getDueCardsFn:   func(string) ([]domain.Card, error) { return nil, errors.New("test error") },
			expectError:     true,
			expectCardCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mocks
			storage := &mockStorageService{}
			cardSvc := &mockCardService{}
			deckSvc := &mockDeckService{
				getDueCardsFn: tc.getDueCardsFn,
			}
			alg := algorithm.NewSM2Algorithm()
			rs := NewReviewService(storage, cardSvc, deckSvc, alg)

			// Start a session
			session, err := rs.StartSession(tc.deckPath)

			// Check error
			if tc.expectError && err == nil {
				t.Error("expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("did not expect an error but got: %v", err)
			}

			// If no error, check session properties
			if !tc.expectError {
				if session.DeckPath != tc.deckPath {
					t.Errorf("expected DeckPath %s, got %s", tc.deckPath, session.DeckPath)
				}
				if len(session.CardPaths) != tc.expectCardCount {
					t.Errorf("expected %d cards, got %d", tc.expectCardCount, len(session.CardPaths))
				}
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	// Setup
	storage := &mockStorageService{}
	cardSvc := &mockCardService{}
	deckSvc := &mockDeckService{}
	alg := algorithm.NewSM2Algorithm()
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)

	// Test getting session when none exists
	_, err := rs.GetSession()
	if err == nil {
		t.Error("expected error for non-existent session, but got nil")
	}

	// Start a session
	deckPath := "/test/deck"
	_, err = rs.StartSession(deckPath)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Get the session
	session, err := rs.GetSession()
	if err != nil {
		t.Errorf("unexpected error getting session: %v", err)
	}

	// Verify session properties
	if session.DeckPath != deckPath {
		t.Errorf("expected DeckPath %s, got %s", deckPath, session.DeckPath)
	}
}

func TestEndSession(t *testing.T) {
	// Setup
	storage := &mockStorageService{}
	cardSvc := &mockCardService{}
	deckSvc := &mockDeckService{}
	alg := algorithm.NewSM2Algorithm()
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)

	// Test ending non-existent session
	_, err := rs.EndSession()
	if err == nil {
		t.Error("expected error ending non-existent session, but got nil")
	}

	// Start a session
	deckPath := "/test/deck"
	_, err = rs.StartSession(deckPath)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// End the session
	summary, err := rs.EndSession()
	if err != nil {
		t.Errorf("unexpected error ending session: %v", err)
	}

	// Verify summary properties
	if summary.DeckPath != deckPath {
		t.Errorf("expected DeckPath %s, got %s", deckPath, summary.DeckPath)
	}

	// Verify session is cleared
	_, err = rs.GetSession()
	if err == nil {
		t.Error("expected error after ending session, but got nil")
	}
}

func TestGetNextCard(t *testing.T) {
	// Setup
	storage := &mockStorageService{
		cards: map[string]domain.Card{
			"card1.md": {FilePath: "card1.md", Title: "Card 1"},
			"card2.md": {FilePath: "card2.md", Title: "Card 2"},
		},
	}
	cardSvc := &mockCardService{}
	deckSvc := &mockDeckService{
		getDueCardsFn: func(string) ([]domain.Card, error) {
			return []domain.Card{
				{FilePath: "card1.md", Title: "Card 1"},
				{FilePath: "card2.md", Title: "Card 2"},
			}, nil
		},
	}
	alg := algorithm.NewSM2Algorithm()
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)

	// Test with no active session
	_, err := rs.GetNextCard()
	if err == nil {
		t.Error("expected error with no active session, but got nil")
	}

	// Start a session
	_, err = rs.StartSession("/test/deck")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Get the first card
	card, err := rs.GetNextCard()
	if err != nil {
		t.Errorf("unexpected error getting next card: %v", err)
	}

	// Verify we got a valid card, but don't check exact file path since they're shuffled
	if card.FilePath == "" {
		t.Error("expected a valid card, got empty filepath")
	}

	// Verify it's one of our expected cards
	validPaths := map[string]bool{
		"card1.md": true,
		"card2.md": true,
	}

	if !validPaths[card.FilePath] {
		t.Errorf("expected card path to be either card1.md or card2.md, got %s", card.FilePath)
	}
}

func TestSubmitRating(t *testing.T) {
	// Setup with a mock card service to test review card calls
	storage := &mockStorageService{
		cards: map[string]domain.Card{
			"card1.md": {FilePath: "card1.md", Title: "Card 1"},
		},
	}
	reviewCardCalled := false
	cardSvc := &mockCardService{
		reviewCardFn: func(cardPath string, rating int) error {
			reviewCardCalled = true
			if cardPath != "card1.md" {
				t.Errorf("expected to review card1.md, got %s", cardPath)
			}
			if rating != 5 {
				t.Errorf("expected rating 5, got %d", rating)
			}
			return nil
		},
	}
	deckSvc := &mockDeckService{
		getDueCardsFn: func(string) ([]domain.Card, error) {
			return []domain.Card{
				{FilePath: "card1.md", Title: "Card 1"},
			}, nil
		},
	}
	alg := algorithm.NewSM2Algorithm()
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)

	// Test with no active session
	err := rs.SubmitRating(5)
	if err == nil {
		t.Error("expected error with no active session, but got nil")
	}

	// Start a session
	_, err = rs.StartSession("/test/deck")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Submit a rating
	err = rs.SubmitRating(5)
	if err != nil {
		t.Errorf("unexpected error submitting rating: %v", err)
	}
	if !reviewCardCalled {
		t.Error("expected ReviewCard to be called on card service")
	}

	// Session should be complete now, verify next GetNextCard returns error
	_, err = rs.GetNextCard()
	if err == nil {
		t.Error("expected error getting next card after session complete, but got nil")
	}
}

func TestGetSessionStats(t *testing.T) {
	// Setup
	storage := &mockStorageService{}
	cardSvc := &mockCardService{}
	deckSvc := &mockDeckService{}
	alg := algorithm.NewSM2Algorithm()
	rs := NewReviewService(storage, cardSvc, deckSvc, alg)

	// Test with no active session
	_, err := rs.GetSessionStats()
	if err == nil {
		t.Error("expected error with no active session, but got nil")
	}

	// Start a session
	_, err = rs.StartSession("/test/deck")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Get stats before any ratings
	stats, err := rs.GetSessionStats()
	if err != nil {
		t.Errorf("unexpected error getting session stats: %v", err)
	}
	if stats["total_cards"].(int) != 3 {
		t.Errorf("expected 3 total cards, got %d", stats["total_cards"].(int))
	}
	if stats["completed_cards"].(int) != 0 {
		t.Errorf("expected 0 completed cards, got %d", stats["completed_cards"].(int))
	}

	// Submit a rating to advance the session
	err = rs.SubmitRating(4)
	if err != nil {
		t.Fatalf("failed to submit rating: %v", err)
	}

	// Get stats after a rating
	stats, err = rs.GetSessionStats()
	if err != nil {
		t.Errorf("unexpected error getting session stats: %v", err)
	}
	if stats["completed_cards"].(int) != 1 {
		t.Errorf("expected 1 completed card, got %d", stats["completed_cards"].(int))
	}
	if stats["average_rating"].(float64) != 4.0 {
		t.Errorf("expected average rating 4.0, got %f", stats["average_rating"].(float64))
	}
}
