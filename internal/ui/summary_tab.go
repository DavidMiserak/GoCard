// File: internal/ui/summary_tab.go

package ui

import (
	"github.com/DavidMiserak/GoCard/internal/data"
	"strings"
)

// renderSummaryStats renders the Summary tab statistics
func renderSummaryStats(store *data.Store) string {
	var sb strings.Builder

	// Placeholder implementation
	// TODO: Replace with actual statistics
	sb.WriteString(statLabelStyle.Render("This is the Summary Tab"))
	return sb.String()
}
