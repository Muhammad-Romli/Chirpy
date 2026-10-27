package auth_test

import (
	"net/http"
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

func TestCorrectBearerToken(t *testing.T) {
	header := http.Header{}
	header.Add("Authorization", "Bearer helloworld")
	have, err := auth.GetBearerToken(header)
	want := "helloworld"
	if err != nil {
		t.Fatalf("Getting bearer token failed: %v", err)
	}
	if have != want {
		t.Errorf("Processed token and expected token does not match")
	}
}

func TestMalformedBearerToken(t *testing.T) {
	header := http.Header{}
	header.Add("Authorization", "only helloworld")
	_, err := auth.GetBearerToken(header)
	if err == nil {
		t.Errorf("Expected error since header malformed: %v", err)
	}
}

func TestMissingHeaderBearerToken(t *testing.T) {
	header := http.Header{}
	_, err := auth.GetBearerToken(header)
	if err == nil {
		t.Errorf("Expected error since header is not set: %v", err)
	}
}
