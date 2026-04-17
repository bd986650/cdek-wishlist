package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/danil/cdek-wishlist/internal/repository"
)

type WishlistService interface {
	Create(ctx context.Context, userID int64, req model.CreateWishlistRequest) (*model.Wishlist, error)
	GetByID(ctx context.Context, userID, id int64) (*model.Wishlist, error)
	GetAll(ctx context.Context, userID int64) ([]model.Wishlist, error)
	Update(ctx context.Context, userID, id int64, req model.UpdateWishlistRequest) (*model.Wishlist, error)
	Delete(ctx context.Context, userID, id int64) error
	GetByToken(ctx context.Context, token string) (*model.Wishlist, error)
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
	itemRepo     repository.ItemRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository, itemRepo repository.ItemRepository) WishlistService {
	return &wishlistService{
		wishlistRepo: wishlistRepo,
		itemRepo:     itemRepo,
	}
}

func (s *wishlistService) Create(ctx context.Context, userID int64, req model.CreateWishlistRequest) (*model.Wishlist, error) {
	w := &model.Wishlist{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
	}

	// Генерируем токен. На случай коллизии по UNIQUE(token) — повторяем несколько раз.
	const attempts = 5
	for i := 0; i < attempts; i++ {
		token, err := generateTokenHex(32)
		if err != nil {
			return nil, err
		}
		w.Token = token

		if err := s.wishlistRepo.Create(ctx, w); err != nil {
			if errors.Is(err, model.ErrAlreadyExists) {
				// коллизия токена — пробуем ещё раз
				continue
			}
			return nil, err
		}
		return w, nil
	}

	return nil, model.ErrAlreadyExists
}

func (s *wishlistService) GetByID(ctx context.Context, userID, id int64) (*model.Wishlist, error) {
	w, err := s.wishlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	items, err := s.itemRepo.GetAllByWishlistID(ctx, w.ID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}
	w.Items = items

	return w, nil
}

func (s *wishlistService) GetAll(ctx context.Context, userID int64) ([]model.Wishlist, error) {
	return s.wishlistRepo.GetAllByUserID(ctx, userID)
}

func (s *wishlistService) Update(ctx context.Context, userID, id int64, req model.UpdateWishlistRequest) (*model.Wishlist, error) {
	w, err := s.wishlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	if req.Title != nil {
		w.Title = *req.Title
	}
	if req.Description != nil {
		w.Description = *req.Description
	}
	if req.EventDate != nil {
		w.EventDate = *req.EventDate
	}

	if err := s.wishlistRepo.Update(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

func (s *wishlistService) Delete(ctx context.Context, userID, id int64) error {
	w, err := s.wishlistRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if w.UserID != userID {
		return model.ErrForbidden
	}

	return s.wishlistRepo.Delete(ctx, id)
}

func (s *wishlistService) GetByToken(ctx context.Context, token string) (*model.Wishlist, error) {
	w, err := s.wishlistRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	items, err := s.itemRepo.GetAllByWishlistID(ctx, w.ID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}
	w.Items = items

	return w, nil
}

func generateTokenHex(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// небольшая страховка от случайного использования локального времени где-то рядом
var _ = time.UTC

