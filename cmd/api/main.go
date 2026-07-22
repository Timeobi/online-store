package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрируем драйвер pgx для database/sql

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

	// Проверяем, что соединение реально рабочее (Open не гарантирует это сам по себе)
	if err := db.Ping(); err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	fmt.Println("Подключение к БД установлено успешно")

	// Собираем слои воедино (это называется Dependency Injection —
	// каждый слой получает свои зависимости через конструктор)
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			categoryHandler.GetAllCategories(w, r)
		case http.MethodPost:
			categoryHandler.CreateCategory(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/categories/", categoryHandler.GetCategoryByID)

	fmt.Printf("Сервер запущен на :%s\n", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
