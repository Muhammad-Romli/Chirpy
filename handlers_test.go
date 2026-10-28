package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateUser_InvalidJSON(t *testing.T) {
	cfg := &apiConfig{}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()

	cfg.createUserHandler(rec, req)

	if rec.Code != 400 {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	cfg := &apiConfig{}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader("{'Email':'', 'password':''}"))
	rec := httptest.NewRecorder()

	cfg.createUserHandler(rec, req)

	if rec.Code != 400 {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	cfg := &apiConfig{}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()

	cfg.createUserHandler(rec, req)

	if rec.Code != 400 {
		t.Errorf("want 400, got %d", rec.Code)
	}
}
