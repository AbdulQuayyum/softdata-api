package db

import (
	"strings"
	"testing"
)

func TestRotateSessionTokensQueryShape(t *testing.T) {
	if got := strings.Count(rotateSessionTokens, "UPDATE sessions"); got != 1 {
		t.Fatalf("expected one UPDATE statement, got %d", got)
	}
	if !strings.Contains(rotateSessionTokens, "WHERE id = $1") {
		t.Fatal("expected session id match in WHERE clause")
	}
	if !strings.Contains(rotateSessionTokens, "refresh_token_hash = $2") {
		t.Fatal("expected current refresh token hash match in WHERE clause")
	}
	if !strings.Contains(rotateSessionTokens, "revoked_at IS NULL") {
		t.Fatal("expected revoked sessions to be excluded")
	}
	if !strings.Contains(rotateSessionTokens, "expires_at > now()") {
		t.Fatal("expected expired sessions to be excluded")
	}
	if !strings.Contains(rotateSessionTokens, "refresh_token_hash = $3") {
		t.Fatal("expected refresh token hash to be updated")
	}
	if !strings.Contains(rotateSessionTokens, "access_token_jti = $4") {
		t.Fatal("expected access token jti to be updated")
	}
	if strings.Contains(rotateSessionTokens, "expires_at =") {
		t.Fatal("expected absolute expires_at to remain unchanged")
	}
}

func TestRotateSessionTokensParamsFields(t *testing.T) {
	var params RotateSessionTokensParams
	if params.SessionID.String() != "" {
		t.Fatal("unexpected zero session id")
	}
	if params.CurrentRefreshTokenHash != "" {
		t.Fatal("unexpected zero current refresh token hash")
	}
	if params.NewRefreshTokenHash != "" {
		t.Fatal("unexpected zero new refresh token hash")
	}
	if params.NewAccessTokenJti.String() != "" {
		t.Fatal("unexpected zero new access token jti")
	}
}
