package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/AlexMickh/shop-backend/docs"
	"github.com/AlexMickh/shop-backend/internal/config"
	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/internal/server/routers"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	srv        *http.Server
	fileServer *http.Server
}

type AuthService interface {
	Register(ctx context.Context, req dtos.RegisterDto) (int64, error)
	Login(ctx context.Context, req dtos.LoginRequest) (string, string, error)
}

type UserService interface {
	VerifyEmail(ctx context.Context, token string) error
}

type SessionService interface {
	Refresh(req dtos.RefreshRequest) (string, string, error)
	ValidateJwt(token string) (int64, error)
}

type CategoryService interface {
	CreateCategory(ctx context.Context, req dtos.CreateCategoryRequest) (int64, error)
	DeleteCategory(ctx context.Context, id int64) error
	AllCategories(ctx context.Context) ([]models.Category, error)
}

type ProductService interface {
	CreateProduct(ctx context.Context, req dtos.CreateProductRequest) (int64, error)
	ProductById(ctx context.Context, id int64) (*models.Product, error)
	ProductCards(ctx context.Context, req dtos.GetProductsRequest) ([]models.ProductCard, error)
	UpdateProduct(ctx context.Context, req *dtos.UpdateProductRequest) error
	DeleteProduct(ctx context.Context, id int64) error
}

type CartService interface {
	AddToCart(ctx context.Context, userId, productId int64) (int64, error)
	Cart(ctx context.Context, userId int64) (models.Cart, error)
	DeleteItem(ctx context.Context, userId, productId int64) error
	Clear(ctx context.Context, userId int64) error
}

// @title						Three Api
// @version					1.0
// @description				Your API description
// @securityDefinitions.apikey	UserAuth
// @in							header
// @name						Authorization
// @securityDefinitions.basic	AdminAuth
func New(
	ctx context.Context,
	cfg config.ServerConfig,
	routers []routers.Router,
) (*Server, error) {
	const op = "server.New"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.ChiMiddleware(ctx))
	r.Use(middleware.Recoverer)

	addr := fmt.Sprintf("http://%s/docs/doc.json", cfg.Addr)
	if strings.Split(cfg.Addr, ":")[0] == "0.0.0.0" {
		addr = "http://127.0.0.1/docs/doc.json"
	}

	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL(addr), //The url pointing to API definition
	))

	for _, router := range routers {
		router.RegisterRoute(r)
	}

	return &Server{
		srv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      r,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		fileServer: &http.Server{
			Addr:         cfg.FileServerAddr,
			Handler:      http.FileServer(http.Dir("./public")),
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	const op = "server.Run"

	go s.fileServer.ListenAndServe()

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

	if err := s.fileServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Server) Addr() string {
	return s.srv.Addr
}
