// File: internal/ui/forecast_tab.go

package ui

import (
	"github.com/DavidMiserak/GoCard/internal/data"
	"strings"
)

// renderReviewForecastStats renders the Review Forecast tab statistics
func renderReviewForecastStats(store *data.Store) string {
	var sb strings.Builder

	// Placeholder implementation
	// TODO: Replace with actual statistics
	sb.WriteString(statLabelStyle.Render("This is the Review Forecast Tab"))
	return sb.String()
}
