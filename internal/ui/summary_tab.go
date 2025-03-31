// File: internal/ui/summary_tab.go

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/charmbracelet/lipgloss"
)

// renderSummaryStats renders the Summary tab statistics
func renderSummaryStats(store *data.Store) string {
	var sb strings.Builder

	// Get stats data
	totalCards := getTotalCards(store)
	cardsDueToday := len(store.GetDueCards())
	studiedToday := getCardsStudiedToday(store)
	retentionRate := calculateRetentionRate(store)
	cardsStudiedPerDay := getCardsStudiedPerDay(store)

	// Layout the stats in two columns
	leftWidth := 20
	rightWidth := 20

	// Left column stats
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		statLabelStyle.Render("Total Cards:")+strings.Repeat(" ", leftWidth-12)+fmt.Sprintf("%4d", totalCards),
		statLabelStyle.Render("Cards Due Today:")+strings.Repeat(" ", leftWidth-16)+fmt.Sprintf("%4d", cardsDueToday),
	)

	// Right column stats
	rightColumn := lipgloss.JoinVertical(lipgloss.Left,
		statLabelStyle.Render("\tStudied Today:")+strings.Repeat(" ", leftWidth-15)+fmt.Sprintf("%4d", studiedToday),
		statLabelStyle.Render("\tRetention Rate:")+strings.Repeat(" ", rightWidth-16)+fmt.Sprintf("%3d%%", retentionRate),
	)

	// Join columns horizontally
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	sb.WriteString(columns)

	// Add chart title with some padding
	sb.WriteString("\n\n")
	sb.WriteString(statLabelStyle.Render("Cards Studied per Day"))
	sb.WriteString("\n\n")

	// Render bar chart for cards studied per day
	chart := renderHorizontalBarChart(cardsStudiedPerDay, 30)
	sb.WriteString(chart)

	return sb.String()
}

// getTotalCards returns the total number of cards across all decks
func getTotalCards(store *data.Store) int {
	count := 0
	for _, deck := range store.GetDecks() {
		count += len(deck.Cards)
	}
	return count
}

// getCardsStudiedToday returns the number of cards studied today
func getCardsStudiedToday(store *data.Store) int {
	count := 0
	today := time.Now().Truncate(24 * time.Hour) // Start of today

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.LastReviewed.After(today) || card.LastReviewed.Equal(today) {
				count++
			}
		}
	}
	return count
}

// calculateRetentionRate calculates retention rate based on card ratings
// Ratings 4-5 are considered "retained"
func calculateRetentionRate(store *data.Store) int {
	var totalReviewed, retained int

	// Get reviews from the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.LastReviewed.After(thirtyDaysAgo) {
				totalReviewed++
				if card.Rating >= 4 {
					retained++
				}
			}
		}
	}

	if totalReviewed == 0 {
		return 0
	}

	return int((float64(retained) / float64(totalReviewed)) * 100)
}

// getCardsStudiedPerDay returns the number of cards studied per day for the last 6 days
func getCardsStudiedPerDay(store *data.Store) map[string]int {
	// Initialize the last 6 days (including today)
	result := make(map[string]int)
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("Jan 2")
		result[dateStr] = 0
	}

	// Count cards studied on each day
	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			// Skip cards that haven't been reviewed
			if card.LastReviewed.IsZero() {
				continue
			}

			// Check if the review was within the last 6 days
			dayDiff := int(time.Since(card.LastReviewed).Hours() / 24)
			if dayDiff <= 5 {
				dateStr := card.LastReviewed.Format("Jan 2")
				result[dateStr]++
			}
		}
	}

	return result
}

// renderHorizontalBarChart creates a text-based horizontal bar chart for cards studied per day
func renderHorizontalBarChart(data map[string]int, maxBarWidth int) string {
	var sb strings.Builder

	// Find the maximum value for scaling
	maxValue := 0
	for _, count := range data {
		if count > maxValue {
			maxValue = count
		}
	}

	// Set a minimum scale if data is empty
	if maxValue == 0 {
		maxValue = 1
	}

	// Sort dates from oldest to newest (last 6 days)
	dates := make([]string, 0, 6)
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dates = append(dates, date.Format("Jan 2"))
	}

	// Draw the bars
	for _, date := range dates {
		count := data[date]

		// Calculate bar width - scale to max width
		barWidth := int((float64(count) / float64(maxValue)) * float64(maxBarWidth))
		if count > 0 && barWidth == 0 {
			barWidth = 1 // Ensure visible bar for non-zero values
		}

		// Format the y-axis label (date)
		labelWidth := 10
		label := fmt.Sprintf("%-*s", labelWidth, date)

		// Draw the bar using block characters
		bar := ""
		if barWidth > 0 {
			bar = lipgloss.NewStyle().Foreground(colorBlue).Render(strings.Repeat("█", barWidth))
		}

		// Combine label and bar
		sb.WriteString(label + " " + bar)

		// Add count at the end of the bar
		if count > 0 {
			sb.WriteString(fmt.Sprintf(" %d", count))
		}

		sb.WriteString("\n\n")
	}

	return sb.String()
}
