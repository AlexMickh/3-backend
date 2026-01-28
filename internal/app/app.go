package app

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/AlexMickh/shop-backend/internal/config"
	"github.com/AlexMickh/shop-backend/internal/lib/jwt"
	"github.com/AlexMickh/shop-backend/internal/models"
	session_repository "github.com/AlexMickh/shop-backend/internal/repository/inmemory/session"
	category_repository "github.com/AlexMickh/shop-backend/internal/repository/sqlite/category"
	token_repository "github.com/AlexMickh/shop-backend/internal/repository/sqlite/token"
	user_repository "github.com/AlexMickh/shop-backend/internal/repository/sqlite/user"
	"github.com/AlexMickh/shop-backend/internal/server"
	auth_service "github.com/AlexMickh/shop-backend/internal/services/auth"
	category_service "github.com/AlexMickh/shop-backend/internal/services/category"
	session_service "github.com/AlexMickh/shop-backend/internal/services/session"
	token_service "github.com/AlexMickh/shop-backend/internal/services/token"
	user_service "github.com/AlexMickh/shop-backend/internal/services/user"
	"github.com/AlexMickh/shop-backend/pkg/cash"
	sqlite_client "github.com/AlexMickh/shop-backend/pkg/clients/sqlite"
	"github.com/AlexMickh/shop-backend/pkg/email"
	"github.com/AlexMickh/shop-backend/pkg/logger"
)

type App struct {
	db     *sql.DB
	server *server.Server
}

func New(ctx context.Context, cfg *config.Config) *App {
	const op = "app.New"

	log := logger.FromCtx(ctx).With(slog.String("op", op))

	log.Info("initing sqlite")
	db, err := sqlite_client.New(ctx, cfg.DB.File)
	if err != nil {
		log.Error("failed to init sqlite", logger.Err(err))
		os.Exit(1)
	}

	userRepository := user_repository.New(db)
	tokenRepository := token_repository.New(db)
	categoryRepository := category_repository.New(db)

	log.Info("initing cash")
	sessionCash := cash.New[string, models.Session](ctx, cfg.Jwt.RefreshTokenTtl)
	sessionRepository := session_repository.New(sessionCash)

	log.Info("initing service layer")
	tokenService := token_service.New(tokenRepository, cfg.Tokens.VerifyEmailTokenTtl)
	userService := user_service.New(userRepository)
	jwtManager := jwt.New(cfg.Jwt.Secret, cfg.Jwt.AccessTokenTtl)
	sessionService := session_service.New(sessionRepository, jwtManager, cfg.Jwt.RefreshTokenTtl)
	categoryService := category_service.New(categoryRepository)

	emailQueue, err := email.New(ctx, email.EmailConfig{
		Host:     cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		FromAddr: cfg.Mail.FromAddr,
		Password: cfg.Mail.Password,
	})
	if err != nil {
		log.Error("failed to init mailer", logger.Err(err))
		os.Exit(1)
	}

	authService := auth_service.New(userService, tokenService, emailQueue, sessionService)

	log.Info("init server")
	server, err := server.New(ctx, cfg.Server, authService, userService, sessionService, categoryService)
	if err != nil {
		log.Error("failed to init server", logger.Err(err))
		os.Exit(1)
	}

	return &App{
		db:     db,
		server: server,
	}
}

func (a *App) Run(ctx context.Context) {
	const op = "app.Run"

	log := logger.FromCtx(ctx).With(slog.String("op", op))

	go func() {
		if err := a.server.Run(ctx); err != nil {
			log.Error("failed to start server", logger.Err(err))
			os.Exit(1)
		}
	}()

	log.Info("server started", slog.String("addr", a.server.Addr()))
}

func (a *App) GracefulStop(ctx context.Context) {
	a.server.GracefulStop(ctx)
	a.db.Close()
}
