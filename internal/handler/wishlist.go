package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/danil/cdek-wishlist/internal/middleware"
	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/danil/cdek-wishlist/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type WishlistHandler struct {
	wishlistService service.WishlistService
	validate        *validator.Validate
}

func NewWishlistHandler(wishlistService service.WishlistService, validate *validator.Validate) *WishlistHandler {
	return &WishlistHandler{
		wishlistService: wishlistService,
		validate:        validate,
	}
}

func (h *WishlistHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req model.CreateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, validationError(err))
		return
	}

	wishlist, err := h.wishlistService.Create(r.Context(), userID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, wishlist)
}

func (h *WishlistHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	wishlists, err := h.wishlistService.GetAll(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, wishlists)
}

func (h *WishlistHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}

	wishlist, err := h.wishlistService.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "wishlist not found")
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

func (h *WishlistHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}

	var req model.UpdateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, validationError(err))
		return
	}

	wishlist, err := h.wishlistService.Update(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "wishlist not found")
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

func (h *WishlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}

	if err := h.wishlistService.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "wishlist not found")
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
