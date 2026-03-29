package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/pkg/utils/retry"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func New(
	ctx context.Context,
	host string,
	port int,
	username string,
	password string,
	database string,
) (*mongo.Database, error) {
	const op = "pkg.clients.mongodb.New"

	connString := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=admin", username, password, host, port)

	var db *mongo.Database

	err := retry.WithDelay(5, 500*time.Millisecond, func() error {
		opts := options.Client().ApplyURI(connString)

		client, err := mongo.Connect(opts)
		if err != nil {
			return err
		}

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			return err
		}

		db = client.Database(database)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}
