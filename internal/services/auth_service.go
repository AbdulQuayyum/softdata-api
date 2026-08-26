package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
)

type RefreshTokenGenerator interface {
	Generate() (plaintext, hash string, err error)
}

type AccessTokenIssuer interface {
	Issue(ctx context.Context, accountID, sessionID string, ttl time.Duration) (token, jti string, err error)
}

type Clock interface {
	Now() time.Time
}

type securityRefreshTokenGenerator struct{}

func NewSecurityRefreshTokenGenerator() RefreshTokenGenerator {
	return securityRefreshTokenGenerator{}
}

func (securityRefreshTokenGenerator) Generate() (string, string, error) {
	return security.GenerateRefreshToken()
}

type securityAccessTokenIssuer struct {
	secret string
}

func NewSecurityAccessTokenIssuer(secret string) (AccessTokenIssuer, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("access token secret is required")
	}
	return securityAccessTokenIssuer{secret: secret}, nil
}

func (i securityAccessTokenIssuer) Issue(ctx context.Context, accountID, sessionID string, ttl time.Duration) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}

	token, err := security.GenerateAccessToken(i.secret, accountID, sessionID, ttl)
	if err != nil {
		return "", "", err
	}

	claims, err := security.ValidateAccessToken(token, i.secret)
	if err != nil {
		return "", "", err
	}

	return token, claims.ID, nil
}

type systemClock struct{}

func NewSystemClock() Clock {
	return systemClock{}
}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type AuthService struct {
	accounts        interfaces.AccountRepository
	sessions        interfaces.SessionRepository
	hasher          PasswordHasher
	accessTokens    AccessTokenIssuer
	refreshTokens   RefreshTokenGenerator
	clock           Clock
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	accounts interfaces.AccountRepository,
	sessions interfaces.SessionRepository,
	hasher PasswordHasher,
	accessTokens AccessTokenIssuer,
	refreshTokens RefreshTokenGenerator,
	clock Clock,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) (*AuthService, error) {
	switch {
	case accounts == nil:
		return nil, fmt.Errorf("account repository is required")
	case sessions == nil:
		return nil, fmt.Errorf("session repository is required")
	case hasher == nil:
		return nil, fmt.Errorf("password hasher is required")
	case accessTokens == nil:
		return nil, fmt.Errorf("access token issuer is required")
	case refreshTokens == nil:
		return nil, fmt.Errorf("refresh token generator is required")
	case clock == nil:
		return nil, fmt.Errorf("clock is required")
	case accessTokenTTL <= 0:
		return nil, fmt.Errorf("access token ttl must be positive")
	case refreshTokenTTL <= 0:
		return nil, fmt.Errorf("refresh token ttl must be positive")
	}

	return &AuthService{
		accounts:        accounts,
		sessions:        sessions,
		hasher:          hasher,
		accessTokens:    accessTokens,
		refreshTokens:   refreshTokens,
		clock:           clock,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (models.LoginResult, error) {
	username = normalizeUsername(username)
	if username == "" || strings.TrimSpace(password) == "" {
		return models.LoginResult{}, ErrInvalidCredentials
	}

	account, err := s.accounts.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.LoginResult{}, ErrInvalidCredentials
		}
		return models.LoginResult{}, fmt.Errorf("get account by username: %w", err)
	}
	if account.Status != models.AccountStatusActive {
		return models.LoginResult{}, ErrInvalidCredentials
	}

	ok, err := s.hasher.Verify(password, account.PasswordHash)
	if err != nil {
		return models.LoginResult{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return models.LoginResult{}, ErrInvalidCredentials
	}

	now := s.clock.Now().UTC()
	refreshPlain, refreshHash, err := s.refreshTokens.Generate()
	if err != nil {
		return models.LoginResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	session, err := s.sessions.Create(ctx, models.Session{
		AccountID:        account.ID,
		RefreshTokenHash: refreshHash,
		AccessTokenJTI:   nil,
		UserAgent:        nil,
		IPAddress:        nil,
		ExpiresAt:        now.Add(s.refreshTokenTTL),
		RevokedAt:        nil,
		LastUsedAt:       &now,
	})
	if err != nil {
		return models.LoginResult{}, fmt.Errorf("create session: %w", err)
	}

	accessToken, accessTokenJTI, err := s.accessTokens.Issue(ctx, account.ID, session.ID, s.accessTokenTTL)
	if err != nil {
		_, _ = s.sessions.RevokeByID(ctx, session.ID)
		return models.LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}

	if _, err := s.sessions.RotateSessionTokens(ctx, session.ID, refreshHash, refreshHash, accessTokenJTI); err != nil {
		_, _ = s.sessions.RevokeByID(ctx, session.ID)
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.LoginResult{}, ErrInvalidCredentials
		}
		return models.LoginResult{}, fmt.Errorf("store session token jti: %w", err)
	}

	return models.LoginResult{
		Account: accountResponseFromModel(account),
		Tokens: models.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshPlain,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.accessTokenTTL / time.Second),
		},
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (models.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return models.TokenPair{}, ErrInvalidRefreshToken
	}

	currentHash := security.HashRefreshToken(refreshToken)
	session, err := s.sessions.GetByRefreshTokenHash(ctx, currentHash)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.TokenPair{}, ErrInvalidRefreshToken
		}
		return models.TokenPair{}, fmt.Errorf("get session by refresh token hash: %w", err)
	}

	now := s.clock.Now().UTC()
	if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
		return models.TokenPair{}, ErrInvalidRefreshToken
	}

	account, err := s.accounts.GetByID(ctx, session.AccountID)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.TokenPair{}, ErrInvalidRefreshToken
		}
		return models.TokenPair{}, fmt.Errorf("get account by id: %w", err)
	}
	if account.Status != models.AccountStatusActive {
		return models.TokenPair{}, ErrInvalidRefreshToken
	}

	newRefreshPlain, newRefreshHash, err := s.refreshTokens.Generate()
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	accessToken, accessTokenJTI, err := s.accessTokens.Issue(ctx, account.ID, session.ID, s.accessTokenTTL)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("issue access token: %w", err)
	}

	if _, err := s.sessions.RotateSessionTokens(ctx, session.ID, currentHash, newRefreshHash, accessTokenJTI); err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return models.TokenPair{}, ErrInvalidRefreshToken
		}
		return models.TokenPair{}, fmt.Errorf("rotate session tokens: %w", err)
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshPlain,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL / time.Second),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ErrInvalidRefreshToken
	}

	if _, err := s.sessions.RevokeByRefreshTokenHash(ctx, security.HashRefreshToken(refreshToken)); err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("revoke session by refresh token hash: %w", err)
	}

	return nil
}
