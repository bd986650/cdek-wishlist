package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danil/cdek-wishlist/internal/config"
	"github.com/danil/cdek-wishlist/internal/handler"
	"github.com/danil/cdek-wishlist/internal/repository"
	"github.com/danil/cdek-wishlist/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	userRepo := repository.NewUserRepository(pool)
	wishlistRepo := repository.NewWishlistRepository(pool)
	itemRepo := repository.NewItemRepository(pool)

	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, int64(cfg.JWT.Expiration.Seconds()))
	wishlistSvc := service.NewWishlistService(wishlistRepo, itemRepo)
	itemSvc := service.NewItemService(itemRepo, wishlistRepo)

	validate := validator.New()

	h := handler.Handlers{
		Auth:     handler.NewAuthHandler(authSvc, validate),
		Wishlist: handler.NewWishlistHandler(wishlistSvc, validate),
		Item:     handler.NewItemHandler(itemSvc, validate),
		Public:   handler.NewPublicHandler(wishlistSvc, itemSvc),
	}

	router := handler.NewRouter(h, cfg.JWT.Secret)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	pool.Close()
	log.Println("server stopped gracefully")
}f
