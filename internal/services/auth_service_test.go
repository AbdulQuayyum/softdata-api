package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
)

type refreshTokenGeneratorStub struct {
	pairs []struct {
		plaintext string
		hash      string
	}
	calls int
}

func (g *refreshTokenGeneratorStub) Generate() (string, string, error) {
	if g.calls >= len(g.pairs) {
		return "", "", nil
	}
	pair := g.pairs[g.calls]
	g.calls++
	return pair.plaintext, pair.hash, nil
}

type accessTokenIssuerStub struct {
	token         string
	jti           string
	calls         int
	lastTTL       time.Duration
	lastAccountID string
	lastSessionID string
}

func (i *accessTokenIssuerStub) Issue(_ context.Context, accountID, sessionID string, ttl time.Duration) (string, string, error) {
	i.calls++
	i.lastTTL = ttl
	i.lastAccountID = accountID
	i.lastSessionID = sessionID
	return i.token, i.jti, nil
}

type clockStub struct {
	now time.Time
}

func (c clockStub) Now() time.Time {
	return c.now
}

type sessionRepoStub struct {
	createCalls          int
	getByHashCalls       int
	rotateCalls          int
	revokeByIDCalls      int
	revokeByHashCalls    int
	deleteExpiredCalls   int
	deleteByAccountCalls int
	lastCreate           models.Session
	lastGetHash          string
	lastRotate           struct {
		sessionID   string
		currentHash string
		newHash     string
		newJTI      string
	}
	lastRevokeHash string
	lastRevokeID   string
	sessionsByID   map[string]models.Session
	sessionsByHash map[string]string
	nextID         int
}

func newSessionRepoStub() *sessionRepoStub {
	return &sessionRepoStub{
		sessionsByID:   map[string]models.Session{},
		sessionsByHash: map[string]string{},
	}
}

func (s *sessionRepoStub) Create(_ context.Context, session models.Session) (models.Session, error) {
	s.createCalls++
	s.lastCreate = session
	s.nextID++
	if session.ID == "" {
		session.ID = "session-" + time.Now().UTC().Format("150405") + "-" + time.Now().UTC().Format("000000")
	}
	s.sessionsByID[session.ID] = session
	s.sessionsByHash[session.RefreshTokenHash] = session.ID
	return session, nil
}

func (s *sessionRepoStub) GetByID(_ context.Context, id string) (models.Session, error) {
	if session, ok := s.sessionsByID[id]; ok {
		return session, nil
	}
	return models.Session{}, interfaces.ErrNotFound
}

func (s *sessionRepoStub) GetByRefreshTokenHash(_ context.Context, refreshTokenHash string) (models.Session, error) {
	s.getByHashCalls++
	s.lastGetHash = refreshTokenHash
	id, ok := s.sessionsByHash[refreshTokenHash]
	if !ok {
		return models.Session{}, interfaces.ErrNotFound
	}
	return s.sessionsByID[id], nil
}

func (s *sessionRepoStub) ListByAccountID(_ context.Context, accountID string, limit, offset int32) ([]models.Session, error) {
	return nil, nil
}

func (s *sessionRepoStub) Touch(_ context.Context, id string) (models.Session, error) {
	if session, ok := s.sessionsByID[id]; ok {
		return session, nil
	}
	return models.Session{}, interfaces.ErrNotFound
}

func (s *sessionRepoStub) RotateSessionTokens(_ context.Context, sessionID string, currentRefreshTokenHash string, newRefreshTokenHash string, newAccessTokenJTI string) (models.Session, error) {
	s.rotateCalls++
	s.lastRotate.sessionID = sessionID
	s.lastRotate.currentHash = currentRefreshTokenHash
	s.lastRotate.newHash = newRefreshTokenHash
	s.lastRotate.newJTI = newAccessTokenJTI

	session, ok := s.sessionsByID[sessionID]
	if !ok {
		return models.Session{}, interfaces.ErrNotFound
	}
	if session.RefreshTokenHash != currentRefreshTokenHash {
		return models.Session{}, interfaces.ErrNotFound
	}

	delete(s.sessionsByHash, session.RefreshTokenHash)
	session.RefreshTokenHash = newRefreshTokenHash
	jti := newAccessTokenJTI
	session.AccessTokenJTI = &jti
	now := time.Now().UTC()
	session.LastUsedAt = &now
	s.sessionsByID[sessionID] = session
	s.sessionsByHash[newRefreshTokenHash] = sessionID
	return session, nil
}

func (s *sessionRepoStub) RevokeByID(_ context.Context, id string) (models.Session, error) {
	s.revokeByIDCalls++
	s.lastRevokeID = id
	session, ok := s.sessionsByID[id]
	if !ok {
		return models.Session{}, interfaces.ErrNotFound
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	s.sessionsByID[id] = session
	return session, nil
}

func (s *sessionRepoStub) RevokeByRefreshTokenHash(_ context.Context, refreshTokenHash string) (models.Session, error) {
	s.revokeByHashCalls++
	s.lastRevokeHash = refreshTokenHash
	id, ok := s.sessionsByHash[refreshTokenHash]
	if !ok {
		return models.Session{}, interfaces.ErrNotFound
	}
	session := s.sessionsByID[id]
	now := time.Now().UTC()
	session.RevokedAt = &now
	s.sessionsByID[id] = session
	return session, nil
}

func (s *sessionRepoStub) DeleteExpired(_ context.Context) error {
	s.deleteExpiredCalls++
	return nil
}

func (s *sessionRepoStub) DeleteByAccountID(_ context.Context, accountID string) error {
	s.deleteByAccountCalls++
	return nil
}

func TestAuthServiceLoginUsesUsernameOnly(t *testing.T) {
	accountRepo := &accountRepoStub{
		getByUsernameFn: func(_ context.Context, username string) (models.Account, error) {
			if username != "alice" {
				t.Fatalf("unexpected username lookup: %q", username)
			}
			return models.Account{
				ID:           "acct-1",
				Username:     "alice",
				PasswordHash: "stored-hash",
				Status:       models.AccountStatusActive,
			}, nil
		},
		getByEmailFn: func(context.Context, string) (models.Account, error) {
			t.Fatal("login must not call email lookup")
			return models.Account{}, nil
		},
	}
	sessionRepo := newSessionRepoStub()
	refreshGen := &refreshTokenGeneratorStub{pairs: []struct {
		plaintext string
		hash      string
	}{{plaintext: "refresh-plain", hash: "refresh-hash"}}}
	issuer := &accessTokenIssuerStub{token: "access-token", jti: "jti-1"}

	svc, err := NewAuthService(
		accountRepo,
		sessionRepo,
		passwordHasherStub{
			verifyFn: func(password, encoded string) (bool, error) {
				if password != "password" || encoded != "stored-hash" {
					t.Fatalf("unexpected verify inputs: %q %q", password, encoded)
				}
				return true, nil
			},
		},
		issuer,
		refreshGen,
		clockStub{now: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)},
		15*time.Minute,
		30*time.Minute,
	)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	result, err := svc.Login(context.Background(), " Alice ", "password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Account.ID != "acct-1" || result.Account.Username != "alice" {
		t.Fatalf("unexpected account in login result: %#v", result.Account)
	}
	if result.Tokens.AccessToken != "access-token" || result.Tokens.RefreshToken != "refresh-plain" {
		t.Fatalf("unexpected tokens in login result: %#v", result.Tokens)
	}
	if result.Tokens.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %q", result.Tokens.TokenType)
	}
	if result.Tokens.ExpiresIn != int64((15*time.Minute)/time.Second) {
		t.Fatalf("unexpected expires_in: %d", result.Tokens.ExpiresIn)
	}
	if accountRepo.getByUsernameCalls != 1 || accountRepo.getByEmailCalls != 0 {
		t.Fatalf("unexpected account lookup counts: username=%d email=%d", accountRepo.getByUsernameCalls, accountRepo.getByEmailCalls)
	}
	if sessionRepo.createCalls != 1 || sessionRepo.rotateCalls != 1 {
		t.Fatalf("unexpected session calls: create=%d rotate=%d", sessionRepo.createCalls, sessionRepo.rotateCalls)
	}
	if sessionRepo.lastCreate.RefreshTokenHash != "refresh-hash" {
		t.Fatalf("refresh token hash was not stored: %q", sessionRepo.lastCreate.RefreshTokenHash)
	}
	if sessionRepo.lastRotate.currentHash != "refresh-hash" || sessionRepo.lastRotate.newHash != "refresh-hash" {
		t.Fatalf("unexpected login rotation args: %#v", sessionRepo.lastRotate)
	}
	if issuer.lastTTL != 15*time.Minute {
		t.Fatalf("unexpected access token ttl: %s", issuer.lastTTL)
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	t.Run("unknown username", func(t *testing.T) {
		svc, err := NewAuthService(
			&accountRepoStub{
				getByUsernameFn: func(context.Context, string) (models.Account, error) {
					return models.Account{}, interfaces.ErrNotFound
				},
			},
			newSessionRepoStub(),
			passwordHasherStub{},
			&accessTokenIssuerStub{},
			&refreshTokenGeneratorStub{},
			clockStub{now: time.Now().UTC()},
			time.Minute,
			time.Minute,
		)
		if err != nil {
			t.Fatalf("NewAuthService() error = %v", err)
		}

		_, err = svc.Login(context.Background(), "alice", "password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		svc, err := NewAuthService(
			&accountRepoStub{
				getByUsernameFn: func(context.Context, string) (models.Account, error) {
					return models.Account{
						ID:           "acct-1",
						Username:     "alice",
						PasswordHash: "stored-hash",
						Status:       models.AccountStatusActive,
					}, nil
				},
			},
			newSessionRepoStub(),
			passwordHasherStub{
				verifyFn: func(string, string) (bool, error) {
					return false, nil
				},
			},
			&accessTokenIssuerStub{},
			&refreshTokenGeneratorStub{},
			clockStub{now: time.Now().UTC()},
			time.Minute,
			time.Minute,
		)
		if err != nil {
			t.Fatalf("NewAuthService() error = %v", err)
		}

		_, err = svc.Login(context.Background(), "alice", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
		}
	})
}

func TestAuthServiceRefreshRotatesTokens(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	sessionRepo := newSessionRepoStub()
	accountRepo := &accountRepoStub{
		getByIDFn: func(_ context.Context, id string) (models.Account, error) {
			if id != "acct-1" {
				t.Fatalf("unexpected account id lookup: %q", id)
			}
			return models.Account{
				ID:           "acct-1",
				Username:     "alice",
				PasswordHash: "stored-hash",
				Status:       models.AccountStatusActive,
			}, nil
		},
	}
	session, err := sessionRepo.Create(context.Background(), models.Session{
		ID:               "session-1",
		AccountID:        "acct-1",
		RefreshTokenHash: security.HashRefreshToken("refresh-plain"),
		ExpiresAt:        now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("seed session create error = %v", err)
	}
	sessionRepo.sessionsByID[session.ID] = session

	refreshGen := &refreshTokenGeneratorStub{pairs: []struct {
		plaintext string
		hash      string
	}{
		{plaintext: "refresh-next", hash: "refresh-next-hash"},
	}}
	issuer := &accessTokenIssuerStub{token: "access-next", jti: "jti-next"}
	svc, err := NewAuthService(
		accountRepo,
		sessionRepo,
		passwordHasherStub{},
		issuer,
		refreshGen,
		clockStub{now: now},
		30*time.Minute,
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	tokens, err := svc.Refresh(context.Background(), "refresh-plain")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if tokens.AccessToken != "access-next" || tokens.RefreshToken != "refresh-next" {
		t.Fatalf("unexpected refresh result: %#v", tokens)
	}
	if tokens.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %q", tokens.TokenType)
	}
	if tokens.ExpiresIn != int64((30*time.Minute)/time.Second) {
		t.Fatalf("unexpected expires_in: %d", tokens.ExpiresIn)
	}
	if sessionRepo.lastGetHash != security.HashRefreshToken("refresh-plain") {
		t.Fatalf("plaintext refresh token was used in lookup: %q", sessionRepo.lastGetHash)
	}
	if sessionRepo.rotateCalls != 1 {
		t.Fatalf("expected one rotation call, got %d", sessionRepo.rotateCalls)
	}
	if sessionRepo.lastRotate.currentHash != security.HashRefreshToken("refresh-plain") {
		t.Fatalf("unexpected current hash: %q", sessionRepo.lastRotate.currentHash)
	}

	_, err = svc.Refresh(context.Background(), "refresh-plain")
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() reuse error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthServiceRefreshRejectsRevokedSession(t *testing.T) {
	sessionRepo := newSessionRepoStub()
	session, err := sessionRepo.Create(context.Background(), models.Session{
		ID:               "session-1",
		AccountID:        "acct-1",
		RefreshTokenHash: security.HashRefreshToken("refresh-plain"),
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("seed session create error = %v", err)
	}
	session.RevokedAt = ptrTime(time.Now().UTC())
	sessionRepo.sessionsByID[session.ID] = session

	svc, err := NewAuthService(
		&accountRepoStub{
			getByIDFn: func(context.Context, string) (models.Account, error) {
				return models.Account{ID: "acct-1", Status: models.AccountStatusActive}, nil
			},
		},
		sessionRepo,
		passwordHasherStub{},
		&accessTokenIssuerStub{},
		&refreshTokenGeneratorStub{},
		clockStub{now: time.Now().UTC()},
		time.Minute,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	_, err = svc.Refresh(context.Background(), "refresh-plain")
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthServiceLogoutHashesRefreshToken(t *testing.T) {
	sessionRepo := newSessionRepoStub()
	_, err := sessionRepo.Create(context.Background(), models.Session{
		ID:               "session-1",
		AccountID:        "acct-1",
		RefreshTokenHash: security.HashRefreshToken("refresh-plain"),
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("seed session create error = %v", err)
	}

	svc, err := NewAuthService(
		&accountRepoStub{},
		sessionRepo,
		passwordHasherStub{},
		&accessTokenIssuerStub{},
		&refreshTokenGeneratorStub{},
		clockStub{now: time.Now().UTC()},
		time.Minute,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	if err := svc.Logout(context.Background(), "refresh-plain"); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if sessionRepo.lastRevokeHash != security.HashRefreshToken("refresh-plain") {
		t.Fatalf("plaintext refresh token was passed to repository: %q", sessionRepo.lastRevokeHash)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
