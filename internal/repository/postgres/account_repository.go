package postgres

import (
	"context"
	"fmt"

	sqlc "github.com/AbdulQuayyum/softdata-api/internal/database/sqlc"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

var _ interfaces.AccountRepository = (*AccountRepository)(nil)

type AccountRepository struct {
	queries *sqlc.Queries
}

func NewAccountRepository(dbt sqlc.DBTX) *AccountRepository {
	return &AccountRepository{queries: sqlc.New(dbt)}
}

func (r *AccountRepository) Create(ctx context.Context, input models.AccountCreateInput, passwordHash string) (models.Account, error) {
	params := sqlc.CreateAccountParams{
		Username:     input.Username,
		Email:        textFromStringPtr(input.Email),
		PasswordHash: passwordHash,
		Column4:      nil,
	}

	row, err := r.queries.CreateAccount(ctx, params)
	if err != nil {
		return models.Account{}, translateError("create account", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("get account by id: %w", err)
	}

	row, err := r.queries.GetAccountByID(ctx, uid)
	if err != nil {
		return models.Account{}, translateError("get account by id", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) GetByUsername(ctx context.Context, username string) (models.Account, error) {
	row, err := r.queries.GetAccountByUsername(ctx, username)
	if err != nil {
		return models.Account{}, translateError("get account by username", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (models.Account, error) {
	row, err := r.queries.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.Account{}, translateError("get account by email", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) UpdateProfile(ctx context.Context, id string, input models.AccountUpdateInput) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("update account profile: %w", err)
	}

	username := ""
	email := sqlc.UpdateAccountProfileParams{}.Email

	if input.Username == nil || input.Email == nil {
		current, err := r.queries.GetAccountByID(ctx, uid)
		if err != nil {
			return models.Account{}, translateError("get account by id", err)
		}
		username = current.Username
		email = current.Email
	}

	if input.Username != nil {
		username = *input.Username
	}
	if input.Email != nil {
		email = textFromStringPtr(input.Email)
	}

	row, err := r.queries.UpdateAccountProfile(ctx, sqlc.UpdateAccountProfileParams{
		ID:       uid,
		Username: username,
		Email:    email,
	})
	if err != nil {
		return models.Account{}, translateError("update account profile", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("update account password hash: %w", err)
	}

	row, err := r.queries.UpdateAccountPassword(ctx, sqlc.UpdateAccountPasswordParams{
		ID:           uid,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return models.Account{}, translateError("update account password hash", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) MarkLogin(ctx context.Context, id string) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("mark account login: %w", err)
	}

	row, err := r.queries.MarkAccountLogin(ctx, uid)
	if err != nil {
		return models.Account{}, translateError("mark account login", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) VerifyEmail(ctx context.Context, id string) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("verify account email: %w", err)
	}

	row, err := r.queries.VerifyAccountEmail(ctx, uid)
	if err != nil {
		return models.Account{}, translateError("verify account email", err)
	}
	return accountFromRow(row), nil
}

func (r *AccountRepository) Deactivate(ctx context.Context, id string) (models.Account, error) {
	uid, err := uuidFromString(id)
	if err != nil {
		return models.Account{}, fmt.Errorf("deactivate account: %w", err)
	}

	row, err := r.queries.SoftDeleteAccount(ctx, uid)
	if err != nil {
		return models.Account{}, translateError("deactivate account", err)
	}
	return accountFromRow(row), nil
}
