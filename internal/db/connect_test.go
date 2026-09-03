package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexWendland/go-games-site/internal/db"
	"github.com/AlexWendland/go-games-site/internal/db/migrations"
)

// Test utility to start up a in-memory database.
func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

// Smoke test the helper.
func TestOpen(t *testing.T) {
	_ = openTestDB(t)
}

func TestWithTx_commit(t *testing.T) {
	d := openTestDB(t)
	err := d.WithTx(context.Background(), func(q *db.Queries) error {
		return nil
	})
	if err != nil {
		t.Errorf("WithTx() unexpected error = %v", err)
	}
}

func TestWithTx_rollback(t *testing.T) {
	d := openTestDB(t)
	sentinel := errors.New("abort")
	err := d.WithTx(context.Background(), func(q *db.Queries) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("WithTx() error = %v, want %v", err, sentinel)
	}
}
