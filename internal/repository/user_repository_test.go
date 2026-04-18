package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
)

func TestUserRepository_Create_UniqueViolationMapped(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("a@b.c", "hash").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	repo := NewUserRepository(poolIface)

	u := &model.User{Email: "a@b.c", PasswordHash: "hash"}
	if err := repo.Create(context.Background(), u); !errors.Is(err, model.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserRepository_GetByEmail_NotFoundMapped(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at`).
		WithArgs("a@b.c").
		WillReturnError(pgx.ErrNoRows)

	repo := NewUserRepository(poolIface)
	_, err = repo.GetByEmail(context.Background(), "a@b.c")
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	ts := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "created_at"}).
		AddRow(int64(1), "a@b.c", "hash", ts)

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at`).
		WithArgs("a@b.c").
		WillReturnRows(rows)

	repo := NewUserRepository(poolIface)
	u, err := repo.GetByEmail(context.Background(), "a@b.c")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.ID != 1 || u.Email != "a@b.c" || u.PasswordHash != "hash" || !u.CreatedAt.Equal(ts) {
		t.Fatalf("unexpected user: %#v", u)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

var _ DB = (pgxmock.PgxPoolIface)(nil)
