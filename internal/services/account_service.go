package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) (bool, error)
}

type securityPasswordHasher struct{}

func NewSecurityPasswordHasher() PasswordHasher {
	return securityPasswordHasher{}
}

func (securityPasswordHasher) Hash(password string) (string, error) {
	return security.HashPassword(password)
}

func (securityPasswordHasher) Verify(password, encoded string) (bool, error) {
	return security.VerifyPassword(password, encoded)
}

type AccountService struct {
	accounts interfaces.AccountRepository
	hasher   PasswordHasher
}

func NewAccountService(accounts interfaces.AccountRepository, hasher PasswordHasher) (*AccountService, error) {
	if accounts == nil {
		return nil, fmt.Errorf("account repository is required")
	}
	if hasher == nil {
		return nil, fmt.Errorf("password hasher is required")
	}

	return &AccountService{accounts: accounts, hasher: hasher}, nil
}

func (s *AccountService) Register(ctx context.Context, input models.AccountCreateInput) (models.AccountResponse, error) {
	username := normalizeUsername(input.Username)
	if username == "" || strings.TrimSpace(input.Password) == "" {
		return models.AccountResponse{}, fmt.Errorf("register account: username and password are required")
	}

	email := normalizeOptionalEmail(input.Email)

	if _, err := s.accounts.GetByUsername(ctx, username); err != nil {
		if !errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, fmt.Errorf("lookup username availability: %w", err)
		}
	} else {
		return models.AccountResponse{}, ErrUsernameUnavailable
	}

	if email != nil {
		if existing, err := s.accounts.GetByEmail(ctx, *email); err != nil {
			if !errors.Is(err, interfaces.ErrNotFound) {
				return models.AccountResponse{}, fmt.Errorf("lookup email availability: %w", err)
			}
		} else if existing.ID != "" {
			return models.AccountResponse{}, ErrEmailUnavailable
		}
	}

	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return models.AccountResponse{}, fmt.Errorf("hash password: %w", err)
	}

	created, err := s.accounts.Create(ctx, models.AccountCreateInput{
		Username: username,
		Email:    email,
	}, passwordHash)
	if err != nil {
		if errors.Is(err, interfaces.ErrConflict) {
			return models.AccountResponse{}, ErrUsernameUnavailable
		}
		return models.AccountResponse{}, fmt.Errorf("create account: %w", err)
	}

	return accountResponseFromModel(created), nil
}

func (s *AccountService) GetCurrent(ctx context.Context, accountID string) (models.AccountResponse, error) {
	account, err := s.accounts.GetByID(ctx, normalizeAccountID(accountID))
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, ErrAccountNotFound
		}
		return models.AccountResponse{}, fmt.Errorf("get current account: %w", err)
	}

	return accountResponseFromModel(account), nil
}

func (s *AccountService) UpdateCurrent(ctx context.Context, accountID string, input models.AccountUpdateInput) (models.AccountResponse, error) {
	current, err := s.accounts.GetByID(ctx, normalizeAccountID(accountID))
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, ErrAccountNotFound
		}
		return models.AccountResponse{}, fmt.Errorf("get current account: %w", err)
	}

	normalized := models.AccountUpdateInput{}
	if input.Username != nil {
		username := normalizeUsername(*input.Username)
		if username == "" {
			return models.AccountResponse{}, fmt.Errorf("update account: username cannot be empty")
		}
		if username != current.Username {
			existing, err := s.accounts.GetByUsername(ctx, username)
			if err != nil {
				if !errors.Is(err, interfaces.ErrNotFound) {
					return models.AccountResponse{}, fmt.Errorf("lookup username availability: %w", err)
				}
			} else if existing.ID != current.ID {
				return models.AccountResponse{}, ErrUsernameUnavailable
			}
		}
		normalized.Username = &username
	}

	if input.Email != nil {
		email := normalizeEmailValue(*input.Email)
		if email != nil && current.Email != nil && *email == strings.ToLower(strings.TrimSpace(*current.Email)) {
			normalized.Email = email
		} else if email != nil {
			existing, err := s.accounts.GetByEmail(ctx, *email)
			if err != nil {
				if !errors.Is(err, interfaces.ErrNotFound) {
					return models.AccountResponse{}, fmt.Errorf("lookup email availability: %w", err)
				}
			} else if existing.ID != current.ID {
				return models.AccountResponse{}, ErrEmailUnavailable
			}
			normalized.Email = email
		} else {
			normalized.Email = nil
		}
	}

	updated, err := s.accounts.UpdateProfile(ctx, current.ID, normalized)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, ErrAccountNotFound
		}
		if errors.Is(err, interfaces.ErrConflict) {
			return models.AccountResponse{}, ErrUsernameUnavailable
		}
		return models.AccountResponse{}, fmt.Errorf("update account: %w", err)
	}

	return accountResponseFromModel(updated), nil
}

func (s *AccountService) ChangePassword(ctx context.Context, accountID, currentPassword, newPassword string) (models.AccountResponse, error) {
	current, err := s.accounts.GetByID(ctx, normalizeAccountID(accountID))
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, ErrAccountNotFound
		}
		return models.AccountResponse{}, fmt.Errorf("get current account: %w", err)
	}

	if current.Status != models.AccountStatusActive {
		return models.AccountResponse{}, ErrAccountInactive
	}
	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
		return models.AccountResponse{}, fmt.Errorf("change password: current and new passwords are required")
	}

	ok, err := s.hasher.Verify(currentPassword, current.PasswordHash)
	if err != nil {
		return models.AccountResponse{}, fmt.Errorf("verify current password: %w", err)
	}
	if !ok {
		return models.AccountResponse{}, ErrCurrentPasswordInvalid
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return models.AccountResponse{}, fmt.Errorf("hash password: %w", err)
	}

	updated, err := s.accounts.UpdatePasswordHash(ctx, current.ID, passwordHash)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.AccountResponse{}, ErrAccountNotFound
		}
		if errors.Is(err, interfaces.ErrConflict) {
			return models.AccountResponse{}, ErrCurrentPasswordInvalid
		}
		return models.AccountResponse{}, fmt.Errorf("update password: %w", err)
	}

	return accountResponseFromModel(updated), nil
}

func (s *AccountService) DeactivateCurrent(ctx context.Context, accountID string) error {
	if _, err := s.accounts.Deactivate(ctx, normalizeAccountID(accountID)); err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return ErrAccountNotFound
		}
		return fmt.Errorf("deactivate account: %w", err)
	}
	return nil
}

func accountResponseFromModel(account models.Account) models.AccountResponse {
	return models.AccountResponse{
		ID:              account.ID,
		Username:        account.Username,
		Email:           account.Email,
		Status:          account.Status,
		EmailVerifiedAt: account.EmailVerifiedAt,
		LastLoginAt:     account.LastLoginAt,
		CreatedAt:       account.CreatedAt,
		UpdatedAt:       account.UpdatedAt,
	}
}

func normalizeAccountID(value string) string {
	return strings.TrimSpace(value)
}

func normalizeUsername(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeEmailValue(value string) *string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizeOptionalEmail(value *string) *string {
	if value == nil {
		return nil
	}
	return normalizeEmailValue(*value)
}
