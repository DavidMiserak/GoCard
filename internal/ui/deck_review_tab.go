// File: internal/ui/deck_review_tab.go

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/charmbracelet/lipgloss"
)

// renderDeckReviewStats renders the Deck Review tab statistics
func renderDeckReviewStats(store *data.Store) string {
	var sb strings.Builder

	// Get stats data
	totalCards := getTotalCards(store)
	matureCards := getMatureCards(store)
	newCards := totalCards - matureCards
	successRate := calculateSuccessRate(store)
	avgInterval := calculateAverageInterval(store)
	lastStudied := getLastStudiedDate(store)
	ratingDistribution := calculateRatingDistribution(store)

	// Layout the stats in two columns
	leftWidth := 20
	rightWidth := 20

	// Left column stats
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		statLabelStyle.Render("Total Cards:")+strings.Repeat(" ", leftWidth-12)+fmt.Sprintf("%4d", totalCards),
		statLabelStyle.Render("Mature Cards:")+strings.Repeat(" ", leftWidth-13)+fmt.Sprintf("%4d", matureCards),
		statLabelStyle.Render("New Cards:")+strings.Repeat(" ", leftWidth-10)+fmt.Sprintf("%4d", newCards),
	)

	// Format the average interval with one decimal place
	intervalStr := fmt.Sprintf("%.1f days", avgInterval)
	// Format the last studied date
	lastStudiedStr := formatLastStudied(lastStudied)

	// Right column stats
	rightColumn := lipgloss.JoinVertical(lipgloss.Left,
		statLabelStyle.Render("\tSuccess Rate:")+strings.Repeat(" ", rightWidth-14)+fmt.Sprintf("%3d%%", successRate),
		statLabelStyle.Render("\tAvg. Interval:")+strings.Repeat(" ", rightWidth-14)+intervalStr,
		statLabelStyle.Render("\tLast Studied:")+strings.Repeat(" ", rightWidth-14)+lastStudiedStr,
	)

	// Join columns horizontally
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	sb.WriteString(columns)

	// Add ratings distribution title with some padding
	sb.WriteString("\n\n")
	sb.WriteString(statLabelStyle.Render("Ratings Distribution"))
	sb.WriteString("\n\n")

	// Render ratings distribution chart
	chart := renderRatingsDistribution(ratingDistribution)
	sb.WriteString(chart)

	return sb.String()
}

// getMatureCards returns the number of cards with interval >= 21 days
func getMatureCards(store *data.Store) int {
	count := 0
	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if card.Interval >= 21 {
				count++
			}
		}
	}
	return count
}

// calculateSuccessRate calculates the percentage of reviews rated 3, 4, or 5
func calculateSuccessRate(store *data.Store) int {
	var totalReviewed, successful int

	// Get reviews from the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if !card.LastReviewed.IsZero() && card.LastReviewed.After(thirtyDaysAgo) {
				totalReviewed++
				if card.Rating >= 3 {
					successful++
				}
			}
		}
	}

	if totalReviewed == 0 {
		return 0
	}

	return int((float64(successful) / float64(totalReviewed)) * 100)
}

// calculateAverageInterval calculates the average interval for all reviewed cards
func calculateAverageInterval(store *data.Store) float64 {
	var totalCards, totalInterval int

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if !card.LastReviewed.IsZero() && card.Interval > 0 {
				totalCards++
				totalInterval += card.Interval
			}
		}
	}

	if totalCards == 0 {
		return 0
	}

	return float64(totalInterval) / float64(totalCards)
}

// getLastStudiedDate returns the most recent study date
func getLastStudiedDate(store *data.Store) time.Time {
	var lastDate time.Time

	for _, deck := range store.GetDecks() {
		if deck.LastStudied.After(lastDate) {
			lastDate = deck.LastStudied
		}

		for _, card := range deck.Cards {
			if card.LastReviewed.After(lastDate) {
				lastDate = card.LastReviewed
			}
		}
	}

	return lastDate
}

// formatLastStudied formats the last studied date
func formatLastStudied(lastDate time.Time) string {
	if lastDate.IsZero() {
		return "Never"
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	lastDateDay := lastDate.Truncate(24 * time.Hour)

	if lastDateDay.Equal(today) {
		return "Today"
	} else if lastDateDay.Equal(today.AddDate(0, 0, -1)) {
		return "Yesterday"
	} else {
		return lastDate.Format("Jan 2")
	}
}

// calculateRatingDistribution calculates the distribution of ratings (1-5)
func calculateRatingDistribution(store *data.Store) map[int]int {
	// Initialize the ratings map
	distribution := make(map[int]int)
	for i := 1; i <= 5; i++ {
		distribution[i] = 0
	}

	// Get ratings from the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, deck := range store.GetDecks() {
		for _, card := range deck.Cards {
			if !card.LastReviewed.IsZero() && card.LastReviewed.After(thirtyDaysAgo) && card.Rating >= 1 && card.Rating <= 5 {
				distribution[card.Rating]++
			}
		}
	}

	return distribution
}

// renderRatingsDistribution creates a horizontal bar chart for ratings distribution
func renderRatingsDistribution(distribution map[int]int) string {
	var sb strings.Builder

	// Calculate total reviews to get percentages
	totalReviews := 0
	for _, count := range distribution {
		totalReviews += count
	}

	if totalReviews == 0 {
		return "No ratings data available"
	}

	// Define rating labels and corresponding styles
	ratingLabels := map[int]string{
		1: "Blackout",
		2: "Wrong",
		3: "Hard",
		4: "Good",
		5: "Easy",
	}

	ratingStyles := map[int]lipgloss.Style{
		1: lipgloss.NewStyle().Foreground(ratingBlackoutColor),
		2: lipgloss.NewStyle().Foreground(ratingWrongColor),
		3: lipgloss.NewStyle().Foreground(ratingHardColor),
		4: lipgloss.NewStyle().Foreground(ratingGoodColor),
		5: lipgloss.NewStyle().Foreground(ratingEasyColor),
	}

	// Max width for the bars
	maxBarWidth := 30

	// Render each rating bar
	for i := 1; i <= 5; i++ {
		count := distribution[i]
		percentage := 0
		if totalReviews > 0 {
			percentage = int((float64(count) / float64(totalReviews)) * 100)
		}

		// Format the label with rating number and name
		label := fmt.Sprintf("%-8s (%d)", ratingLabels[i], i)
		labelWidth := 15
		formattedLabel := fmt.Sprintf("%-*s", labelWidth, label)

		// Calculate bar width based on percentage
		barWidth := int((float64(percentage) / 100.0) * float64(maxBarWidth))
		if percentage > 0 && barWidth == 0 {
			barWidth = 1 // Ensure visible bar for non-zero values
		}

		// Draw the bar using the appropriate style
		bar := ""
		if barWidth > 0 {
			bar = ratingStyles[i].Render(strings.Repeat("█", barWidth))
		}

		// Combine label and bar
		sb.WriteString(formattedLabel + " " + bar)

		// Add percentage at the end of the bar
		if percentage > 0 {
			sb.WriteString(fmt.Sprintf(" %d%%", percentage))
		}

		// Add spacing between bars except for the last one
		if i < 5 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}
