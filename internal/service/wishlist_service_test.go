package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
)

type mockWishlistRepo struct {
	createFn          func(ctx context.Context, w *model.Wishlist) error
	getByIDFn         func(ctx context.Context, id int64) (*model.Wishlist, error)
	getAllByUserIDFn  func(ctx context.Context, userID int64) ([]model.Wishlist, error)
	updateFn          func(ctx context.Context, w *model.Wishlist) error
	deleteFn          func(ctx context.Context, id int64) error
	getByTokenFn      func(ctx context.Context, token string) (*model.Wishlist, error)
}

func (m *mockWishlistRepo) Create(ctx context.Context, w *model.Wishlist) error {
	if m.createFn == nil {
		return errors.New("createFn not set")
	}
	return m.createFn(ctx, w)
}

func (m *mockWishlistRepo) GetByID(ctx context.Context, id int64) (*model.Wishlist, error) {
	if m.getByIDFn == nil {
		return nil, errors.New("getByIDFn not set")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockWishlistRepo) GetAllByUserID(ctx context.Context, userID int64) ([]model.Wishlist, error) {
	if m.getAllByUserIDFn == nil {
		return nil, errors.New("getAllByUserIDFn not set")
	}
	return m.getAllByUserIDFn(ctx, userID)
}

func (m *mockWishlistRepo) Update(ctx context.Context, w *model.Wishlist) error {
	if m.updateFn == nil {
		return errors.New("updateFn not set")
	}
	return m.updateFn(ctx, w)
}

func (m *mockWishlistRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn == nil {
		return errors.New("deleteFn not set")
	}
	return m.deleteFn(ctx, id)
}

func (m *mockWishlistRepo) GetByToken(ctx context.Context, token string) (*model.Wishlist, error) {
	if m.getByTokenFn == nil {
		return nil, errors.New("getByTokenFn not set")
	}
	return m.getByTokenFn(ctx, token)
}

type mockItemRepo struct {
	createFn              func(ctx context.Context, it *model.Item) error
	getByIDFn             func(ctx context.Context, id int64) (*model.Item, error)
	getAllByWishlistIDFn  func(ctx context.Context, wishlistID int64) ([]model.Item, error)
	updateFn              func(ctx context.Context, it *model.Item) error
	deleteFn              func(ctx context.Context, id int64) error
	reserveFn             func(ctx context.Context, id int64) error
}

func (m *mockItemRepo) Create(ctx context.Context, it *model.Item) error {
	if m.createFn == nil {
		return errors.New("createFn not set")
	}
	return m.createFn(ctx, it)
}

func (m *mockItemRepo) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	if m.getByIDFn == nil {
		return nil, errors.New("getByIDFn not set")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockItemRepo) GetAllByWishlistID(ctx context.Context, wishlistID int64) ([]model.Item, error) {
	if m.getAllByWishlistIDFn == nil {
		return nil, errors.New("getAllByWishlistIDFn not set")
	}
	return m.getAllByWishlistIDFn(ctx, wishlistID)
}

func (m *mockItemRepo) Update(ctx context.Context, it *model.Item) error {
	if m.updateFn == nil {
		return errors.New("updateFn not set")
	}
	return m.updateFn(ctx, it)
}

func (m *mockItemRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn == nil {
		return errors.New("deleteFn not set")
	}
	return m.deleteFn(ctx, id)
}

func (m *mockItemRepo) Reserve(ctx context.Context, id int64) error {
	if m.reserveFn == nil {
		return errors.New("reserveFn not set")
	}
	return m.reserveFn(ctx, id)
}

func TestWishlistService_Create_RetriesOnTokenCollision(t *testing.T) {
	calls := 0
	wlRepo := &mockWishlistRepo{
		createFn: func(ctx context.Context, w *model.Wishlist) error {
			calls++
			if calls == 1 {
				return model.ErrAlreadyExists
			}
			w.ID = 100
			w.CreatedAt = time.Now().UTC()
			w.UpdatedAt = w.CreatedAt
			return nil
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewWishlistService(wlRepo, itemRepo)
	w, err := svc.Create(context.Background(), 1, model.CreateWishlistRequest{Title: "t"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if w.ID != 100 {
		t.Fatalf("expected id assigned, got %+v", w)
	}
	if w.Token == "" {
		t.Fatalf("expected token")
	}
	if calls != 2 {
		t.Fatalf("expected 2 create attempts, got %d", calls)
	}
}

func TestWishlistService_Create_FailsAfterMaxAttempts(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		createFn: func(ctx context.Context, w *model.Wishlist) error {
			return model.ErrAlreadyExists
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewWishlistService(wlRepo, itemRepo)
	_, err := svc.Create(context.Background(), 1, model.CreateWishlistRequest{Title: "t"})
	if !errors.Is(err, model.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestWishlistService_GetByID_Forbidden(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 2}, nil
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewWishlistService(wlRepo, itemRepo)
	_, err := svc.GetByID(context.Background(), 1, 10)
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestWishlistService_GetByID_LoadsItems_IgnoresNotFound(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 1}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getAllByWishlistIDFn: func(ctx context.Context, wishlistID int64) ([]model.Item, error) {
			return nil, model.ErrNotFound
		},
	}

	svc := NewWishlistService(wlRepo, itemRepo)
	w, err := svc.GetByID(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if w.Items != nil {
		t.Fatalf("expected nil items slice on not found, got %#v", w.Items)
	}
}

func TestWishlistService_GetByID_PropagatesItemsError(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 1}, nil
		},
	}
	itemRepo := &mockItemRepo{
		getAllByWishlistIDFn: func(ctx context.Context, wishlistID int64) ([]model.Item, error) {
			return nil, errors.New("boom")
		},
	}

	svc := NewWishlistService(wlRepo, itemRepo)
	_, err := svc.GetByID(context.Background(), 1, 10)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom, got %v", err)
	}
}

func TestWishlistService_Update_Forbidden(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 2, Title: "old"}, nil
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewWishlistService(wlRepo, itemRepo)
	title := "new"
	_, err := svc.Update(context.Background(), 1, 10, model.UpdateWishlistRequest{Title: &title})
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestWishlistService_Delete_Forbidden(t *testing.T) {
	wlRepo := &mockWishlistRepo{
		getByIDFn: func(ctx context.Context, id int64) (*model.Wishlist, error) {
			return &model.Wishlist{ID: id, UserID: 2}, nil
		},
	}
	itemRepo := &mockItemRepo{}

	svc := NewWishlistService(wlRepo, itemRepo)
	err := svc.Delete(context.Background(), 1, 10)
	if !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
