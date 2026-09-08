package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/security"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestIssuedAccessTokenIDFitsSessionUUIDColumn(t *testing.T) {
	secret := strings.Repeat("test-only-secret-", 3)
	issuer, err := services.NewSecurityAccessTokenIssuer(secret)
	if err != nil {
		t.Fatal(err)
	}
	token, jti, err := issuer.Issue(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	// Use the same conversion as SessionRepository.RotateSessionTokens.
	id, err := uuidFromString(jti)
	if err != nil || !id.Valid {
		t.Fatalf("issued JTI cannot be stored in sessions.access_token_jti: %v", err)
	}
	if id.Bytes[6]>>4 != 4 || id.Bytes[8]>>6 != 2 {
		t.Fatal("expected a random version-4 UUID")
	}
	claims, err := security.ValidateAccessToken(token, secret)
	if err != nil || claims.ID != jti {
		t.Fatal("signed token JTI differs from persisted JTI")
	}
	_, next, err := issuer.Issue(context.Background(), claims.Subject, claims.SessionID, time.Minute)
	if err != nil || next == jti {
		t.Fatal("token rotation must issue a new JTI")
	}
}
