package verify

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
)

type EmailVerifier interface {
	VerifyEmail(ctx context.Context, token string) error
}

func New(emailVerifier EmailVerifier) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.user.verify.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		token := r.PathValue("token")
		if token == "" {
			log.Error("token is empty")
			return response.Error("token is required", http.StatusBadRequest)
		}

		err := emailVerifier.VerifyEmail(ctx, token)
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
}
