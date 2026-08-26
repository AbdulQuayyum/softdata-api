package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type accountRepoStub struct {
	createFn            func(context.Context, models.AccountCreateInput, string) (models.Account, error)
	getByIDFn           func(context.Context, string) (models.Account, error)
	getByUsernameFn     func(context.Context, string) (models.Account, error)
	getByEmailFn        func(context.Context, string) (models.Account, error)
	updateProfileFn     func(context.Context, string, models.AccountUpdateInput) (models.Account, error)
	updatePasswordFn    func(context.Context, string, string) (models.Account, error)
	deactivateFn        func(context.Context, string) (models.Account, error)
	markLoginFn         func(context.Context, string) (models.Account, error)
	verifyEmailFn       func(context.Context, string) (models.Account, error)
	createCalls         int
	getByIDCalls        int
	getByUsernameCalls  int
	getByEmailCalls     int
	updateProfileCalls  int
	updatePasswordCalls int
	deactivateCalls     int
	markLoginCalls      int
	verifyEmailCalls    int
	lastCreateInput     models.AccountCreateInput
	lastCreateHash      string
	lastUpdateProfileID string
	lastUpdatePassword  string
}

func (s *accountRepoStub) Create(ctx context.Context, input models.AccountCreateInput, passwordHash string) (models.Account, error) {
	s.createCalls++
	s.lastCreateInput = input
	s.lastCreateHash = passwordHash
	if s.createFn != nil {
		return s.createFn(ctx, input, passwordHash)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) GetByID(ctx context.Context, id string) (models.Account, error) {
	s.getByIDCalls++
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) GetByUsername(ctx context.Context, username string) (models.Account, error) {
	s.getByUsernameCalls++
	if s.getByUsernameFn != nil {
		return s.getByUsernameFn(ctx, username)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) GetByEmail(ctx context.Context, email string) (models.Account, error) {
	s.getByEmailCalls++
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, email)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) UpdateProfile(ctx context.Context, id string, input models.AccountUpdateInput) (models.Account, error) {
	s.updateProfileCalls++
	s.lastUpdateProfileID = id
	if s.updateProfileFn != nil {
		return s.updateProfileFn(ctx, id, input)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (models.Account, error) {
	s.updatePasswordCalls++
	s.lastUpdatePassword = passwordHash
	if s.updatePasswordFn != nil {
		return s.updatePasswordFn(ctx, id, passwordHash)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) MarkLogin(ctx context.Context, id string) (models.Account, error) {
	s.markLoginCalls++
	if s.markLoginFn != nil {
		return s.markLoginFn(ctx, id)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) VerifyEmail(ctx context.Context, id string) (models.Account, error) {
	s.verifyEmailCalls++
	if s.verifyEmailFn != nil {
		return s.verifyEmailFn(ctx, id)
	}
	return models.Account{}, interfaces.ErrNotFound
}

func (s *accountRepoStub) Deactivate(ctx context.Context, id string) (models.Account, error) {
	s.deactivateCalls++
	if s.deactivateFn != nil {
		return s.deactivateFn(ctx, id)
	}
	return models.Account{}, interfaces.ErrNotFound
}

type passwordHasherStub struct {
	hashFn   func(string) (string, error)
	verifyFn func(string, string) (bool, error)
}

func (h passwordHasherStub) Hash(password string) (string, error) {
	if h.hashFn != nil {
		return h.hashFn(password)
	}
	return "", nil
}

func (h passwordHasherStub) Verify(password, encoded string) (bool, error) {
	if h.verifyFn != nil {
		return h.verifyFn(password, encoded)
	}
	return false, nil
}

func TestAccountServiceRegisterReturnsSafeAccount(t *testing.T) {
	repo := &accountRepoStub{
		getByUsernameFn: func(context.Context, string) (models.Account, error) {
			return models.Account{}, interfaces.ErrNotFound
		},
		getByEmailFn: func(context.Context, string) (models.Account, error) {
			return models.Account{}, interfaces.ErrNotFound
		},
		createFn: func(_ context.Context, input models.AccountCreateInput, passwordHash string) (models.Account, error) {
			if input.Username != "alice" {
				t.Fatalf("username was not normalized: got %q", input.Username)
			}
			if input.Email == nil || *input.Email != "alice@example.com" {
				t.Fatalf("email was not normalized: got %#v", input.Email)
			}
			if passwordHash == "plaintext-password" {
				t.Fatal("plaintext password reached repository")
			}

			email := "alice@example.com"
			now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
			return models.Account{
				ID:              "acct-1",
				Username:        input.Username,
				Email:           &email,
				PasswordHash:    passwordHash,
				Status:          models.AccountStatusActive,
				EmailVerifiedAt: &now,
				LastLoginAt:     &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil
		},
	}
	svc, err := NewAccountService(repo, passwordHasherStub{
		hashFn: func(password string) (string, error) {
			if password != "plaintext-password" {
				t.Fatalf("unexpected password to hash: %q", password)
			}
			return "hashed-password", nil
		},
	})
	if err != nil {
		t.Fatalf("NewAccountService() error = %v", err)
	}

	email := " Alice@Example.com "
	result, err := svc.Register(context.Background(), models.AccountCreateInput{
		Username: " Alice ",
		Email:    &email,
		Password: "plaintext-password",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	marshaled, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(marshaled), "password_hash") {
		t.Fatalf("safe account leaked password hash: %s", marshaled)
	}
	if result.Username != "alice" {
		t.Fatalf("unexpected username: %q", result.Username)
	}
	if result.Email == nil || *result.Email != "alice@example.com" {
		t.Fatalf("unexpected email: %#v", result.Email)
	}
	if repo.getByEmailCalls != 1 {
		t.Fatalf("expected one email availability check, got %d", repo.getByEmailCalls)
	}
	if repo.lastCreateHash != "hashed-password" {
		t.Fatalf("unexpected stored password hash: %q", repo.lastCreateHash)
	}
}

func TestAccountServiceRegisterDuplicateUsername(t *testing.T) {
	repo := &accountRepoStub{
		getByUsernameFn: func(context.Context, string) (models.Account, error) {
			return models.Account{ID: "acct-1"}, nil
		},
	}
	svc, err := NewAccountService(repo, passwordHasherStub{
		hashFn: func(string) (string, error) { return "hashed", nil },
	})
	if err != nil {
		t.Fatalf("NewAccountService() error = %v", err)
	}

	_, err = svc.Register(context.Background(), models.AccountCreateInput{
		Username: "alice",
		Password: "password",
	})
	if !errors.Is(err, ErrUsernameUnavailable) {
		t.Fatalf("Register() error = %v, want ErrUsernameUnavailable", err)
	}
}

func TestAccountServiceRegisterDuplicateEmail(t *testing.T) {
	repo := &accountRepoStub{
		getByUsernameFn: func(context.Context, string) (models.Account, error) {
			return models.Account{}, interfaces.ErrNotFound
		},
		getByEmailFn: func(context.Context, string) (models.Account, error) {
			return models.Account{ID: "acct-1"}, nil
		},
	}
	svc, err := NewAccountService(repo, passwordHasherStub{
		hashFn: func(string) (string, error) { return "hashed", nil },
	})
	if err != nil {
		t.Fatalf("NewAccountService() error = %v", err)
	}

	email := "alice@example.com"
	_, err = svc.Register(context.Background(), models.AccountCreateInput{
		Username: "alice",
		Email:    &email,
		Password: "password",
	})
	if !errors.Is(err, ErrEmailUnavailable) {
		t.Fatalf("Register() error = %v, want ErrEmailUnavailable", err)
	}
}

func TestAccountServiceChangePassword(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repo := &accountRepoStub{
		getByIDFn: func(context.Context, string) (models.Account, error) {
			return models.Account{
				ID:           "acct-1",
				Username:     "alice",
				Status:       models.AccountStatusActive,
				PasswordHash: "stored-hash",
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
		updatePasswordFn: func(_ context.Context, id, passwordHash string) (models.Account, error) {
			if id != "acct-1" {
				t.Fatalf("unexpected account ID: %q", id)
			}
			if passwordHash == "new-password" {
				t.Fatal("plaintext password reached repository")
			}
			return models.Account{
				ID:           "acct-1",
				Username:     "alice",
				Status:       models.AccountStatusActive,
				PasswordHash: passwordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}
	svc, err := NewAccountService(repo, passwordHasherStub{
		verifyFn: func(password, encoded string) (bool, error) {
			if password != "current-password" || encoded != "stored-hash" {
				t.Fatalf("unexpected verify inputs: %q %q", password, encoded)
			}
			return true, nil
		},
		hashFn: func(password string) (string, error) {
			if password != "new-password" {
				t.Fatalf("unexpected new password: %q", password)
			}
			return "new-hash", nil
		},
	})
	if err != nil {
		t.Fatalf("NewAccountService() error = %v", err)
	}

	result, err := svc.ChangePassword(context.Background(), " acct-1 ", "current-password", "new-password")
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	marshaled, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(marshaled), "password_hash") {
		t.Fatalf("safe account leaked password hash: %s", marshaled)
	}
	if repo.lastUpdatePassword != "new-hash" {
		t.Fatalf("unexpected stored password hash: %q", repo.lastUpdatePassword)
	}
}

func TestAccountServiceChangePasswordInvalidCurrentPassword(t *testing.T) {
	repo := &accountRepoStub{
		getByIDFn: func(context.Context, string) (models.Account, error) {
			return models.Account{
				ID:           "acct-1",
				Username:     "alice",
				Status:       models.AccountStatusActive,
				PasswordHash: "stored-hash",
			}, nil
		},
	}
	svc, err := NewAccountService(repo, passwordHasherStub{
		verifyFn: func(string, string) (bool, error) {
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("NewAccountService() error = %v", err)
	}

	_, err = svc.ChangePassword(context.Background(), "acct-1", "wrong", "new-password")
	if !errors.Is(err, ErrCurrentPasswordInvalid) {
		t.Fatalf("ChangePassword() error = %v, want ErrCurrentPasswordInvalid", err)
	}
}
