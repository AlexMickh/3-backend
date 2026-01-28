package refresh

import (
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

type Refresher interface {
	Refresh(req dtos.RefreshRequest) (string, string, error)
}

// New godoc
//
//	@Summary		generate new pare of tokens
//	@Description	generate new pare of tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh_token	body		string	true	"Refresh token"
//	@Success		201				{object}	dtos.RefreshResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/auth/refresh [put]
func New(validator *validator.Validate, refresher Refresher) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.auth.refresh.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var req dtos.RefreshRequest
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", logger.Err(err))
			return response.Error("failed to decode request", http.StatusBadRequest)
		}

		if err = validator.Struct(&req); err != nil {
			log.Error("failed to validate request", logger.Err(err))
			return response.Error("failed to validate request", http.StatusBadRequest)
		}

		accessToken, refreshToken, err := refresher.Refresh(req)
		if err != nil {
			if errors.Is(err, errs.ErrSessionNotFound) {
				log.Error(errs.ErrSessionNotFound.Error())
				return response.Error(errs.ErrSessionNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to refresh", logger.Err(err))
			response.Error("failed to refresh", http.StatusInternalServerError)
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dtos.RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})

		return nil
	}
}
