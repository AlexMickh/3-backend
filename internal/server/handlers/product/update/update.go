package update_product

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
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

var ErrNothingToUpdate = errors.New("nothing to update")

type ProductUpdater interface {
	UpdateProduct(ctx context.Context, req *dtos.UpdateProductRequest) error
}

// New godoc
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
func New(validator *validator.Validate, productUpdater ProductUpdater) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.product.update.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		productId := r.PathValue("id")
		if productId == "" {
			log.Error("id is empty")
			return response.Error("product id is required", http.StatusBadRequest)
		}

		req := new(dtos.UpdateProductRequest)
		var err error
		req.ID, err = strconv.ParseInt(productId, 10, 64)
		if err != nil {
			log.Error("failed to parse id", logger.Err(err))
			return response.Error("id is not a valid id", http.StatusBadRequest)
		}

		if err = parseForm(r, req); err != nil {
			if errors.Is(err, ErrNothingToUpdate) {
				log.Error("nothing to update")
				return response.Error(ErrNothingToUpdate.Error(), http.StatusBadRequest)
			}

			log.Error("failed to parse form", logger.Err(err))
			return response.Error("failed to perse request", http.StatusBadRequest)
		}

		if err = validator.Struct(req); err != nil {
			log.Error("failed to validate request", logger.Err(err))
			return response.Error("failed to validate request", http.StatusBadRequest)
		}

		if err = productUpdater.UpdateProduct(ctx, req); err != nil {
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
}

func parseForm(r *http.Request, req *dtos.UpdateProductRequest) error {
	const op = "handlers.product.create.parseForm"

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
