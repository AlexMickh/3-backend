package product_router

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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ProductService interface {
	ProductById(ctx context.Context, id string) (*models.Product, error)
	ProductCards(ctx context.Context, req dtos.GetProductsRequest) ([]models.ProductCard, error)
}

type ProductRouter struct {
	productService ProductService
}

func New(productService ProductService) *ProductRouter {
	return &ProductRouter{
		productService: productService,
	}
}

func (p *ProductRouter) RegisterRoute(r *chi.Mux) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/", response.ErrorWrapper(p.Products))
		r.Get("/{id}", response.ErrorWrapper(p.ProductById))
	})
}

// Products godoc
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
func (p *ProductRouter) Products(w http.ResponseWriter, r *http.Request) error {
	const op = "router.produuct.Products"
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

	categoryId := r.URL.Query().Get("category_id")

	search := r.URL.Query().Get("search")
	search = strings.ReplaceAll(search, "_", " ")

	req := dtos.GetProductsRequest{
		Page:       page,
		Popularity: popularity,
		Price:      price,
		CategoryID: categoryId,
		Search:     search,
	}

	products, err := p.productService.ProductCards(ctx, req)
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

// ProductById godoc
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
func (p *ProductRouter) ProductById(w http.ResponseWriter, r *http.Request) error {
	const op = "router.product.ProductById"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	// id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	// if err != nil {
	// 	log.Error("failed to get id", logger.Err(err))
	// 	return response.Error("failed to get id", http.StatusBadRequest)
	// }
	id := r.PathValue("id")

	product, err := p.productService.ProductById(ctx, id)
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
			ID   string "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   product.Category.ID.String(),
			Name: product.Category.Name,
		},
	}
	for _, v := range product.ExistingSizes {
		resp.ExistingSizes = append(resp.ExistingSizes, string(v))
	}

	render.JSON(w, r, resp)

	return nil
}
