package handler

import (
	"github.com/danil/cdek-wishlist/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Handlers struct {
	Auth     *AuthHandler
	Wishlist *WishlistHandler
	Item     *ItemHandler
	Public   *PublicHandler
}

func NewRouter(h Handlers, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.CORS)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Auth.Register)
		r.Post("/auth/login", h.Auth.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtSecret))

			r.Route("/wishlists", func(r chi.Router) {
				r.Post("/", h.Wishlist.Create)
				r.Get("/", h.Wishlist.GetAll)
				r.Get("/{id}", h.Wishlist.GetByID)
				r.Put("/{id}", h.Wishlist.Update)
				r.Delete("/{id}", h.Wishlist.Delete)

				r.Route("/{wishlistID}/items", func(r chi.Router) {
					r.Post("/", h.Item.Create)
					r.Get("/", h.Item.GetAll)
					r.Get("/{itemID}", h.Item.GetByID)
					r.Put("/{itemID}", h.Item.Update)
					r.Delete("/{itemID}", h.Item.Delete)
				})
			})
		})

		r.Get("/shared/{token}", h.Public.GetWishlistByToken)
		r.Post("/shared/{token}/items/{itemID}/reserve", h.Public.ReserveItem)
	})

	return r
}
