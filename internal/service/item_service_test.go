package service

import (
	"context"
	"errors"
	"testing"

	"github.com/danil/cdek-wishlist/internal/model"
)

func TestItemService_Create_Forbidden(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 2}, nil
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewItemService(itemRepo, wlRepo)
	_, err := svc.Create(context.Background(), 1, 10, model.CreateItemRequest{Name: "x", Priority: 1})
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestItemService_GetByID_WrongWishlist_ReturnsNotFound(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 1}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Item, error) {
			return &model.Item{ID: id, WishlistID: 999}, nil
		},
	}

	svc := NewItemService(itemRepo, wlRepo)
	_, err := svc.GetByID(context.Background(), 1, 10, 5)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestItemService_Reserve_WrongWishlist_ReturnsNotFound(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByTokenFn: func(ctx context.Context, token string) (*model.Wishlist, error) {
			return &model.Wishlist{ID: 10, Token: token}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Item, error) {
			return &model.Item{ID: id, WishlistID: 999}, nil
		},
	}

	svc := NewItemService(itemRepo, wlRepo)
	err := svc.Reserve(context.Background(), "tok", 1)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestItemService_Reserve_AlreadyReservedShortCircuit(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByTokenFn: func(ctx context.Context, token string) (*model.Wishlist, error) {
			return &model.Wishlist{ID: 10, Token: token}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Item, error) {
			return &model.Item{ID: id, WishlistID: 10, IsReserved: true}, nil
		},
		reserveFn: func(ctx context.Context, id int64) error {
			t.Fatalf("reserve should not be called when already reserved")
			return nil
		},
	}

	svc := NewItemService(itemRepo, wlRepo)
	err := svc.Reserve(context.Background(), "tok", 1)
	if !errors.Is(err, model.ErrAlreadyReserved) {
		t.Fatalf("expected ErrAlreadyReserved, got %v", err)
	}
}

func TestItemService_Reserve_Success(t *testing.T) {
	called := false
	wlRepo := &mockWishlistRepo{
		getByTokenFn: func(ctx context.Context, token string) (*model.Wishlist, error) {
			return &model.Wishlist{ID: 10, Token: token}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Item, error) {
			return &model.Item{ID: id, WishlistID: 10, IsReserved: false}, nil
		},
		reserveFn: func(ctx context.Context, id int64) error {
			called = true
			if id != 7 {
				t.Fatalf("unexpected id: %d", id)
			}
			return nil
		},
	}

	svc := NewItemService(itemRepo, wlRepo)
	if err := svc.Reserve(context.Background(), "tok", 7); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatalf("expected reserve call")
	}
}

func TestItemService_Reserve_PropagatesRepoReserveError(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByTokenFn: func(ctx context.Context, token string) (*model.Wishlist, error) {
			return &model.Wishlist{ID: 10, Token: token}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Item, error) {
			return &model.Item{ID: id, WishlistID: 10, IsReserved: false}, nil
		},
		reserveFn: func(ctx context.Context, id int64) error {
			return errors.New("db")
		},
	}

	svc := NewItemService(itemRepo, wlRepo)
	err := svc.Reserve(context.Background(), "tok", 7)
	if err == nil || err.Error() != "db" {
		t.Fatalf("expected db error, got %v", err)
	}
}
