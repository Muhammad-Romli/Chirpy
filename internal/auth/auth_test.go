package auth_test

import (
	"testing"
	"time"

	auth "github.com/Muhammad-Romli/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestValidJWT(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	got, err := auth.ValidateJWT(token, "secret")
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if got != userID {
		t.Errorf("expected %v, got %v", userID, got)
	}
}

func TestExpiredJWT(t *testing.T) {
	token, _ := auth.MakeJWT(uuid.New(), "secret", -time.Hour)

	if _, err := auth.ValidateJWT(token, "secret"); err == nil {
		t.Error("expected an error for an expired token")
	}
}

func TestWrongSecret(t *testing.T) {
	token, _ := auth.MakeJWT(uuid.New(), "secret", time.Hour)

	if _, err := auth.ValidateJWT(token, "wrong-secret"); err == nil {
		t.Error("expected an error for the wrong secret")
	}
}
