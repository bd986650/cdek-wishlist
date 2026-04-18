package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestWishlistRepository_Delete_NotFoundWhenZeroRows(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectExec("DELETE FROM wishlists").
		WithArgs(int64(10)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewWishlistRepository(poolIface)
	if err := repo.Delete(context.Background(), 10); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWishlistRepository_Update_NotFoundMapped(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectQuery("UPDATE wishlists").
		WithArgs("t", "d", "2026-01-01", int64(10)).
		WillReturnError(pgx.ErrNoRows)

	repo := NewWishlistRepository(poolIface)
	w := &model.Wishlist{ID: 10, Title: "t", Description: "d", EventDate: "2026-01-01"}
	if err := repo.Update(context.Background(), w); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWishlistRepository_GetAllByUserID_Success(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	ts1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	ts2 := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "title", "description", "event_date", "token", "created_at", "updated_at",
	}).AddRow(int64(1), int64(2), "a", "da", "e1", "tok1", ts1, ts1).
		AddRow(int64(2), int64(2), "b", "db", "e2", "tok2", ts2, ts2)

	mock.ExpectQuery("SELECT id, user_id, title, description, event_date, token, created_at, updated_at").
		WithArgs(int64(2)).
		WillReturnRows(rows)

	repo := NewWishlistRepository(poolIface)
	list, err := repo.GetAllByUserID(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 wishlists, got %d", len(list))
	}
	if list[0].ID != 1 || list[0].Token != "tok1" {
		t.Fatalf("unexpected first item: %#v", list[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
