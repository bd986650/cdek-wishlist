package repository

import (
	"context"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WishlistRepository interface {
	Create(ctx context.Context, wishlist *model.Wishlist) error
	GetByID(ctx context.Context, id int64) (*model.Wishlist, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]model.Wishlist, error)
	Update(ctx context.Context, wishlist *model.Wishlist) error
	Delete(ctx context.Context, id int64) error
	GetByToken(ctx context.Context, token string) (*model.Wishlist, error)
}

type wishlistRepository struct {
	db *pgxpool.Pool
}

func NewWishlistRepository(db *pgxpool.Pool) WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Create(ctx context.Context, w *model.Wishlist) error {
	const q = `
		INSERT INTO wishlists (user_id, title, description, event_date, token)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, q, w.UserID, w.Title, w.Description, w.EventDate, w.Token).
		Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
	return mapPgError(err)
}

func (r *wishlistRepository) GetByID(ctx context.Context, id int64) (*model.Wishlist, error) {
	const q = `
		SELECT id, user_id, title, description, event_date, token, created_at, updated_at
		FROM wishlists
		WHERE id = $1
	`

	var w model.Wishlist
	err := r.db.QueryRow(ctx, q, id).Scan(
		&w.ID,
		&w.UserID,
		&w.Title,
		&w.Description,
		&w.EventDate,
		&w.Token,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &w, nil
}

func (r *wishlistRepository) GetAllByUserID(ctx context.Context, userID int64) ([]model.Wishlist, error) {
	const q = `
		SELECT id, user_id, title, description, event_date, token, created_at, updated_at
		FROM wishlists
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	var res []model.Wishlist
	for rows.Next() {
		var w model.Wishlist
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Title,
			&w.Description,
			&w.EventDate,
			&w.Token,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, mapPgError(err)
		}
		res = append(res, w)
	}
	if rows.Err() != nil {
		return nil, mapPgError(rows.Err())
	}

	return res, nil
}

func (r *wishlistRepository) Update(ctx context.Context, w *model.Wishlist) error {
	const q = `
		UPDATE wishlists
		SET title = $1,
		    description = $2,
		    event_date = $3,
		    updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, q, w.Title, w.Description, w.EventDate, w.ID).Scan(&w.UpdatedAt)
	return mapPgError(err)
}

func (r *wishlistRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM wishlists WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *wishlistRepository) GetByToken(ctx context.Context, token string) (*model.Wishlist, error) {
	const q = `
		SELECT id, user_id, title, description, event_date, token, created_at, updated_at
		FROM wishlists
		WHERE token = $1
	`

	var w model.Wishlist
	err := r.db.QueryRow(ctx, q, token).Scan(
		&w.ID,
		&w.UserID,
		&w.Title,
		&w.Description,
		&w.EventDate,
		&w.Token,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &w, nil
}

var _ WishlistRepository = (*wishlistRepository)(nil)

