package repository

import (
	"context"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository interface {
	Create(ctx context.Context, item *model.Item) error
	GetByID(ctx context.Context, id int64) (*model.Item, error)
	GetAllByWishlistID(ctx context.Context, wishlistID int64) ([]model.Item, error)
	Update(ctx context.Context, item *model.Item) error
	Delete(ctx context.Context, id int64) error
	Reserve(ctx context.Context, id int64) error
}

type itemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(ctx context.Context, it *model.Item) error {
	const q = `
		INSERT INTO wishlist_items (wishlist_id, name, description, url, priority)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_reserved, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, q, it.WishlistID, it.Name, it.Description, it.URL, it.Priority).
		Scan(&it.ID, &it.IsReserved, &it.CreatedAt, &it.UpdatedAt)
	return mapPgError(err)
}

func (r *itemRepository) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	const q = `
		SELECT id, wishlist_id, name, description, url, priority, is_reserved, created_at, updated_at
		FROM wishlist_items
		WHERE id = $1
	`

	var it model.Item
	err := r.db.QueryRow(ctx, q, id).Scan(
		&it.ID,
		&it.WishlistID,
		&it.Name,
		&it.Description,
		&it.URL,
		&it.Priority,
		&it.IsReserved,
		&it.CreatedAt,
		&it.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &it, nil
}

func (r *itemRepository) GetAllByWishlistID(ctx context.Context, wishlistID int64) ([]model.Item, error) {
	const q = `
		SELECT id, wishlist_id, name, description, url, priority, is_reserved, created_at, updated_at
		FROM wishlist_items
		WHERE wishlist_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, q, wishlistID)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	var res []model.Item
	for rows.Next() {
		var it model.Item
		if err := rows.Scan(
			&it.ID,
			&it.WishlistID,
			&it.Name,
			&it.Description,
			&it.URL,
			&it.Priority,
			&it.IsReserved,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, mapPgError(err)
		}
		res = append(res, it)
	}
	if rows.Err() != nil {
		return nil, mapPgError(rows.Err())
	}

	return res, nil
}

func (r *itemRepository) Update(ctx context.Context, it *model.Item) error {
	const q = `
		UPDATE wishlist_items
		SET name = $1,
		    description = $2,
		    url = $3,
		    priority = $4,
		    updated_at = NOW()
		WHERE id = $5
		RETURNING wishlist_id, is_reserved, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, q, it.Name, it.Description, it.URL, it.Priority, it.ID).
		Scan(&it.WishlistID, &it.IsReserved, &it.CreatedAt, &it.UpdatedAt)
	return mapPgError(err)
}

func (r *itemRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM wishlist_items WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// Reserve делает атомарное бронирование: ставит is_reserved=true только если оно было false.
// Если строка не обновилась (RowsAffected=0), вызывающий слой решает: not found vs already reserved.
func (r *itemRepository) Reserve(ctx context.Context, id int64) error {
	const q = `
		UPDATE wishlist_items
		SET is_reserved = TRUE,
		    updated_at = NOW()
		WHERE id = $1 AND is_reserved = FALSE
	`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		// Это может быть либо "не существует", либо "уже зарезервировано".
		// Сервисный слой различит через GetByID/или более строгий UPDATE по wishlist_id.
		return model.ErrAlreadyReserved
	}
	return nil
}

var _ ItemRepository = (*itemRepository)(nil)

