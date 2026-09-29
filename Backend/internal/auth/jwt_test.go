package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	manager := NewManager("test-secret", time.Hour, "pulsegraph-test")
	user := User{ID: 42, Name: "Analyst", Email: "a@example.com", Role: "analyst"}

	token, expiresAt, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token == "" {
		t.Fatal("expected a signed token")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiry should be in the future, got %s", expiresAt)
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "analyst" || claims.Email != "a@example.com" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	manager := NewManager("test-secret", time.Hour, "pulsegraph-test")
	token, _, err := manager.GenerateToken(User{ID: 1, Email: "x@example.com", Role: "viewer"})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := manager.ParseToken(token + "x"); err == nil {
		t.Fatal("expected an error for a tampered token")
	}
}

func TestParseRejectsWrongIssuer(t *testing.T) {
	issuerA := NewManager("shared-secret", time.Hour, "issuer-a")
	issuerB := NewManager("shared-secret", time.Hour, "issuer-b")

	token, _, err := issuerA.GenerateToken(User{ID: 2, Email: "b@example.com", Role: "viewer"})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := issuerB.ParseToken(token); err == nil {
		t.Fatal("expected issuer validation to reject a foreign token")
	}
}
