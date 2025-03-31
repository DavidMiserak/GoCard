// File: internal/ui/summary_tab_test.go

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

// Summary Tab Tests
func TestGetTotalCards(t *testing.T) {
	store := createTestStoreForSummary()

	expectedCount := 14 // Total count of cards across all decks in our test store
	actualCount := getTotalCards(store)

	if actualCount != expectedCount {
		t.Errorf("Expected total cards to be %d, got %d", expectedCount, actualCount)
	}
}

func TestGetCardsStudiedToday(t *testing.T) {
	store := createTestStoreForSummary()

	// We need to modify our test store to have some cards studied today
	today := time.Now()

	// Set a fixed number of cards to have been studied today
	studiedToday := 0
	for i, deck := range store.GetDecks() {
		for j, card := range deck.Cards {
			if i == 0 && j < 2 { // Make two cards in the first deck studied today
				card.LastReviewed = today
				deck.Cards[j] = card
				studiedToday++
			}
		}
		store.Decks[i] = deck
	}

	actualCount := getCardsStudiedToday(store)

	if actualCount != studiedToday {
		t.Errorf("Expected cards studied today to be %d, got %d", studiedToday, actualCount)
	}
}

func TestCalculateRetentionRate(t *testing.T) {
	// We'll create a store with known ratings to test the calculation
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:           "card-1",
						LastReviewed: time.Now(),
						Rating:       5, // Retained (rating >= 4)
					},
					{
						ID:           "card-2",
						LastReviewed: time.Now(),
						Rating:       4, // Retained
					},
					{
						ID:           "card-3",
						LastReviewed: time.Now(),
						Rating:       3, // Not retained
					},
					{
						ID:           "card-4",
						LastReviewed: time.Now(),
						Rating:       2, // Not retained
					},
				},
			},
		},
	}

	// Expected retention rate: 2 retained out of 4 = 50%
	expectedRate := 50
	actualRate := calculateRetentionRate(testStore)

	if actualRate != expectedRate {
		t.Errorf("Expected retention rate to be %d%%, got %d%%", expectedRate, actualRate)
	}
}

func TestGetCardsStudiedPerDay(t *testing.T) {
	// Create a store with cards studied on specific dates
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	dayBeforeYesterday := now.AddDate(0, 0, -2)

	// Format dates to the expected format
	nowStr := now.Format("Jan 2")
	yesterdayStr := yesterday.Format("Jan 2")
	dayBeforeYesterdayStr := dayBeforeYesterday.Format("Jan 2")

	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:           "card-1",
						LastReviewed: now,
					},
					{
						ID:           "card-2",
						LastReviewed: now,
					},
					{
						ID:           "card-3",
						LastReviewed: yesterday,
					},
					{
						ID:           "card-4",
						LastReviewed: dayBeforeYesterday,
					},
					{
						ID:           "card-5",
						LastReviewed: dayBeforeYesterday,
					},
				},
			},
		},
	}

	result := getCardsStudiedPerDay(testStore)

	// Check counts for specific days
	if result[nowStr] != 2 {
		t.Errorf("Expected %s to have 2 cards, got %d", nowStr, result[nowStr])
	}

	if result[yesterdayStr] != 1 {
		t.Errorf("Expected %s to have 1 card, got %d", yesterdayStr, result[yesterdayStr])
	}

	if result[dayBeforeYesterdayStr] != 2 {
		t.Errorf("Expected %s to have 2 cards, got %d", dayBeforeYesterdayStr, result[dayBeforeYesterdayStr])
	}
}

func TestRenderHorizontalBarChart(t *testing.T) {
	// Create test data with specific days that match our format
	data := map[string]int{
		"Mar 29": 10,
		"Mar 30": 20,
		"Mar 31": 5,
	}

	// Render the chart
	result := renderHorizontalBarChart(data, 10)

	// Check for presence of key elements rather than exact matches
	if !strings.Contains(result, "Mar") {
		t.Error("Expected chart to contain month abbreviation 'Mar'")
	}

	if !strings.Contains(result, "29") {
		t.Error("Expected chart to contain day '29'")
	}

	if !strings.Contains(result, "30") {
		t.Error("Expected chart to contain day '30'")
	}

	if !strings.Contains(result, "31") {
		t.Error("Expected chart to contain day '31'")
	}

	// Check that the output contains the values (may be formatted differently)
	if !strings.Contains(result, "10") {
		t.Error("Expected chart to contain value '10'")
	}

	if !strings.Contains(result, "20") {
		t.Error("Expected chart to contain value '20'")
	}

	if !strings.Contains(result, "5") {
		t.Error("Expected chart to contain value '5'")
	}
}

func TestRenderSummaryStats(t *testing.T) {
	store := createTestStoreForSummary()

	// Just test that rendering doesn't panic and returns a non-empty string
	result := renderSummaryStats(store)

	if result == "" {
		t.Error("Expected renderSummaryStats to return a non-empty string")
	}

	// Check if the result contains expected headers
	expectedHeaders := []string{
		"Total Cards:",
		"Cards Due Today:",
		"Studied Today:",
		"Retention Rate:",
		"Cards Studied per Day",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Expected output to contain '%s'", header)
		}
	}
}

// Helper function to create a test store
func createTestStoreForSummary() *data.Store {
	return data.NewStore()
}
