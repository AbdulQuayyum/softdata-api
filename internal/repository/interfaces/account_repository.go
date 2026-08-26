package interfaces

import (
	"context"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

// AccountRepository defines account persistence operations.
type AccountRepository interface {
	Create(ctx context.Context, input models.AccountCreateInput, passwordHash string) (models.Account, error)
	GetByID(ctx context.Context, id string) (models.Account, error)
	GetByUsername(ctx context.Context, username string) (models.Account, error)
	GetByEmail(ctx context.Context, email string) (models.Account, error)
	UpdateProfile(ctx context.Context, id string, input models.AccountUpdateInput) (models.Account, error)
	UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (models.Account, error)
	MarkLogin(ctx context.Context, id string) (models.Account, error)
	VerifyEmail(ctx context.Context, id string) (models.Account, error)
	Deactivate(ctx context.Context, id string) (models.Account, error)
}
