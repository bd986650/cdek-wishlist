package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/go-playground/validator/v10"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error)
	loginFn    func(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error)
}

func (m *mockAuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error) {
	if m.registerFn == nil {
		return nil, errors.New("registerFn not set")
	}
	return m.registerFn(ctx, req)
}

func (m *mockAuthService) Login(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error) {
	if m.loginFn == nil {
		return nil, errors.New("loginFn not set")
	}
	return m.loginFn(ctx, req)
}

func TestAuthHandler_Register_Conflict(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error) {
			return nil, model.ErrAlreadyExists
		},
	}

	v := validator.New()
	h := NewAuthHandler(svc, v)

	body := bytes.NewBufferString(`{"email":"a@b.c","password":"password12"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "already exists") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error) {
			return &model.TokenResponse{Token: "t1"}, nil
		},
	}

	v := validator.New()
	h := NewAuthHandler(svc, v)

	body := bytes.NewBufferString(`{"email":"a@b.c","password":"password12"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp["token"] != "t1" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}
