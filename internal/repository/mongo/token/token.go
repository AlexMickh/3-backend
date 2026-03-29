package token_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TokenRepository struct {
	collection *mongo.Collection
}

func New(ctx context.Context, db *mongo.Database) (*TokenRepository, error) {
	const op = "repository.mongo.token.New"

	collection := db.Collection("tokens")

	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &TokenRepository{
		collection: collection,
	}, nil
}

func (t *TokenRepository) SaveToken(ctx context.Context, token models.Token) error {
	const op = "repository.mongo.token.SaveToken"

	_, err := t.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (t *TokenRepository) Token(ctx context.Context, token string, tokenType models.TokenType) (models.Token, error) {
	const op = "repository.mongo.token.Token"

	filter := bson.M{
		"$and": bson.A{
			bson.M{"token": token},
			bson.M{"type": tokenType},
		},
	}
	var tok models.Token
	err := t.collection.FindOneAndDelete(ctx, filter).Decode(&tok)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Token{}, fmt.Errorf("%s: %w", op, errs.ErrTokenNotFound)
		}

		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	tok.ExpiresAt = tok.ExpiresAt.Local()

	fmt.Println(tok.ExpiresAt)

	return tok, nil
}
