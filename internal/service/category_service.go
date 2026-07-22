package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
)

var ErrCategoryNotFound = errors.New("category not found")

// CategoryService содержит бизнес-логику для работы с категориями.
type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// CreateCategory валидирует входные данные и создаёт новую категорию.
func (s *CategoryService) CreateCategory(ctx context.Context, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}

	slug := generateSlug(name)

	return s.repo.Create(ctx, name, slug)
}

// GetAllCategories возвращает все категории.
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	return s.repo.GetAll(ctx)
}

// GetCategoryByID возвращает категорию по ID или ErrCategoryNotFound, если не найдена.
func (s *CategoryService) GetCategoryByID(ctx context.Context, id int) (*model.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}
	return category, nil
}

// generateSlug — простая генерация slug из названия (заменяем пробелы на дефисы, приводим к нижнему регистру).
// На следующих этапах можем улучшить (транслитерация кириллицы и т.д.)
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}
