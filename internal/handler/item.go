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

type ItemHandler struct {
	itemService service.ItemService
	validate    *validator.Validate
}

func NewItemHandler(itemService service.ItemService, validate *validator.Validate) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
		validate:    validate,
	}
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	wishlistID, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}

	var req model.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, validationError(err))
		return
	}

	item, err := h.itemService.Create(r.Context(), userID, wishlistID, req)
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

	respondJSON(w, http.StatusCreated, item)
}

func (h *ItemHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	wishlistID, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}

	items, err := h.itemService.GetAll(r.Context(), userID, wishlistID)
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

	respondJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	wishlistID, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.itemService.GetByID(r.Context(), userID, wishlistID, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "item not found")
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	wishlistID, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var req model.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, validationError(err))
		return
	}

	item, err := h.itemService.Update(r.Context(), userID, wishlistID, id, req)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "item not found")
			return
		}
		if errors.Is(err, model.ErrForbidden) {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	wishlistID, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	if err := h.itemService.Delete(r.Context(), userID, wishlistID, id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondError(w, http.StatusNotFound, "item not found")
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
