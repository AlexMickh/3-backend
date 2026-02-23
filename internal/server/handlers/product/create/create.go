package create_product

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ProductCreater interface {
	CreateProduct(ctx context.Context, req dtos.CreateProductRequest) (int64, error)
}

// New godoc
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
func New(validator *validator.Validate, productCreater ProductCreater) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.product.create.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var req dtos.CreateProductRequest
		err := parseForm(r, &req)
		if err != nil {
			if errors.Is(err, errs.ErrUnsupportedImageType) {
				log.Error(errs.ErrUnsupportedImageType.Error())
				return response.Error(errs.ErrUnsupportedImageType.Error(), http.StatusBadRequest)
			}

			log.Error("failed to parse form", logger.Err(err))
			return response.Error("failed to parse form", http.StatusBadRequest)
		}

		if err = validator.Struct(&req); err != nil {
			log.Error("failed to validate request", logger.Err(err))
			return response.Error("failed to validate request", http.StatusBadRequest)
		}

		productID, err := productCreater.CreateProduct(ctx, req)
		if err != nil {
			if errors.Is(err, errs.ErrProductAlreadyExists) {
				log.Error(errs.ErrProductAlreadyExists.Error())
				return response.Error(errs.ErrProductAlreadyExists.Error(), http.StatusConflict)
			}

			log.Error("failed to create product", logger.Err(err))
			return response.Error("failed to create product", http.StatusInternalServerError)
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dtos.CreateCategoryResponse{
			ID: productID,
		})

		return nil
	}
}

func parseForm(r *http.Request, req *dtos.CreateProductRequest) error {
	const op = "handlers.product.create.parseForm"

	req.Name = r.FormValue("name")
	req.Description = r.FormValue("description")

	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	req.Price = price

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	req.CategoryID = categoryID

	req.Quantity, err = strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.ExistingSizes = strings.Split(r.FormValue("existing_sizes"), " ")

	var header *multipart.FileHeader
	req.Image, header, err = r.FormFile("image")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	arr := strings.Split(header.Filename, ".")
	if arr[len(arr)-1] != "png" {
		return fmt.Errorf("%s: %w", op, errs.ErrUnsupportedImageType)
	}

	return nil
}
