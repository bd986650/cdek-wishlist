package service

import (
	"context"
	"errors"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/danil/cdek-wishlist/internal/repository"
)

type ItemService interface {
	Create(ctx context.Context, userID, wishlistID int64, req model.CreateItemRequest) (*model.Item, error)
	GetByID(ctx context.Context, userID, wishlistID, id int64) (*model.Item, error)
	GetAll(ctx context.Context, userID, wishlistID int64) ([]model.Item, error)
	Update(ctx context.Context, userID, wishlistID, id int64, req model.UpdateItemRequest) (*model.Item, error)
	Delete(ctx context.Context, userID, wishlistID, id int64) error
	Reserve(ctx context.Context, token string, itemID int64) error
}

type itemService struct {
	itemRepo     repository.ItemRepository
	wishlistRepo repository.WishlistRepository
}

func NewItemService(itemRepo repository.ItemRepository, wishlistRepo repository.WishlistRepository) ItemService {
	return &itemService{
		itemRepo:     itemRepo,
		wishlistRepo: wishlistRepo,
	}
}

func (s *itemService) Create(ctx context.Context, userID, wishlistID int64, req model.CreateItemRequest) (*model.Item, error) {
	w, err := s.wishlistRepo.GetByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	it := &model.Item{
		WishlistID:  wishlistID,
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Priority:    req.Priority,
	}

	if err := s.itemRepo.Create(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *itemService) GetByID(ctx context.Context, userID, wishlistID, id int64) (*model.Item, error) {
	w, err := s.wishlistRepo.GetByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	it, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if it.WishlistID != wishlistID {
		return nil, model.ErrNotFound
	}
	return it, nil
}

func (s *itemService) GetAll(ctx context.Context, userID, wishlistID int64) ([]model.Item, error) {
	w, err := s.wishlistRepo.GetByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	return s.itemRepo.GetAllByWishlistID(ctx, wishlistID)
}

func (s *itemService) Update(ctx context.Context, userID, wishlistID, id int64, req model.UpdateItemRequest) (*model.Item, error) {
	w, err := s.wishlistRepo.GetByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, model.ErrForbidden
	}

	it, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if it.WishlistID != wishlistID {
		return nil, model.ErrNotFound
	}

	if req.Name != nil {
		it.Name = *req.Name
	}
	if req.Description != nil {
		it.Description = *req.Description
	}
	if req.URL != nil {
		it.URL = *req.URL
	}
	if req.Priority != nil {
		it.Priority = *req.Priority
	}

	if err := s.itemRepo.Update(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *itemService) Delete(ctx context.Context, userID, wishlistID, id int64) error {
	w, err := s.wishlistRepo.GetByID(ctx, wishlistID)
	if err != nil {
		return err
	}
	if w.UserID != userID {
		return model.ErrForbidden
	}

	it, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if it.WishlistID != wishlistID {
		return model.ErrNotFound
	}

	return s.itemRepo.Delete(ctx, id)
}

func (s *itemService) Reserve(ctx context.Context, token string, itemID int64) error {
	w, err := s.wishlistRepo.GetByToken(ctx, token)
	if err != nil {
		return err
	}

	it, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return err
	}
	// Не раскрываем существование item из других вишлистов по публичному токену.
	if it.WishlistID != w.ID {
		return model.ErrNotFound
	}

	if it.IsReserved {
		return model.ErrAlreadyReserved
	}

	err = s.itemRepo.Reserve(ctx, itemID)
	if err != nil {
		// репозиторий возвращает ErrAlreadyReserved, если update не обновил строку
		if errors.Is(err, model.ErrAlreadyReserved) {
			return model.ErrAlreadyReserved
		}
		return err
	}

	return nil
}

