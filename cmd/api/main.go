package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/Timeobi/go-ecommerce/docs/swagger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/Timeobi/go-ecommerce/internal/config"
	"github.com/Timeobi/go-ecommerce/internal/handler"
	"github.com/Timeobi/go-ecommerce/internal/logger"
	authmw "github.com/Timeobi/go-ecommerce/internal/middleware"
	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
	"github.com/Timeobi/go-ecommerce/internal/service"
)

// @title           Go E-commerce API
// @version         1.0
// @description     REST API для учебного проекта интернет-магазина на Go.
// @description     Слоистая архитектура, JWT-аутентификация, транзакционное оформление заказов.

// @contact.name    Timeobi
// @contact.url     https://github.com/Timeobi

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                Введите токен в формате: Bearer {ваш JWT токен}

func main() {

	cfg, err := config.Load()
	if err != nil {
		panic("не удалось загрузить конфигурацию: " + err.Error())
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	handler.SetLogger(log)

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Error("не удалось открыть подключение к БД", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Error("не удалось подключиться к БД", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("подключение к БД установлено успешно")

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	cartRepo := repository.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo, productRepo)
	cartHandler := handler.NewCartHandler(cartService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	r := chi.NewRouter()

	// ВАЖНО: RequestID должен быть подключён ПЕРВЫМ — остальные middleware
	// (наш логгер, recovery) полагаются на то, что request_id уже в контексте.
	r.Use(chimiddleware.RequestID)
	r.Use(authmw.Recovery(log))
	r.Use(authmw.RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			_ = err
		}
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", categoryHandler.GetAllCategories)
		r.Get("/{id}", categoryHandler.GetCategoryByID)

		r.Group(func(r chi.Router) {
			r.Use(authmw.Auth(cfg.JWTSecret))
			r.Use(authmw.RequireRole(model.RoleAdmin))
			r.Post("/", categoryHandler.CreateCategory)
		})
	})

	r.Route("/products", func(r chi.Router) {
		r.Get("/", productHandler.GetAllProducts)
		r.Get("/{id}", productHandler.GetProductByID)

		r.Group(func(r chi.Router) {
			r.Use(authmw.Auth(cfg.JWTSecret))
			r.Use(authmw.RequireRole(model.RoleAdmin))
			r.Post("/", productHandler.CreateProduct)
			r.Put("/{id}", productHandler.UpdateProduct)
			r.Delete("/{id}", productHandler.DeleteProduct)
		})
	})

	r.Route("/cart", func(r chi.Router) {
		r.Use(authmw.Auth(cfg.JWTSecret))
		r.Get("/", cartHandler.GetCart)
		r.Post("/items", cartHandler.AddItem)
		r.Put("/items/{productID}", cartHandler.UpdateItemQuantity)
		r.Delete("/items/{productID}", cartHandler.RemoveItem)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(authmw.Auth(cfg.JWTSecret))
		r.Post("/", orderHandler.Checkout)
		r.Get("/", orderHandler.ListMyOrders)
		r.Get("/{id}", orderHandler.GetOrderByID)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("сервер запущен", slog.String("port", cfg.ServerPort))
	if err := srv.ListenAndServe(); err != nil {
		log.Error("сервер остановлен с ошибкой", slog.String("error", err.Error()))
		os.Exit(1)
	}

	loginLimiter := authmw.NewRateLimiter(1, 5) // 1 запрос/сек в среднем, всплеск до

	r.Route("/auth", func(r chi.Router) {
		r.With(loginLimiter.Limit).Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
	})
}
