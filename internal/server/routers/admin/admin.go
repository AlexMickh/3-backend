package admin_router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, req dtos.CreateCategoryRequest) (uuid.UUID, error)
	DeleteCategory(ctx context.Context, id string) error
}

type ProductService interface {
	CreateProduct(ctx context.Context, req dtos.CreateProductRequest) (uuid.UUID, error)
	UpdateProduct(ctx context.Context, req *dtos.UpdateProductRequest) error
	DeleteProduct(ctx context.Context, id string) error
}

type AdminRouter struct {
	login           string
	password        string
	categoryService CategoryService
	productService  ProductService
}

var ErrNothingToUpdate = errors.New("nothing to update")

func New(login, password string, categoryService CategoryService, productService ProductService) *AdminRouter {
	return &AdminRouter{
		login:           login,
		password:        password,
		categoryService: categoryService,
		productService:  productService,
	}
}

func (a *AdminRouter) RegisterRoute(r *chi.Mux) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin-auth", map[string]string{
			a.login: a.password,
		}))

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", response.ErrorWrapper(a.CreateCategory))
			r.Delete("/{id}", response.ErrorWrapper(a.DeleteCategory))
		})

		r.Route("/products", func(r chi.Router) {
			r.Post("/", response.ErrorWrapper(a.CreateProduct))
			r.Patch("/{id}", response.ErrorWrapper(a.UpdateProduct))
			r.Delete("/{id}", response.ErrorWrapper(a.DeleteProduct))
		})
	})
}

// Create godoc
//
//	@Summary		create new category
//	@Description	create new category
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			name	body		string	true	"category name"
//	@Success		201		{object}	dtos.CreateCategoryResponse
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/categories [post]
func (a *AdminRouter) CreateCategory(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.admin.CreateCategory"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	var req dtos.CreateCategoryRequest
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Error("failed to decode request", logger.Err(err))
		return response.Error("failed to decode request", http.StatusBadRequest)
	}
	defer r.Body.Close()

	// if err = validator.Struct(&req); err != nil {
	// 	log.Error("failed to validate request", logger.Err(err))
	// 	return response.Error("failed to validate request", http.StatusBadRequest)
	// }

	id, err := a.categoryService.CreateCategory(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrCategoryAlreadyExists) {
			log.Error(errs.ErrCategoryAlreadyExists.Error())
			return response.Error(errs.ErrCategoryAlreadyExists.Error(), http.StatusConflict)
		}

		log.Error("failed to create category", logger.Err(err))
		return response.Error("failed to create category", http.StatusInternalServerError)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.CreateCategoryResponse{
		ID: id.String(),
	})

	return nil
}

// DeleteCategory godoc
//
//	@Summary		delete category
//	@Description	delete category
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"category id"
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/categories/{id} [delete]
func (a *AdminRouter) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.admin.DeleteCattegory"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	id := r.PathValue("id")

	err := a.categoryService.DeleteCategory(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrCategoryNotFound) {
			log.Error(errs.ErrCategoryNotFound.Error())
			return response.Error(errs.ErrCategoryNotFound.Error(), http.StatusNotFound)
		}

		log.Error("failed to delete category", logger.Err(err))
		return response.Error("failed to delete category", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// CreateProduct godoc
//
//	@Summary		create new product
//	@Description	create new product
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			name			formData	string	true	"product name"
//	@Param			description		formData	string	true	"product description"
//	@Param			price			formData	integer	true	"product proce"
//	@Param			category_id		formData	string	true	"product category id"
//	@Param			quantity		formData	integer	true	"product quantity"
//	@Param			existing_sizes	formData	string	true	"product sizes"
//	@Param			image			formData	file	true	"product image"
//	@Success		201				{object}	dtos.CreateCategoryResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		409				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/products [post]
func (a *AdminRouter) CreateProduct(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.admin.CreateProduct"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	var req dtos.CreateProductRequest
	err := parseCreateProductForm(r, &req)
	if err != nil {
		if errors.Is(err, errs.ErrUnsupportedImageType) {
			log.Error(errs.ErrUnsupportedImageType.Error())
			return response.Error(errs.ErrUnsupportedImageType.Error(), http.StatusBadRequest)
		}

		log.Error("failed to parse form", logger.Err(err))
		return response.Error("failed to parse form", http.StatusBadRequest)
	}

	// if err = validator.Struct(&req); err != nil {
	// 	log.Error("failed to validate request", logger.Err(err))
	// 	return response.Error("failed to validate request", http.StatusBadRequest)
	// }

	productID, err := a.productService.CreateProduct(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrProductAlreadyExists) {
			log.Error(errs.ErrProductAlreadyExists.Error())
			return response.Error(errs.ErrProductAlreadyExists.Error(), http.StatusConflict)
		}

		log.Error("failed to create product", logger.Err(err))
		return response.Error("failed to create product", http.StatusInternalServerError)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.CreateProductResponse{
		ID: productID.String(),
	})

	return nil
}

// UpdateProduct godoc
//
//	@Summary		update product
//	@Description	update product
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id					path		int		true	"product id"
//	@Param			name				formData	string	false	"new product name"
//	@Param			description			formData	string	false	"new product description"
//	@Param			price				formData	integer	false	"new product proce"
//	@Param			quantity			formData	integer	false	"new product quantity"
//	@Param			existing_sizes		formData	string	false	"new product sizes"
//	@Param			image				formData	file	false	"new product image"
//	@Param			discount			formData	int		false	"new product discount"
//	@Param			discount_expires_at	formData	string	false	"new product discount expires at (only if discount exists)"
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/products/{id} [patch]
func (a *AdminRouter) UpdateProduct(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.admin.UpdateProduct"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	req := new(dtos.UpdateProductRequest)

	req.ID = r.PathValue("id")
	if req.ID == "" {
		log.Error("id is empty")
		return response.Error("product id is required", http.StatusBadRequest)
	}

	err := parseUpdateProductForm(r, req)
	if err != nil {
		if errors.Is(err, ErrNothingToUpdate) {
			log.Error("nothing to update")
			return response.Error(ErrNothingToUpdate.Error(), http.StatusBadRequest)
		}

		log.Error("failed to parse form", logger.Err(err))
		return response.Error("failed to perse request", http.StatusBadRequest)
	}

	// if err = validator.Struct(req); err != nil {
	// 	log.Error("failed to validate request", logger.Err(err))
	// 	return response.Error("failed to validate request", http.StatusBadRequest)
	// }

	if err = a.productService.UpdateProduct(ctx, req); err != nil {
		if errors.Is(err, errs.ErrProductNotFound) {
			log.Error("product not found")
			return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
		}

		log.Error("failed to update product", logger.Err(err))
		return response.Error("failed to update product", http.StatusInternalServerError)
	}

	render.Status(r, http.StatusNoContent)

	return nil
}

// DeleteProduct godoc
//
//	@Summary		delete product
//	@Description	delete product
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"product id"
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/products/{id} [delete]
func (a *AdminRouter) DeleteProduct(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.admin.DeleteProduct"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	id := r.PathValue("id")

	if err := a.productService.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, errs.ErrProductNotFound) {
			log.Error(errs.ErrProductNotFound.Error())
			return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
		}

		log.Error("failed to delete product", logger.Err(err))
		return response.Error("failed to delete product", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func parseCreateProductForm(r *http.Request, req *dtos.CreateProductRequest) error {
	const op = "routers.admin.parseCreateProductForm"

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.Name = r.FormValue("name")
	req.Description = r.FormValue("description")

	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	req.Price = price

	req.CategoryID = r.FormValue("category_id")

	req.Quantity, err = strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.ExistingSizes = strings.Split(r.FormValue("existing_sizes"), " ")

	var header *multipart.FileHeader
	req.Image, header, err = r.FormFile("image")
	if err != nil {
		return fmt.Errorf("1111%s: %w", op, err)
	}

	arr := strings.Split(header.Filename, ".")
	if arr[len(arr)-1] != "png" {
		return fmt.Errorf("%s: %w", op, errs.ErrUnsupportedImageType)
	}

	return nil
}

func parseUpdateProductForm(r *http.Request, req *dtos.UpdateProductRequest) error {
	const op = "routers.admin.parseUpdateProductForm"

	hasSomething := false

	req.Name = r.FormValue("name")
	if req.Name != "" {
		hasSomething = true
	}

	req.Description = r.FormValue("description")
	if req.Description != "" && !hasSomething {
		hasSomething = true
	}

	var err error
	req.Price, err = strconv.Atoi(r.FormValue("price"))
	if err == nil && !hasSomething {
		hasSomething = true
	}

	req.Quantity, err = strconv.Atoi(r.FormValue("quantity"))
	if err == nil && !hasSomething {
		hasSomething = true
	}

	req.ExistingSizes = strings.Split(r.FormValue("existing_sizes"), " ")
	if req.ExistingSizes != nil && !hasSomething {
		hasSomething = true
	}

	var header *multipart.FileHeader
	req.Image, header, err = r.FormFile("image")
	if err == nil {
		arr := strings.Split(header.Filename, ".")
		if arr[len(arr)-1] != "png" {
			return fmt.Errorf("%s: %w", op, errs.ErrUnsupportedImageType)
		}

		hasSomething = true
	}

	req.Price, err = strconv.Atoi(r.FormValue("discount"))
	if err == nil && !hasSomething {
		hasSomething = true
	}

	discountExpiresAt := r.FormValue("discount_expires_at")
	if discountExpiresAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", discountExpiresAt)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		req.DiscountExpiresAt = &t
		hasSomething = true
	}

	if !hasSomething {
		return fmt.Errorf("%s: %w", op, ErrNothingToUpdate)
	}

	return nil
}
