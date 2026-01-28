package login

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

type Loginer interface {
	Login(ctx context.Context, req dtos.LoginRequest) (string, string, error)
}

// New godoc
//
//	@Summary		login user
//	@Description	login user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			email		body		string	true	"User email"	Format(email)
//	@Param			password	body		string	true	"User password"
//	@Success		201			{object}	dtos.LoginResponse
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		404			{object}	response.ErrorResponse
//	@Failure		424			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/auth/login [post]
func New(validator *validator.Validate, loginer Loginer) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.auth.login.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var req dtos.LoginRequest
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

		accessToken, refreshToken, err := loginer.Login(ctx, req)
		if err != nil {
			if errors.Is(err, errs.ErrUserNotFound) {
				log.Error("user not found")
				return response.Error(errs.ErrUserNotFound.Error(), http.StatusNotFound)
			}
			if errors.Is(err, errs.ErrEmailNotVerified) {
				log.Error("email not verified")
				return response.Error(errs.ErrEmailNotVerified.Error(), http.StatusFailedDependency)
			}

			log.Error("failed to login user", logger.Err(err))
			return response.Error("failed to login user", http.StatusInternalServerError)
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dtos.LoginResponse{
			AccessToken:  accessToken, // I know that this is bad, but I don't care about it. Maybe I will change it later
			RefreshToken: refreshToken,
		})

		return nil
	}
}
