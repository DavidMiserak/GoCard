// Package storage provides compatibility functions for UI integration.
package storage

import (
	"github.com/DavidMiserak/GoCard/internal/storage/parser"
)

// RenderMarkdown is a compatibility function for UI components
// that forwards to the parser package
func (s *CardStore) RenderMarkdown(content string) (string, error) {
	return parser.RenderMarkdown(content)
}
