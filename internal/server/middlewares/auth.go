package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
)

type TokenValidator interface {
	ValidateJwt(token string) (int64, error)
}

const UserIdKey = "user_id"

func Login(tokenValidator TokenValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(response.ErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
			const op = "middlewares.auth.Login"
			ctx := r.Context()
			log := logger.FromCtx(ctx).With(slog.String("op", op))

			header := r.Header.Get("Authorization")
			if header == "" {
				log.Error("header is empty")
				return response.Error("authorization header is empty", http.StatusUnauthorized)
			}

			content := strings.Split(header, " ")
			if len(content) != 2 && content[0] != "Bearer" {
				log.Error("bad format")
				return response.Error("bad format", http.StatusUnauthorized)
			}

			userID, err := tokenValidator.ValidateJwt(content[1])
			if err != nil {
				log.Error("failed to validate token")
				return response.Error("failed to validate token", http.StatusUnauthorized)
			}

			ctx = context.WithValue(ctx, UserIdKey, userID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

			return nil
		}))
	}
}
