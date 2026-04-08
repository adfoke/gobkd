package repository

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTransactorCommitsOnSuccess(t *testing.T) {
	db := openRepositoryTestDB(t)
	transactor := NewTransactor(db)

	if err := transactor.WithinTransaction(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`)
		return err
	}); err != nil {
		t.Fatalf("expected commit success, got %v", err)
	}

	if _, err := db.Exec(`INSERT INTO items (name) VALUES ('ok')`); err != nil {
		t.Fatalf("expected committed table to exist, got %v", err)
	}
}

func TestTransactorRollsBackOnError(t *testing.T) {
	db := openRepositoryTestDB(t)
	transactor := NewTransactor(db)

	wantErr := errors.New("boom")
	err := transactor.WithinTransaction(context.Background(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(context.Background(), `CREATE TABLE rolled_back_items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected rollback error %v, got %v", wantErr, err)
	}

	if _, err := db.Exec(`INSERT INTO rolled_back_items (name) VALUES ('missing')`); err == nil {
		t.Fatal("expected table creation to roll back")
	}
}

func openRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "repository.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return db
}
