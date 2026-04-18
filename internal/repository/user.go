package repository

import (
	"context"

	"github.com/danil/cdek-wishlist/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type userRepository struct {
	db DB
}

func NewUserRepository(db DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, q, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	return mapPgError(err)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &u, nil
}

var _ UserRepository = (*userRepository)(nil)
