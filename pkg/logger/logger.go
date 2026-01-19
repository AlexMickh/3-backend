package logger

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type key string

var (
	Key       = key("logger")
	RequestID = "request_id"
)

func New(env string, w io.Writer) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.Default()
	}

	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func ContextWithLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, Key, log)
}

func FromCtx(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(Key).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return log
}

func ChiMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := FromCtx(ctx)

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			guid := middleware.GetReqID(r.Context())
			//nolint:staticcheck
			ctx = context.WithValue(r.Context(), RequestID, guid)
			log = log.With(slog.String("request_id", guid))
			log.Info(
				"new request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remove_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				FromCtx(ctx).Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			ctx = context.WithValue(ctx, Key, log)
			r = r.WithContext(ctx)

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
