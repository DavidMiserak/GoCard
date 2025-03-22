// Package views contains compatibility shims for working with the storage package.
package views

import (
	"github.com/DavidMiserak/GoCard/internal/storage"
)

// These functions are provided as compatibility shims during refactoring.
// They can be removed after updating UI code to use the new storage API.

// renderMarkdown is a helper function that forwards to the storage package
func renderMarkdown(store *storage.CardStore, content string) (string, error) {
	return store.RenderMarkdown(content)
}
