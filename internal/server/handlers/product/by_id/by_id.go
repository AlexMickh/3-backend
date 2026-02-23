package product_by_id

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ProductProvider interface {
	ProductById(ctx context.Context, id int64) (*models.Product, error)
}

// New godoc
//
//	@Summary		get product by id
//	@Description	get product by id
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"product id"
//	@Success		200	{object}	dtos.ProductByIdResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/products/{id} [get]
func New(validator *validator.Validate, productProvider ProductProvider) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.product.by_id.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Error("failed to get id", logger.Err(err))
			return response.Error("failed to get id", http.StatusBadRequest)
		}

		product, err := productProvider.ProductById(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrProductNotFound) {
				log.Error(errs.ErrProductNotFound.Error())
				return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to get product", logger.Err(err))
			return response.Error("failed to get product", http.StatusInternalServerError)
		}

		resp := dtos.ProductByIdResponse{
			ID:                id,
			Name:              product.Name,
			Description:       product.Description,
			Price:             product.Price,
			Quantity:          product.Quantity,
			ExistingSizes:     make([]string, 0, len(product.ExistingSizes)),
			ImageUrl:          product.ImageUrl,
			Discount:          product.Discount,
			DiscountExpiresAt: product.DiscountExpiresAt,
			Category: struct {
				ID   int64
				Name string
			}{
				ID:   product.Category.ID,
				Name: product.Category.Name,
			},
		}
		for _, v := range product.ExistingSizes {
			resp.ExistingSizes = append(resp.ExistingSizes, string(v))
		}

		render.JSON(w, r, resp)

		return nil
	}
}
