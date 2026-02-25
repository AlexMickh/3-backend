package server

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/AlexMickh/shop-backend/docs"
	"github.com/AlexMickh/shop-backend/internal/config"
	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/auth/login"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/auth/refresh"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/auth/register"
	cart_add "github.com/AlexMickh/shop-backend/internal/server/handlers/cart/add"
	clear_cart "github.com/AlexMickh/shop-backend/internal/server/handlers/cart/clear"
	delete_cart_item "github.com/AlexMickh/shop-backend/internal/server/handlers/cart/delete_item"
	get_cart "github.com/AlexMickh/shop-backend/internal/server/handlers/cart/get"
	create_category "github.com/AlexMickh/shop-backend/internal/server/handlers/category/create"
	delete_category "github.com/AlexMickh/shop-backend/internal/server/handlers/category/delete"
	get_categories "github.com/AlexMickh/shop-backend/internal/server/handlers/category/get"
	product_by_id "github.com/AlexMickh/shop-backend/internal/server/handlers/product/by_id"
	create_product "github.com/AlexMickh/shop-backend/internal/server/handlers/product/create"
	delete_product "github.com/AlexMickh/shop-backend/internal/server/handlers/product/delete"
	get_products "github.com/AlexMickh/shop-backend/internal/server/handlers/product/get"
	update_product "github.com/AlexMickh/shop-backend/internal/server/handlers/product/update"
	"github.com/AlexMickh/shop-backend/internal/server/handlers/user/verify"
	"github.com/AlexMickh/shop-backend/internal/server/middlewares"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
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
	authService AuthService,
	userService UserService,
	sessionService SessionService,
	categoryService CategoryService,
	productService ProductService,
	cartService CartService,
) (*Server, error) {
	const op = "server.New"

	r := chi.NewRouter()

	validator := validator.New()

	r.Use(middleware.RequestID)
	r.Use(logger.ChiMiddleware(ctx))
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", cfg.Addr)), //The url pointing to API definition
	))

	r.Get("/health-check", response.ErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		logger.FromCtx(r.Context()).Info("hello")
		w.WriteHeader(200)
		return nil
	}))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", response.ErrorWrapper(register.New(validator, authService)))
		r.Post("/login", response.ErrorWrapper(login.New(validator, authService)))
		r.Put("/refresh", response.ErrorWrapper(refresh.New(validator, sessionService)))
	})

	r.Route("/user", func(r chi.Router) {
		r.Get("/verify/{token}", response.ErrorWrapper(verify.New(userService)))
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin-auth", map[string]string{
			cfg.AdminLogin: cfg.AdminPassword,
		}))

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", response.ErrorWrapper(create_category.New(validator, categoryService)))
			r.Delete("/{id}", response.ErrorWrapper(delete_category.New(validator, categoryService)))
		})

		r.Route("/products", func(r chi.Router) {
			r.Post("/", response.ErrorWrapper(create_product.New(validator, productService)))
			r.Patch("/{id}", response.ErrorWrapper(update_product.New(validator, productService)))
			r.Delete("/{id}", response.ErrorWrapper(delete_product.New(productService)))
		})
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", response.ErrorWrapper(get_categories.New(validator, categoryService)))
	})

	r.Route("/products", func(r chi.Router) {
		r.Get("/{id}", response.ErrorWrapper(product_by_id.New(validator, productService)))
		r.Get("/", response.ErrorWrapper(get_products.New(validator, productService)))
	})

	r.Route("/carts", func(r chi.Router) {
		r.Use(middlewares.Login(sessionService))

		r.Post("/add/{product_id}", response.ErrorWrapper(cart_add.New(cartService)))
		r.Get("/", response.ErrorWrapper(get_cart.New(cartService)))
		r.Delete("/{item_id}", response.ErrorWrapper(delete_cart_item.New(cartService)))
		r.Delete("/", response.ErrorWrapper(clear_cart.New(cartService)))
	})

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
