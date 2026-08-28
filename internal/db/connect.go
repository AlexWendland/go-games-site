package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// DB wraps the sqlc Queries and the underlying sql.DB to allow transaction management.
type DB struct {
	*Queries
	db *sql.DB
}

// Open opens a SQLite database at the given path, applies any pending migrations,
// and returns a DB ready for use.
func Open(dbPath string, migrationsFS fs.FS) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	goose.SetBaseFS(migrationsFS)
	goose.SetDialect("sqlite3")
	if err := goose.Up(sqlDB, "."); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{
		Queries: New(sqlDB),
		db:      sqlDB,
	}, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// WithTx runs fn inside a transaction. If fn returns an error the transaction
// is rolled back, otherwise it is committed.
func (d *DB) WithTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(d.Queries.WithTx(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
