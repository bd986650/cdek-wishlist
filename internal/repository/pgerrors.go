package repository

import (
	"errors"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation = "23505"
)

func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgUniqueViolation {
			return model.ErrAlreadyExists
		}
	}

	return err
}

