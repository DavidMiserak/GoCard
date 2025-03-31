// File: internal/ui/deck_review_tab.go

package ui

import (
	"github.com/DavidMiserak/GoCard/internal/data"
	"strings"
)

// renderDeckReviewStats renders the Deck Review tab statistics
func renderDeckReviewStats(store *data.Store) string {
	var sb strings.Builder

	// Placeholder implementation
	// TODO: Replace with actual statistics
	sb.WriteString(statLabelStyle.Render("This is the Deck Review Tab"))

	return sb.String()
}
