package auth_router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/AlexMickh/shop-backend/pkg/utils/cookies"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req dtos.RegisterDto) (uuid.UUID, error)
	Login(ctx context.Context, req dtos.LoginRequest) (string, string, error)
}

type SessionService interface {
	Refresh(req dtos.RefreshRequest) (string, string, error)
}

type AuthRouter struct {
	authService     AuthService
	sessionService  SessionService
	refreshTokenTtl time.Duration
}

func New(authService AuthService, sessionService SessionService, refreshTokenTtl time.Duration) *AuthRouter {
	return &AuthRouter{
		authService:     authService,
		sessionService:  sessionService,
		refreshTokenTtl: refreshTokenTtl,
	}
}

func (a *AuthRouter) RegisterRoute(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", response.ErrorWrapper(a.Register))
		r.Post("/login", response.ErrorWrapper(a.Login))
		r.Put("/refresh", response.ErrorWrapper(a.Refresh))
	})
}

// Register godoc
//
//	@Summary		register user
//	@Description	register user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			email		body		string	true	"User email"	Format(email)
//	@Param			password	body		string	true	"User password"
//	@Success		201			{object}	dtos.RegisterResponse
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		409			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/auth/register [post]
func (a *AuthRouter) Register(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.auth.Register"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	var req dtos.RegisterDto
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Error("failed to decode request body", logger.Err(err))
		return response.Error("failed to decode request body", http.StatusBadRequest)
	}
	defer r.Body.Close()

	// if err = validator.Struct(&req); err != nil {
	// 	log.Error("failed to validate request", logger.Err(err))
	// 	return response.Error("failed to validate request", http.StatusBadRequest)
	// }

	userID, err := a.authService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			log.Error("user already exists")
			return response.Error(errs.ErrUserAlreadyExists.Error(), http.StatusConflict)
		}

		log.Error("failed to register user", logger.Err(err))
		return response.Error("failed to register user", http.StatusInternalServerError)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.RegisterResponse{
		ID: userID.String(),
	})

	return nil
}

// Login godoc
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
func (a *AuthRouter) Login(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.auth.Login"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	var req dtos.LoginRequest
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Error("failed to decode request body", logger.Err(err))
		return response.Error("failed to decode request body", http.StatusBadRequest)
	}
	defer r.Body.Close()

	// if err = validator.Struct(&req); err != nil {
	// 	log.Error("failed to validate request", logger.Err(err))
	// 	return response.Error("failed to validate request", http.StatusBadRequest)
	// }

	accessToken, refreshToken, err := a.authService.Login(ctx, req)
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

	cookies.Set(w, "refresh_token", refreshToken, a.refreshTokenTtl)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.LoginResponse{
		AccessToken: accessToken,
	})

	return nil
}

// Refresh godoc
//
//	@Summary		generate new pare of tokens
//	@Description	generate new pare of tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	dtos.RefreshResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/auth/refresh [put]
func (a *AuthRouter) Refresh(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.auth.Refresh"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Error("failed to get cookie", logger.Err(err))
		return response.Error("failed to get refresh token", http.StatusBadRequest)
	}

	req := dtos.RefreshRequest{
		RefreshToken: cookie.Value,
	}

	accessToken, refreshToken, err := a.sessionService.Refresh(req)
	if err != nil {
		if errors.Is(err, errs.ErrSessionNotFound) {
			log.Error(errs.ErrSessionNotFound.Error())
			return response.Error(errs.ErrSessionNotFound.Error(), http.StatusNotFound)
		}

		log.Error("failed to refresh", logger.Err(err))
		response.Error("failed to refresh", http.StatusInternalServerError)
	}

	cookies.Set(w, "refresh_token", refreshToken, a.refreshTokenTtl)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.RefreshResponse{
		AccessToken: accessToken,
	})

	return nil
}
