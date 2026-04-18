package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

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

	// Retry on UNIQUE(token) collision — collision probability is negligible (32 random bytes),
	// but we guard against it anyway.
	const maxAttempts = 5
	for range maxAttempts {
		token, err := generateTokenHex(32)
		if err != nil {
			return nil, err
		}
		w.Token = token

		if err := s.wishlistRepo.Create(ctx, w); err != nil {
			if errors.Is(err, model.ErrAlreadyExists) {
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
	if err != nil {
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
	if err != nil {
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
