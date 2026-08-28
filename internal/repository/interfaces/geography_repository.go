package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// GeographyRepository defines state lookup operations backed by a geography dataset.
type GeographyRepository interface {
	ListStates(ctx context.Context) ([]models.State, error)
	GetStateByID(ctx context.Context, stateID string) (models.State, error)
}
