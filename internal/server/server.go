package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/config"
	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/auth/login"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/auth/register"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	srv *http.Server
}

type AuthService interface {
	Register(ctx context.Context, req dtos.RegisterDto) (int64, error)
	Login(ctx context.Context, req dtos.LoginRequest) (string, string, error)
}

func New(
	ctx context.Context,
	cfg config.ServerConfig,
	authService AuthService,
) (*Server, error) {
	const op = "server.New"

	r := chi.NewRouter()

	validator := validator.New()

	r.Use(middleware.RequestID)
	r.Use(logger.ChiMiddleware(ctx))
	r.Use(middleware.Recoverer)

	r.Get("/health-check", response.ErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		logger.FromCtx(r.Context()).Info("hello")
		w.WriteHeader(200)
		return nil
	}))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", response.ErrorWrapper(register.New(validator, authService)))
		r.Post("/login", response.ErrorWrapper(login.New(validator, authService)))
	})

	return &Server{
		srv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      r,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	const op = "server.Run"

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Server) GracefulStop(ctx context.Context) error {
	const op = "server.GracefulStop"

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Server) Addr() string {
	return s.srv.Addr
}
