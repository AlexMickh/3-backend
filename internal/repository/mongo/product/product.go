package product_repository

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

type ProductRepository struct {
	collection *mongo.Collection
}

func New(ctx context.Context, db *mongo.Database) (*ProductRepository, error) {
	const op = "repository.mongo.product.New"

	collection := db.Collection("products")

	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "name", Value: "text"}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ProductRepository{
		collection: collection,
	}, nil
}

func (p *ProductRepository) SaveProduct(ctx context.Context, product *models.Product) (bson.ObjectID, error) {
	const op = "repository.mongo.product.SaveProduct"

	result, err := p.collection.InsertOne(ctx, product)
	if err != nil {
		var mongoErr mongo.WriteException
		if errors.As(err, &mongoErr) {
			for _, writeErr := range mongoErr.WriteErrors {
				if writeErr.Code == 11000 {
					return bson.NilObjectID, fmt.Errorf("%s: %w", op, errs.ErrProductAlreadyExists)
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

func (p *ProductRepository) SaveImage(ctx context.Context, id string, imageUrl string) error {
	const op = "repository.mongo.product.SaveImage"

	update := bson.M{
		"$set": bson.M{"image_url": imageUrl},
	}

	_, err := p.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *ProductRepository) ProductById(ctx context.Context, id bson.ObjectID) (*models.Product, error) {
	const op = "repository.mongo.product.ProductById"

	filter := bson.M{"_id": id}

	product := new(models.Product)
	err := p.collection.FindOne(ctx, filter).Decode(product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return product, nil
}

func (p *ProductRepository) ProductCards(
	ctx context.Context,
	page int,
	popularity bool,
	price int,
	categoryId bson.ObjectID,
	search string,
) ([]models.ProductCard, error) {
	const op = "repository.mongo.product.ProductCards"

	filter := bson.M{}
	sort := bson.D{}

	isCategoryIdSet := false

	if categoryId != bson.NilObjectID {
		filter["category_id"] = categoryId
		isCategoryIdSet = true
	}

	if search != "" {
		if isCategoryIdSet {
			delete(filter, "category_id")
			filter["$set"] = bson.M{
				"category_id": categoryId,
				"$text":       bson.M{"$search": fmt.Sprintf(`"%s"`, search)},
			}
		} else {
			filter["$text"] = bson.M{"$search": fmt.Sprintf(`"%s"`, search)}
		}
	}

	if popularity {
		sort = append(sort, bson.E{Key: "pieces_sold", Value: 1})
	}

	switch price {
	case 1:
		sort = append(sort, bson.E{Key: "price", Value: 1})
	case 0:
		sort = append(sort, bson.E{Key: "price", Value: -1})
	}

	opts := options.Find().
		SetSkip(int64(page * 10)).
		SetLimit(10).
		SetSort(sort)

	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	cards := make([]models.ProductCard, 0)
	for cursor.Next(ctx) {
		var card models.ProductCard

		err = cursor.Decode(&card)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		cards = append(cards, card)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return cards, nil
}

func (p *ProductRepository) UpdateProduct(ctx context.Context, productToUpdate *models.Product) error {
	const op = "repository.mongo.product.UpdateProduct"

	update := bson.M{}

	if productToUpdate.Name != "" {
		update["name"] = productToUpdate.Name
	}

	if productToUpdate.Description == "" {
		update["description"] = productToUpdate.Description
	}

	if productToUpdate.Price != -1 {
		update["price"] = productToUpdate.Price
	}

	if productToUpdate.Quantity == -1 {
		update["quantity"] = productToUpdate.Quantity
	}

	if len(productToUpdate.ExistingSizes) != 0 {
		update["existing_sizes"] = productToUpdate.ExistingSizes
	}

	if productToUpdate.ImageUrl != "" {
		update["image_url"] = productToUpdate.ImageUrl
	}

	if productToUpdate.Discount != -1 {
		update["discount"] = productToUpdate.Discount
	}

	if productToUpdate.DiscountExpiresAt != nil {
		update["discount_expires_at"] = productToUpdate.DiscountExpiresAt
	}

	update["updated_at"] = time.Now()

	result, err := p.collection.UpdateByID(ctx, productToUpdate.ID, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}

func (p *ProductRepository) DeleteProduct(ctx context.Context, id bson.ObjectID) error {
	const op = "repository.mongo.product.DeleteProduct"

	filter := bson.M{"_id": id}

	result, err := p.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}
