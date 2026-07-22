package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

// CategoryRepository отвечает за работу с таблицей categories в БД.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository — конструктор репозитория.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create добавляет новую категорию в БД и возвращает её с заполненным ID и CreatedAt.
func (r *CategoryRepository) Create(ctx context.Context, name, slug string) (*model.Category, error) {
	query := `
		INSERT INTO categories (name, slug)
		VALUES ($1, $2)
		RETURNING id, name, slug, created_at
	`

	var c model.Category
	err := r.db.QueryRowContext(ctx, query, name, slug).Scan(
		&c.ID, &c.Name, &c.Slug, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create category: %w", err)
	}

	return &c, nil
}

// GetAll возвращает список всех категорий.
func (r *CategoryRepository) GetAll(ctx context.Context) ([]model.Category, error) {
	query := `SELECT id, name, slug, created_at FROM categories ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: get all categories: %w", err)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan category row: %w", err)
		}
		categories = append(categories, c)
	}

	// Проверяем ошибку, которая могла возникнуть во время итерации (не только в Scan)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}

	return categories, nil
}

// GetByID возвращает одну категорию по её ID.
func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*model.Category, error) {
	query := `SELECT id, name, slug, created_at FROM categories WHERE id = $1`

	var c model.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // категория не найдена — не ошибка, просто nil
		}
		return nil, fmt.Errorf("repository: get category by id: %w", err)
	}

	return &c, nil
}
