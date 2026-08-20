package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Timeobi/go-ecommerce/internal/config"
	"github.com/Timeobi/go-ecommerce/internal/handler"
	authmw "github.com/Timeobi/go-ecommerce/internal/middleware"
	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
	"github.com/Timeobi/go-ecommerce/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("не удалось загрузить конфигурацию: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("не удалось открыть подключение к БД: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	fmt.Println("Подключение к БД установлено успешно")

	// Categories
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// Products
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// Auth
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Публичные маршруты аутентификации
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Категории — чтение публичное, изменение — только для admin
	r.Route("/categories", func(r chi.Router) {
		r.Get("/", categoryHandler.GetAllCategories)
		r.Get("/{id}", categoryHandler.GetCategoryByID)

		r.Group(func(r chi.Router) {
			r.Use(authmw.Auth(cfg.JWTSecret))
			r.Use(authmw.RequireRole(model.RoleAdmin))
			r.Post("/", categoryHandler.CreateCategory)
		})
	})

	// Товары — чтение публичное, изменение — только для admin
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

	fmt.Printf("Сервер запущен на :%s\n", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
