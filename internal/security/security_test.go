package security

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashPassword(t *testing.T) {
	hashOne, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	hashTwo, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hashOne == hashTwo {
		t.Fatal("expected unique salts to produce different hashes")
	}

	ok, err := VerifyPassword("correct horse battery staple", hashOne)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Fatal("expected password verification to succeed")
	}

	ok, err = VerifyPassword("incorrect", hashOne)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if ok {
		t.Fatal("expected password verification to fail")
	}
}

func TestVerifyPasswordRejectsMalformedOrOversizedHashes(t *testing.T) {
	t.Run("malformed", func(t *testing.T) {
		ok, err := VerifyPassword("password", "not-a-password-hash")
		if err == nil {
			t.Fatal("expected error")
		}
		if ok {
			t.Fatal("expected verification failure")
		}
	})

	t.Run("oversized parameters", func(t *testing.T) {
		salt := base64.RawStdEncoding.EncodeToString(make([]byte, passwordSaltLength))
		hash := base64.RawStdEncoding.EncodeToString(make([]byte, passwordKeyLength))
		encoded := "$argon2id$v=19$m=999999,t=3,p=2$" + salt + "$" + hash
		ok, err := VerifyPassword("password", encoded)
		if err == nil {
			t.Fatal("expected error")
		}
		if ok {
			t.Fatal("expected verification failure")
		}
	})
}

func TestAccessTokenLifecycle(t *testing.T) {
	secret := strings.Repeat("s", 32)

	token, err := GenerateAccessToken(secret, "account-123", "session-456", time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := ValidateAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.Subject != "account-123" {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}
	if claims.SessionID != "session-456" {
		t.Fatalf("unexpected session id: %s", claims.SessionID)
	}
	if claims.TokenType != "access" {
		t.Fatalf("unexpected token type: %s", claims.TokenType)
	}
	if claims.Issuer != accessTokenIssuer {
		t.Fatalf("unexpected issuer: %s", claims.Issuer)
	}

	t.Run("expired", func(t *testing.T) {
		expired := mustSignAccessToken(t, secret, AccessTokenClaims{
			TokenType: "access",
			SessionID: "session-456",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    accessTokenIssuer,
				Subject:   "account-123",
				ID:        "expired-token",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Minute)),
			},
		}, jwt.SigningMethodHS256)
		if _, err := ValidateAccessToken(expired, secret); err == nil {
			t.Fatal("expected expired token error")
		}
	})

	t.Run("altered signature", func(t *testing.T) {
		altered := token[:len(token)-1] + flipTokenChar(token[len(token)-1])
		if _, err := ValidateAccessToken(altered, secret); err == nil {
			t.Fatal("expected signature error")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		if _, err := ValidateAccessToken(token, strings.Repeat("x", 32)); err == nil {
			t.Fatal("expected secret mismatch error")
		}
	})

	t.Run("wrong algorithm", func(t *testing.T) {
		wrongAlg := mustSignAccessToken(t, secret, AccessTokenClaims{
			TokenType: "access",
			SessionID: "session-456",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    accessTokenIssuer,
				Subject:   "account-123",
				ID:        "wrong-alg",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
			},
		}, jwt.SigningMethodHS512)
		if _, err := ValidateAccessToken(wrongAlg, secret); err == nil {
			t.Fatal("expected algorithm error")
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		wrongIssuer := mustSignAccessToken(t, secret, AccessTokenClaims{
			TokenType: "access",
			SessionID: "session-456",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "other-service",
				Subject:   "account-123",
				ID:        "wrong-issuer",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
			},
		}, jwt.SigningMethodHS256)
		if _, err := ValidateAccessToken(wrongIssuer, secret); err == nil {
			t.Fatal("expected issuer error")
		}
	})

	t.Run("wrong token type", func(t *testing.T) {
		wrongType := mustSignAccessToken(t, secret, AccessTokenClaims{
			TokenType: "refresh",
			SessionID: "session-456",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    accessTokenIssuer,
				Subject:   "account-123",
				ID:        "wrong-type",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
			},
		}, jwt.SigningMethodHS256)
		if _, err := ValidateAccessToken(wrongType, secret); err == nil {
			t.Fatal("expected token type error")
		}
	})

	t.Run("missing required claim", func(t *testing.T) {
		missingClaim := mustSignAccessToken(t, secret, AccessTokenClaims{
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    accessTokenIssuer,
				Subject:   "account-123",
				ID:        "",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
			},
		}, jwt.SigningMethodHS256)
		if _, err := ValidateAccessToken(missingClaim, secret); err == nil {
			t.Fatal("expected missing-claim error")
		}
	})
}

func TestRefreshTokens(t *testing.T) {
	plaintext, hash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if plaintext == "" {
		t.Fatal("expected plaintext refresh token")
	}
	if hash == "" {
		t.Fatal("expected refresh token hash")
	}
	if strings.Contains(hash, plaintext) {
		t.Fatal("hash unexpectedly contains plaintext token")
	}

	otherPlaintext, otherHash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if plaintext == otherPlaintext {
		t.Fatal("expected unique refresh tokens")
	}
	if hash == otherHash {
		t.Fatal("expected distinct refresh-token hashes")
	}

	if got := HashRefreshToken(plaintext); got != hash {
		t.Fatalf("unexpected deterministic refresh-token hash: %s", got)
	}
}

func TestAPIKeys(t *testing.T) {
	plaintext, hash, prefix, last4, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if !strings.HasPrefix(plaintext, apiKeyPrefix) {
		t.Fatalf("unexpected plaintext prefix: %s", plaintext)
	}
	if prefix != apiKeyPrefix {
		t.Fatalf("unexpected key prefix: %s", prefix)
	}
	if len(last4) != 4 {
		t.Fatalf("unexpected last4: %s", last4)
	}
	if got := HashAPIKey(plaintext); got != hash {
		t.Fatalf("unexpected deterministic api key hash: %s", got)
	}
	if strings.Contains(hash, plaintext) {
		t.Fatal("hash unexpectedly contains plaintext key")
	}

	secondPlaintext, _, secondPrefix, secondLast4, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if plaintext == secondPlaintext {
		t.Fatal("expected unique api keys")
	}
	if secondPrefix != apiKeyPrefix {
		t.Fatalf("unexpected second key prefix: %s", secondPrefix)
	}
	if secondLast4 == "" {
		t.Fatal("expected second last4")
	}
}

func TestAnonymousIDs(t *testing.T) {
	secret := strings.Repeat("a", 32)
	now := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	sameDayOne, err := DeriveAnonymousID(secret, "203.0.113.1", "Mozilla/5.0", now)
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	sameDayTwo, err := DeriveAnonymousID(secret, "203.0.113.1", "Mozilla/5.0", now.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	if sameDayOne != sameDayTwo {
		t.Fatal("expected same-day identifiers to match")
	}

	nextDay, err := DeriveAnonymousID(secret, "203.0.113.1", "Mozilla/5.0", now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	if sameDayOne == nextDay {
		t.Fatal("expected daily rotation to change the identifier")
	}

	otherSecret, err := DeriveAnonymousID(strings.Repeat("b", 32), "203.0.113.1", "Mozilla/5.0", now)
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	if sameDayOne == otherSecret {
		t.Fatal("expected secret changes to change the identifier")
	}

	otherInput, err := DeriveAnonymousID(secret, "203.0.113.2", "Mozilla/5.0", now)
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	if sameDayOne == otherInput {
		t.Fatal("expected different inputs to change the identifier")
	}

	if strings.Contains(sameDayOne, "203.0.113.1") || strings.Contains(sameDayOne, "Mozilla/5.0") {
		t.Fatal("derived anonymous id leaked raw inputs")
	}

	encodedA, err := DeriveAnonymousID(secret, "1", "23", now)
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	encodedB, err := DeriveAnonymousID(secret, "12", "3", now)
	if err != nil {
		t.Fatalf("DeriveAnonymousID() error = %v", err)
	}
	if encodedA == encodedB {
		t.Fatal("expected length-prefixed inputs to remain collision safe")
	}
}

func TestAPIKeyDisplayParts(t *testing.T) {
	prefix, last4 := APIKeyDisplayParts("sd_live_abcdefghijklmnopqrstuvwxyz")
	if prefix != apiKeyPrefix {
		t.Fatalf("unexpected prefix: %s", prefix)
	}
	if last4 != "wxyz" {
		t.Fatalf("unexpected last4: %s", last4)
	}
}

func TestHashRefreshTokenDeterministic(t *testing.T) {
	token := "refresh-token"
	sum := sha256.Sum256([]byte(token))
	want := hex.EncodeToString(sum[:])
	if got := HashRefreshToken(token); got != want {
		t.Fatalf("unexpected hash: %s", got)
	}
}

func mustSignAccessToken(t *testing.T, secret string, claims AccessTokenClaims, method jwt.SigningMethod) string {
	t.Helper()

	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func flipTokenChar(ch byte) string {
	if ch == 'a' {
		return "b"
	}
	return "a"
}
