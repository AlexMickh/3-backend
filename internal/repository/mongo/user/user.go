package user_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func New(ctx context.Context, db *mongo.Database) (*UserRepository, error) {
	const op = "repository.mongo.user.New"

	collection := db.Collection("users")

	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &UserRepository{
		collection: collection,
	}, nil
}

func (u *UserRepository) SaveUser(ctx context.Context, user models.User) (bson.ObjectID, error) {
	const op = "repository.mongo.user.CreateUser"

	result, err := u.collection.InsertOne(ctx, user)
	if err != nil {
		var mongoErr mongo.WriteException
		if errors.As(err, &mongoErr) {
			for _, writeErr := range mongoErr.WriteErrors {
				if writeErr.Code == 11000 {
					return bson.NilObjectID, fmt.Errorf("%s: %w", op, errs.ErrUserAlreadyExists)
				}
			}
		}

		return bson.NilObjectID, fmt.Errorf("%s: %w", op, err)
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, fmt.Errorf("%s: failed to get id", op)
	}

	return id, nil
}

func (u *UserRepository) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "repository.mongo.user.UserByEmail"

	filter := bson.M{"email": email}
	var user models.User
	err := u.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (u *UserRepository) VerifyEmail(ctx context.Context, id bson.ObjectID) error {
	const op = "repository.mongo.user.VerifyEmail"

	update := bson.M{
		"$set": bson.M{
			"is_email_verified": true,
			"updated_at":        time.Now(),
		},
	}
	result, err := u.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
	}

	return nil
}

func (u *UserRepository) CanBuy(ctx context.Context, userId int64) (bool, error) {
	return false, nil
}
