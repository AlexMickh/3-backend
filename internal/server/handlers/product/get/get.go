package get_products

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ProductCardsProvider interface {
	ProductCards(ctx context.Context, req dtos.GetProductsRequest) ([]models.ProductCard, error)
}

// New godoc
//
//	@Summary		get products list
//	@Description	get products list
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		true	"page for pagination"
//	@Param			popularity	query		bool	false	"enable filtering by popularity"
//	@Param			price		query		int		false	"enable filtering by price"
//	@Param			category_id	query		int		false	"products category id"
//	@Param			search		query		string	false	"search patern"
//	@Success		200			{object}	dtos.GetProductsResponse
//	@Failure		404			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/products [get]
func New(validator *validator.Validate, productCardsProvider ProductCardsProvider) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.product.get.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var page int
		var err error

		pageStr := r.URL.Query().Get("page")
		if pageStr == "" {
			page = 0
		} else {
			page, err = strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				page = 0
			} else {
				page -= 1
			}
		}

		popularity := r.URL.Query().Has("popularity")

		var price int
		priceStr := r.URL.Query().Get("price")
		if priceStr == "" {
			price = -1
		} else {
			price, err = strconv.Atoi(priceStr)
			if err != nil {
				price = -1
			}
		}

		var categoryId int64
		categoryStr := r.URL.Query().Get("category_id")
		if categoryStr == "" {
			categoryId = -1
		} else {
			categoryId, err = strconv.ParseInt(categoryStr, 10, 64)
			if err != nil {
				categoryId = -1
			}
		}

		search := r.URL.Query().Get("search")
		search = strings.ReplaceAll(search, "_", " ")

		req := dtos.GetProductsRequest{
			Page:       page,
			Popularity: popularity,
			Price:      price,
			CategoryID: categoryId,
			Search:     search,
		}

		products, err := productCardsProvider.ProductCards(ctx, req)
		if err != nil {
			if errors.Is(err, errs.ErrProductNotFound) {
				log.Error(errs.ErrProductNotFound.Error())
				return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to get product cards", logger.Err(err))
			return response.Error("failed to get product cards", http.StatusInternalServerError)
		}

		resp := dtos.GetProductsResponse{
			Products: make([]dtos.Product, 0, len(products)),
		}
		for _, v := range products {
			product := dtos.Product{
				ID:                v.ID,
				Name:              v.Name,
				Price:             v.Price,
				ImageUrl:          v.ImageUrl,
				Discount:          v.Discount,
				DiscountExpiresAt: v.DiscountExpiresAt,
			}
			resp.Products = append(resp.Products, product)
		}

		render.JSON(w, r, resp)

		return nil
	}
}
