package sqlite_client

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func New(ctx context.Context, storagePath string) (*sql.DB, error) {
	const op = "pkg.clients.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}
