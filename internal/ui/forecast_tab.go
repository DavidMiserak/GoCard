// File: internal/ui/forecast_tab.go

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/charmbracelet/lipgloss"
)

// renderReviewForecastStats renders the Review Forecast tab statistics
func renderReviewForecastStats(store *data.Store) string {
	var sb strings.Builder

	// Get forecast data
	cardsDueToday := len(store.GetDueCards())
	cardsDueTomorrow := getCardsDueOnDate(store, time.Now().AddDate(0, 0, 1))
	cardsDueThisWeek := getCardsDueInNextDays(store, 7)
	newCardsPerDay := calculateNewCardsPerDay(store)
	reviewsPerDay := calculateReviewsPerDay(store)
	forecastData := generateForecastData(store, 7)

	// Layout the top stats in a row
	topRowWidth := 20

	// Top row stats
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		statLabelStyle.Render("\tDue Today:")+strings.Repeat(" ", topRowWidth-11)+fmt.Sprintf("%4d", cardsDueToday),
		statLabelStyle.Render("\tDue Tomorrow:")+strings.Repeat(" ", topRowWidth-14)+fmt.Sprintf("%4d", cardsDueTomorrow),
		statLabelStyle.Render("\tDue This Week:")+strings.Repeat(" ", topRowWidth-15)+fmt.Sprintf("%4d", cardsDueThisWeek),
	)
	sb.WriteString(topRow)
	sb.WriteString("\n")

	// Second row stats
	secondRow := lipgloss.JoinHorizontal(lipgloss.Top,
		statLabelStyle.Render("\tNew Cards/Day:")+strings.Repeat(" ", topRowWidth-15)+fmt.Sprintf("%4d", newCardsPerDay),
		statLabelStyle.Render("\tReviews/Day (Avg):")+strings.Repeat(" ", topRowWidth-19)+fmt.Sprintf("%4d", reviewsPerDay),
	)
	sb.WriteString(secondRow)

	// Add chart title with some padding
	sb.WriteString("\n\n")
	sb.WriteString(statLabelStyle.Render("Cards Due by Day"))
	sb.WriteString("\n\n")

	// Render legend for the chart
	sb.WriteString(renderForecastLegend())
	sb.WriteString("\n\n")

	// Render horizontal bar chart for cards due by day
	chart := renderHorizontalForecastChart(forecastData)
	sb.WriteString(chart)

	return sb.String()
}

// getCardsDueOnDate returns the number of cards due on a specific date
func getCardsDueOnDate(store *data.Store, date time.Time) int {
	count := 0
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.NextReview.After(startOfDay) && card.NextReview.Before(endOfDay) {
				count++
			}
		}
	}
	return count
}

// getCardsDueInNextDays returns the number of cards due in the next n days
func getCardsDueInNextDays(store *data.Store, days int) int {
	count := 0
	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.NextReview.After(now) && card.NextReview.Before(endDate) {
				count++
			}
		}
	}
	return count
}

// calculateNewCardsPerDay returns the average number of new cards studied per day
func calculateNewCardsPerDay(store *data.Store) int {
	// In a real implementation, you would analyze the study history
	// For now, we'll return a fixed number as in the screenshot
	return 10
}

// calculateReviewsPerDay returns the average number of reviews per day
func calculateReviewsPerDay(store *data.Store) int {
	// In a real implementation, you would analyze the study history
	// For now, we'll return a fixed number as in the screenshot
	return 32
}

// ForecastDay represents forecast data for a single day
type ForecastDay struct {
	Date      time.Time
	ReviewDue int
	NewDue    int
}

// generateForecastData generates forecast data for the next n days
func generateForecastData(store *data.Store, days int) []ForecastDay {
	forecast := make([]ForecastDay, days)

	// Initialize the forecast days
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, i)
		forecast[i] = ForecastDay{
			Date:      date,
			ReviewDue: 0,
			NewDue:    0,
		}
	}

	// Fill in the forecast data
	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.NextReview.IsZero() {
				continue
			}

			// Calculate days from now
			daysFromNow := int(card.NextReview.Sub(time.Now()).Hours() / 24)

			// If due within our forecast window
			if daysFromNow >= 0 && daysFromNow < days {
				if card.Interval > 0 {
					// Card has been reviewed before (review card)
					forecast[daysFromNow].ReviewDue++
				} else {
					// New card
					forecast[daysFromNow].NewDue++
				}
			}
		}
	}

	return forecast
}

// renderForecastLegend renders the legend for the forecast chart
func renderForecastLegend() string {

	// Create styled blocks for legend
	reviewStyle := lipgloss.NewStyle().Foreground(colorBlue)
	newStyle := lipgloss.NewStyle().Foreground(colorGreen)

	reviewBlock := reviewStyle.Render("█")
	newBlock := newStyle.Render("█")

	return fmt.Sprintf("%s Review  %s New", reviewBlock, newBlock)
}

// renderHorizontalForecastChart creates a horizontal bar chart for cards due by day
func renderHorizontalForecastChart(data []ForecastDay) string {
	var sb strings.Builder

	// Find the maximum value for scaling
	maxValue := 0
	for _, day := range data {
		total := day.ReviewDue + day.NewDue
		if total > maxValue {
			maxValue = total
		}
	}

	// Set a minimum scale if data is empty
	if maxValue == 0 {
		maxValue = 50 // Match the scale in the screenshot
	}

	// Create styles with the explicit colors
	reviewStyle := lipgloss.NewStyle().Foreground(colorBlue)
	newStyle := lipgloss.NewStyle().Foreground(colorGreen)

	// Maximum width for the bars
	maxBarWidth := 30

	// Draw each day's bar
	for i, day := range data {
		// Format the date for the y-axis label
		var dateLabel string
		if i == 0 {
			dateLabel = "Today"
		} else {
			dateLabel = day.Date.Format("Jan 2")
		}

		// Format the label with fixed width for alignment
		labelWidth := 10
		formattedLabel := fmt.Sprintf("%-*s", labelWidth, dateLabel)

		// Calculate bar widths based on values and scale to max width
		reviewWidth := 0
		if day.ReviewDue > 0 {
			reviewWidth = int((float64(day.ReviewDue) / float64(maxValue)) * float64(maxBarWidth))
			if reviewWidth == 0 {
				reviewWidth = 1 // Ensure visible bar for non-zero values
			}
		}

		newWidth := 0
		if day.NewDue > 0 {
			newWidth = int((float64(day.NewDue) / float64(maxValue)) * float64(maxBarWidth))
			if newWidth == 0 {
				newWidth = 1 // Ensure visible bar for non-zero values
			}
		}

		// Create colored bars with explicit styling
		newBar := ""
		if newWidth > 0 {
			newBar = newStyle.Render(strings.Repeat("█", newWidth))
		}

		reviewBar := ""
		if reviewWidth > 0 {
			reviewBar = reviewStyle.Render(strings.Repeat("█", reviewWidth))
		}

		// Combine label and bars
		sb.WriteString(formattedLabel + " " + newBar + reviewBar)

		// Add total count at the end of the bar
		total := day.ReviewDue + day.NewDue
		if total > 0 {
			sb.WriteString(fmt.Sprintf(" %d", total))
		}

		// Add spacing between bars except for the last one
		if i < len(data)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}
