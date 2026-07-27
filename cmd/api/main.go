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

	r := chi.NewRouter()

	// Встроенные middleware chi — применяются ко ВСЕМ запросам
	r.Use(middleware.Logger)    // логирует каждый запрос: метод, путь, статус, время выполнения
	r.Use(middleware.Recoverer) // перехватывает панику в хендлерах, чтобы сервер не падал целиком

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", categoryHandler.GetAllCategories)
		r.Post("/", categoryHandler.CreateCategory)
		r.Get("/{id}", categoryHandler.GetCategoryByID)
	})

	r.Route("/products", func(r chi.Router) {
		r.Get("/", productHandler.GetAllProducts)
		r.Post("/", productHandler.CreateProduct)
		r.Get("/{id}", productHandler.GetProductByID)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
	})

	fmt.Printf("Сервер запущен на :%s\n", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
