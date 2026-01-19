package register

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Registerer interface {
	Register(ctx context.Context, req dtos.RegisterDto) (string, error)
}

func New(validator *validator.Validate, registerer Registerer) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.auth.register.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var req dtos.RegisterDto
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", logger.Err(err))
			return response.Error("failed to decode request body", http.StatusBadRequest)
		}
		defer r.Body.Close()

		if err = validator.Struct(&req); err != nil {
			log.Error("failed to validate request", logger.Err(err))
			return response.Error("failed to validate request", http.StatusBadRequest)
		}

		userID, err := registerer.Register(ctx, req)
		if err != nil {
			if errors.Is(err, errs.ErrUserAlreadyExists) {
				log.Error("user already exists")
				return response.Error(errs.ErrUserAlreadyExists.Error(), http.StatusConflict)
			}

			log.Error("failed to register user", logger.Err(err))
			return response.Error("failed to register user", http.StatusInternalServerError)
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dtos.ToRegisterResponse(userID))

		return nil
	}
}
