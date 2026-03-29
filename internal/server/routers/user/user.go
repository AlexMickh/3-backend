package user_router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/chi/v5"
)

type UserService interface {
	VerifyEmail(ctx context.Context, token string) error
}

type UserRouter struct {
	userService UserService
}

func New(userService UserService) *UserRouter {
	return &UserRouter{
		userService: userService,
	}
}

func (u *UserRouter) RegisterRoute(r *chi.Mux) {
	r.Get("/users/verify/{token}", response.ErrorWrapper(u.VerifyEmail))
}

func (u *UserRouter) VerifyEmail(w http.ResponseWriter, r *http.Request) error {
	const op = "router.user.VerifyEmail"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	token := r.PathValue("token")
	if token == "" {
		log.Error("token is empty")
		return response.Error("token is required", http.StatusBadRequest)
	}

	err := u.userService.VerifyEmail(ctx, token)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			log.Error("token not found")
			return response.Error("user not found", http.StatusNotFound)
		}

		log.Error("failed to verify token", logger.Err(err))
		return response.Error("failed to verify token", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
