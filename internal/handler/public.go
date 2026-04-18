package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/danil/cdek-wishlist/internal/service"
	"github.com/go-chi/chi/v5"
)

type PublicHandler struct {
	wishlistService service.WishlistService
	itemService     service.ItemService
}

func NewPublicHandler(wishlistService service.WishlistService, itemService service.ItemService) *PublicHandler {
	return &PublicHandler{
		wishlistService: wishlistService,
		itemService:     itemService,
	}
}

func (h *PublicHandler) GetWishlistByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	wishlist, err := h.wishlistService.GetByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "wishlist not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

func (h *PublicHandler) ReserveItem(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	if err := h.itemService.Reserve(r.Context(), token, itemID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "item not found")
			return
		}
		if errors.Is(err, model.ErrAlreadyReserved) {
			respondError(w, http.StatusConflict, "item is already reserved")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "item reserved successfully"})
}
