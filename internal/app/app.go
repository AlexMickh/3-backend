package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/AlexMickh/shop-backend/internal/config"
	file_storage "github.com/AlexMickh/shop-backend/internal/file_storage/fs"
	"github.com/AlexMickh/shop-backend/internal/models"
	session_repository "github.com/AlexMickh/shop-backend/internal/repository/inmemory/session"
	cart_repository "github.com/AlexMickh/shop-backend/internal/repository/postgres/cart"
	category_repository "github.com/AlexMickh/shop-backend/internal/repository/postgres/category"
	product_repository "github.com/AlexMickh/shop-backend/internal/repository/postgres/product"
	token_repository "github.com/AlexMickh/shop-backend/internal/repository/postgres/token"
	user_repository "github.com/AlexMickh/shop-backend/internal/repository/postgres/user"
	"github.com/AlexMickh/shop-backend/internal/server"
	"github.com/AlexMickh/shop-backend/internal/server/routers"
	admin_router "github.com/AlexMickh/shop-backend/internal/server/routers/admin"
	auth_router "github.com/AlexMickh/shop-backend/internal/server/routers/auth"
	cart_router "github.com/AlexMickh/shop-backend/internal/server/routers/cart"
	category_router "github.com/AlexMickh/shop-backend/internal/server/routers/category"
	product_router "github.com/AlexMickh/shop-backend/internal/server/routers/product"
	user_router "github.com/AlexMickh/shop-backend/internal/server/routers/user"
	auth_service "github.com/AlexMickh/shop-backend/internal/services/auth"
	cart_service "github.com/AlexMickh/shop-backend/internal/services/cart"
	category_service "github.com/AlexMickh/shop-backend/internal/services/category"
	product_service "github.com/AlexMickh/shop-backend/internal/services/product"
	session_service "github.com/AlexMickh/shop-backend/internal/services/session"
	token_service "github.com/AlexMickh/shop-backend/internal/services/token"
	user_service "github.com/AlexMickh/shop-backend/internal/services/user"
	"github.com/AlexMickh/shop-backend/pkg/cash"
	"github.com/AlexMickh/shop-backend/pkg/clients/postgresql"
	"github.com/AlexMickh/shop-backend/pkg/email"
	"github.com/AlexMickh/shop-backend/pkg/jwt"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	db     *pgxpool.Pool
	server *server.Server
	cfg    *config.Config
}

func New(ctx context.Context, cfg *config.Config) *App {
	const op = "app.New"

	log := logger.FromCtx(ctx).With(slog.String("op", op))

	log.Info("initing cash")
	sessionCash := cash.New[string, models.Session](ctx, cfg.Jwt.RefreshTokenTtl)
	sessionRepository := session_repository.New(sessionCash)

	log.Info("initing file storage")
	fileStorage, err := file_storage.New("./public", cfg.Server.FileServerAddr)
	if err != nil {
		log.Error("failed to init file storage", logger.Err(err))
		os.Exit(1)
	}

	log.Info("initing postgres")
	db, err := postgresql.New(
		ctx,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.MinPools,
		cfg.DB.MaxPools,
	)
	if err != nil {
		log.Error("failed to init postgres")
		os.Exit(1)
	}

	userRepository := user_repository.New(db)
	tokenRepository := token_repository.New(db)
	categoryRepository := category_repository.New(db)
	productRepository := product_repository.New(db)
	cartRepository := cart_repository.New(db)

	log.Info("initing service layer")

	validator := validator.New()

	tokenService := token_service.New(tokenRepository, cfg.Tokens.VerifyEmailTokenTtl)
	userService := user_service.New(userRepository, tokenService)
	jwtManager := jwt.New(cfg.Jwt.Secret, cfg.Jwt.AccessTokenTtl)
	sessionService := session_service.New(sessionRepository, jwtManager, cfg.Jwt.RefreshTokenTtl, validator)
	categoryService := category_service.New(categoryRepository, validator)
	productService := product_service.New(productRepository, fileStorage, validator)
	cartService := cart_service.New(cartRepository)

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

	authService := auth_service.New(userService, tokenService, emailQueue, sessionService, validator)

	log.Info("init server")

	authRouter := auth_router.New(authService, sessionService, cfg.Jwt.RefreshTokenTtl)
	userRouter := user_router.New(userService)
	categoryRouter := category_router.New(categoryService)
	productRouter := product_router.New(productService)
	cartRouter := cart_router.New(cartService, sessionService)
	adminRouter := admin_router.New(
		cfg.Server.AdminLogin,
		cfg.Server.AdminPassword,
		categoryService,
		productService,
	)

	server, err := server.New(
		ctx,
		cfg.Server,
		[]routers.Router{authRouter, userRouter, categoryRouter, productRouter, cartRouter, adminRouter},
	)
	if err != nil {
		log.Error("failed to init server", logger.Err(err))
		os.Exit(1)
	}

	return &App{
		db:     db,
		server: server,
		cfg:    cfg,
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
