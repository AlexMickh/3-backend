package category_repository

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

type CategoryRepository struct {
	collection *mongo.Collection
}

func New(ctx context.Context, db *mongo.Database) (*CategoryRepository, error) {
	const op = "repository.mongo.category.New"

	collection := db.Collection("categories")

	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &CategoryRepository{
		collection: collection,
	}, nil
}

func (c *CategoryRepository) SaveCategory(ctx context.Context, category models.Category) (bson.ObjectID, error) {
	const op = "repository.mongo.category.SaveCategory"

	result, err := c.collection.InsertOne(ctx, category)
	if err != nil {
		var mongoErr mongo.WriteException
		if errors.As(err, &mongoErr) {
			for _, writeErr := range mongoErr.WriteErrors {
				if writeErr.Code == 11000 {
					return bson.NilObjectID, fmt.Errorf("%s: %w", op, errs.ErrCategoryAlreadyExists)
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

func (c *CategoryRepository) DeleteCategory(ctx context.Context, id bson.ObjectID) error {
	const op = "repository.mongo.category.DeleteCategory"

	filter := bson.M{"_id": id}
	result, err := c.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrCategoryNotFound)
	}

	return nil
}

func (c *CategoryRepository) AllCategories(ctx context.Context) ([]models.Category, error) {
	const op = "repository.mongo.category.AllCategories"

	cursor, err := c.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	categories := make([]models.Category, 0)
	for cursor.Next(ctx) {
		var category models.Category

		err := cursor.Decode(&category)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		categories = append(categories, category)
	}

	return categories, nil
}
