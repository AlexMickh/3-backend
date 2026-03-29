package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/pkg/utils/retry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(
	ctx context.Context,
	username string,
	password string,
	host string,
	port int,
	database string,
	minPools int,
	maxPools int,
) (*pgxpool.Pool, error) {
	const op = "pkg.clients.postgresql.New"

	var pool *pgxpool.Pool

	err := retry.WithDelay(5, 500*time.Millisecond, func() error {
		var err error

		connString := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d",
			username,
			password,
			host,
			port,
			database,
			maxPools,
			minPools,
		)

		pool, err = pgxpool.New(ctx, connString)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = pool.Ping(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pool, nil
}
