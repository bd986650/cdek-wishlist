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

func TestItemRepository_Delete_NotFoundWhenZeroRows(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectExec("DELETE FROM wishlist_items").
		WithArgs(int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewItemRepository(poolIface)
	if err := repo.Delete(context.Background(), 99); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestItemRepository_Reserve_NoRowsMappedToAlreadyReserved(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectExec("UPDATE wishlist_items").
		WithArgs(int64(7)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	repo := NewItemRepository(poolIface)
	if err := repo.Reserve(context.Background(), 7); !errors.Is(err, model.ErrAlreadyReserved) {
		t.Fatalf("expected ErrAlreadyReserved, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestItemRepository_Update_NotFoundMapped(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	mock.ExpectQuery("UPDATE wishlist_items").
		WithArgs("n", "d", "u", 5, int64(3)).
		WillReturnError(pgx.ErrNoRows)

	repo := NewItemRepository(poolIface)
	it := &model.Item{ID: 3, Name: "n", Description: "d", URL: "u", Priority: 5}
	if err := repo.Update(context.Background(), it); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestItemRepository_GetByID_Success(t *testing.T) {
	poolIface, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock: %v", err)
	}
	mock, ok := poolIface.(pgxmock.PgxCommonIface)
	if !ok {
		t.Fatalf("expected pgxmock expecter")
	}

	ts := time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)

	rows := pgxmock.NewRows([]string{
		"id", "wishlist_id", "name", "description", "url", "priority", "is_reserved", "created_at", "updated_at",
	}).AddRow(int64(5), int64(10), "n", "d", "u", 3, true, ts, ts)

	mock.ExpectQuery("SELECT id, wishlist_id, name, description, url, priority, is_reserved, created_at, updated_at").
		WithArgs(int64(5)).
		WillReturnRows(rows)

	repo := NewItemRepository(poolIface)
	it, err := repo.GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if it.ID != 5 || it.WishlistID != 10 || !it.IsReserved {
		t.Fatalf("unexpected item: %#v", it)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
