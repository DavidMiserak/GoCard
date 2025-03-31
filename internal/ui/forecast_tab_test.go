// File: internal/ui/forecast_tab_test.go

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

// Forecast Tab Tests
func TestGetCardsDueOnDate(t *testing.T) {
	// Create a store with cards due on specific dates
	tomorrow := time.Now().AddDate(0, 0, 1)
	dayAfterTomorrow := time.Now().AddDate(0, 0, 2)

	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:         "card-1",
						NextReview: tomorrow.Add(1 * time.Hour), // Due tomorrow
					},
					{
						ID:         "card-2",
						NextReview: tomorrow.Add(2 * time.Hour), // Also due tomorrow
					},
					{
						ID:         "card-3",
						NextReview: dayAfterTomorrow, // Not due tomorrow
					},
				},
			},
		},
	}

	// Expected: 2 cards due tomorrow
	expectedCount := 2
	actualCount := getCardsDueOnDate(testStore, tomorrow)

	if actualCount != expectedCount {
		t.Errorf("Expected cards due tomorrow to be %d, got %d", expectedCount, actualCount)
	}
}

func TestGetCardsDueInNextDays(t *testing.T) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	threeDaysFromNow := now.AddDate(0, 0, 3)
	sevenDaysFromNow := now.AddDate(0, 0, 7)

	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:         "card-1",
						NextReview: tomorrow,
					},
					{
						ID:         "card-2",
						NextReview: threeDaysFromNow,
					},
					{
						ID:         "card-3",
						NextReview: sevenDaysFromNow,
					},
					{
						ID:         "card-4",
						NextReview: now.AddDate(0, 0, 10), // Outside the 7-day window
					},
				},
			},
		},
	}

	// Expected: 3 cards due in the next 7 days
	expectedCount := 3
	actualCount := getCardsDueInNextDays(testStore, 7)

	if actualCount != expectedCount {
		t.Errorf("Expected cards due in 7 days to be %d, got %d", expectedCount, actualCount)
	}
}

func TestCalculateNewCardsPerDay(t *testing.T) {
	store := createTestStoreForForcast()

	// This is a fixed function in the code, so we just check it returns a reasonable value
	result := calculateNewCardsPerDay(store)

	if result <= 0 {
		t.Errorf("Expected new cards per day to be positive, got %d", result)
	}
}

func TestCalculateReviewsPerDay(t *testing.T) {
	store := createTestStoreForForcast()

	// This is a fixed function in the code, so we just check it returns a reasonable value
	result := calculateReviewsPerDay(store)

	if result <= 0 {
		t.Errorf("Expected reviews per day to be positive, got %d", result)
	}
}

func TestGenerateForecastData(t *testing.T) {
	// Create dates centered around "now" to control test timing
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1).Truncate(24 * time.Hour) // Ensure clean day boundary

	// Create test cards with controlled dates and intervals
	testStore := &data.Store{
		Decks: []model.Deck{
			{
				ID: "test-deck",
				Cards: []model.Card{
					{
						ID:         "card-1",
						NextReview: tomorrow.Add(5 * time.Hour), // Due tomorrow, time doesn't matter
						Interval:   0,                           // New card
					},
					{
						ID:         "card-2",
						NextReview: tomorrow.Add(10 * time.Hour), // Due tomorrow, time doesn't matter
						Interval:   5,                            // Review card (Interval > 0)
					},
				},
			},
		},
	}

	// Generate forecast for the next 3 days
	forecast := generateForecastData(testStore, 3)

	// Validate the forecast length
	if len(forecast) != 3 {
		t.Errorf("Expected forecast length to be 3, got %d", len(forecast))
	}

	// Find the index for tomorrow's date in the forecast
	tomorrowIndex := -1
	for i, day := range forecast {
		if isSameDay(day.Date, tomorrow) {
			tomorrowIndex = i
			break
		}
	}

	if tomorrowIndex == -1 {
		t.Fatalf("Tomorrow's date not found in forecast data")
	}

	// Check counts for tomorrow
	if forecast[tomorrowIndex].NewDue != 1 {
		t.Errorf("Expected 1 new card due tomorrow, got %d", forecast[tomorrowIndex].NewDue)
	}

	if forecast[tomorrowIndex].ReviewDue != 1 {
		t.Errorf("Expected 1 review card due tomorrow, got %d", forecast[tomorrowIndex].ReviewDue)
	}
}

// Helper function to check if two dates are the same day
func isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func TestRenderForecastLegend(t *testing.T) {
	result := renderForecastLegend()

	if result == "" {
		t.Error("Expected renderForecastLegend to return a non-empty string")
	}

	// Check if the result contains expected text
	if !strings.Contains(result, "Review") {
		t.Error("Expected legend to contain 'Review'")
	}

	if !strings.Contains(result, "New") {
		t.Error("Expected legend to contain 'New'")
	}
}

func TestRenderHorizontalForecastChart(t *testing.T) {
	// Create test forecast data
	today := time.Now()
	forecast := []ForecastDay{
		{
			Date:      today,
			ReviewDue: 10,
			NewDue:    5,
		},
		{
			Date:      today.AddDate(0, 0, 1),
			ReviewDue: 8,
			NewDue:    3,
		},
	}

	// Render the chart
	result := renderHorizontalForecastChart(forecast)

	// Check for basic content
	if result == "" {
		t.Error("Expected renderHorizontalForecastChart to return a non-empty string")
	}

	// Check for "Today" label
	if !strings.Contains(result, "Today") {
		t.Error("Expected chart to contain 'Today' label")
	}
}

func TestRenderReviewForecastStats(t *testing.T) {
	store := createTestStore()

	// Just test that rendering doesn't panic and returns a non-empty string
	result := renderReviewForecastStats(store)

	if result == "" {
		t.Error("Expected renderReviewForecastStats to return a non-empty string")
	}

	// Check if the result contains expected headers
	expectedHeaders := []string{
		"Due Today:",
		"Due Tomorrow:",
		"Due This Week:",
		"New Cards/Day:",
		"Reviews/Day",
		"Cards Due by Day",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Expected output to contain '%s'", header)
		}
	}
}

// Helper function to create a test store
func createTestStoreForForcast() *data.Store {
	return data.NewStore()
}
