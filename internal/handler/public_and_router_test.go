package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danil/cdek-wishlist/internal/middleware"
	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/go-playground/validator/v10"
)

type mockWishlistServiceForHTTP struct {
	getByTokenFn func(ctx context.Context, token string) (*model.Wishlist, error)
	getAllFn     func(ctx context.Context, userID int64) ([]model.Wishlist, error)
}

func (m *mockWishlistServiceForHTTP) Create(ctx context.Context, userID int64, req model.CreateWishlistRequest) (*model.Wishlist, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWishlistServiceForHTTP) GetByID(ctx context.Context, userID, id int64) (*model.Wishlist, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWishlistServiceForHTTP) GetAll(ctx context.Context, userID int64) ([]model.Wishlist, error) {
	if m.getAllFn == nil {
		return nil, errors.New("getAllFn not set")
	}
	return m.getAllFn(ctx, userID)
}

func (m *mockWishlistServiceForHTTP) Update(ctx context.Context, userID, id int64, req model.UpdateWishlistRequest) (*model.Wishlist, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWishlistServiceForHTTP) Delete(ctx context.Context, userID, id int64) error {
	return errors.New("not implemented")
}

func (m *mockWishlistServiceForHTTP) GetByToken(ctx context.Context, token string) (*model.Wishlist, error) {
	if m.getByTokenFn == nil {
		return nil, errors.New("getByTokenFn not set")
	}
	return m.getByTokenFn(ctx, token)
}

type mockItemServiceForHTTP struct {
	reserveFn func(ctx context.Context, token string, itemID int64) error
}

func (m *mockItemServiceForHTTP) Create(ctx context.Context, userID, wishlistID int64, req model.CreateItemRequest) (*model.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemServiceForHTTP) GetByID(ctx context.Context, userID, wishlistID, id int64) (*model.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemServiceForHTTP) GetAll(ctx context.Context, userID, wishlistID int64) ([]model.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemServiceForHTTP) Update(ctx context.Context, userID, wishlistID, id int64, req model.UpdateItemRequest) (*model.Item, error) {
	return nil, errors.New("not implemented")
}

func (m *mockItemServiceForHTTP) Delete(ctx context.Context, userID, wishlistID, id int64) error {
	return errors.New("not implemented")
}

func (m *mockItemServiceForHTTP) Reserve(ctx context.Context, token string, itemID int64) error {
	if m.reserveFn == nil {
		return errors.New("reserveFn not set")
	}
	return m.reserveFn(ctx, token, itemID)
}

func TestRouter_JWTAuthMissingHeader(t *testing.T) {
	wl := &mockWishlistServiceForHTTP{
		getAllFn: func(ctx context.Context, userID int64) ([]model.Wishlist, error) {
			panic("should not call service without auth")
		},
	}
	it := &mockItemServiceForHTTP{}

	h := Handlers{
		Auth:     NewAuthHandler(&mockAuthService{}, validator.New()),
		Wishlist: NewWishlistHandler(wl, validator.New()),
		Item:     NewItemHandler(it, validator.New()),
		Public:   NewPublicHandler(wl, it),
	}
	r := NewRouter(h, "secret")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wishlists", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "authorization header required") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestRouter_PublicReserve_Conflict(t *testing.T) {
	wl := &mockWishlistServiceForHTTP{}
	it := &mockItemServiceForHTTP{
		reserveFn: func(ctx context.Context, token string, itemID int64) error {
			if token != "tok" || itemID != 7 {
				t.Fatalf("unexpected args: %s %d", token, itemID)
			}
			return model.ErrAlreadyReserved
		},
	}

	h := Handlers{
		Auth:     NewAuthHandler(&mockAuthService{}, validator.New()),
		Wishlist: NewWishlistHandler(wl, validator.New()),
		Item:     NewItemHandler(it, validator.New()),
		Public:   NewPublicHandler(wl, it),
	}
	r := NewRouter(h, "secret")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shared/tok/items/7/reserve", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestMiddleware_GetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, int64(123))
	if middleware.GetUserID(ctx) != 123 {
		t.Fatalf("unexpected user id")
	}
}
