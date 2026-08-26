package security

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenIssuer = "softdata-api"

type AccessTokenClaims struct {
	TokenType string `json:"token_type"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(secret, accountID, sessionID string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("access token secret is required")
	}
	if strings.TrimSpace(accountID) == "" {
		return "", errors.New("account id is required")
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("session id is required")
	}
	if ttl <= 0 {
		return "", errors.New("access token ttl must be positive")
	}

	now := time.Now().UTC()
	claims := AccessTokenClaims{
		TokenType: "access",
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    accessTokenIssuer,
			Subject:   accountID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	jti, err := newTokenID()
	if err != nil {
		return "", fmt.Errorf("generate access token id: %w", err)
	}
	claims.ID = jti

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

func ValidateAccessToken(tokenString, secret string) (*AccessTokenClaims, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("access token secret is required")
	}
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("access token is required")
	}

	claims := &AccessTokenClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(accessTokenIssuer))
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected access token algorithm")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate access token: %w", err)
	}
	if token == nil || !token.Valid {
		return nil, errors.New("access token is invalid")
	}

	now := time.Now().UTC()
	if claims.IssuedAt == nil {
		return nil, errors.New("access token issued-at is required")
	}
	if claims.NotBefore == nil {
		return nil, errors.New("access token not-before is required")
	}
	if now.Before(claims.NotBefore.Time) {
		return nil, errors.New("access token not-before is in the future")
	}
	if claims.ExpiresAt == nil {
		return nil, errors.New("access token expiry is required")
	}
	if now.After(claims.ExpiresAt.Time) {
		return nil, errors.New("access token is expired")
	}
	if claims.TokenType != "access" {
		return nil, errors.New("access token type is invalid")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return nil, errors.New("access token subject is required")
	}
	if strings.TrimSpace(claims.SessionID) == "" {
		return nil, errors.New("access token session id is required")
	}
	if strings.TrimSpace(claims.ID) == "" {
		return nil, errors.New("access token jti is required")
	}

	return claims, nil
}
