// internal/service/interfaces/card_service.go
package interfaces

import (
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
)

// CardService manages operations on individual cards
type CardService interface {
	// Card read operations (no creation/editing in-app)
	GetCard(cardPath string) (domain.Card, error)

	// Review operations
	ReviewCard(cardPath string, rating int) error
	IsDue(cardPath string) bool
	GetDueDate(cardPath string) time.Time
}
