package service

import (
	"context"
	"testing"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/stretchr/testify/assert"
)

// mockCategoryRepository — поддельная реализация categoryRepository для тестов.
// Хранит данные просто в срезе в памяти — никакой реальной БД не нужно.
type mockCategoryRepository struct {
	categories []model.Category
	nextID     int
	// forceError позволяет тесту искусственно сымитировать ошибку БД —
	// полезно для проверки, что сервис правильно обрабатывает сбои репозитория.
	forceError error
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{nextID: 1}
}

func (m *mockCategoryRepository) Create(ctx context.Context, name, slug string) (*model.Category, error) {
	if m.forceError != nil {
		return nil, m.forceError
	}
	c := model.Category{ID: m.nextID, Name: name, Slug: slug}
	m.nextID++
	m.categories = append(m.categories, c)
	return &c, nil
}

func (m *mockCategoryRepository) GetAll(ctx context.Context) ([]model.Category, error) {
	if m.forceError != nil {
		return nil, m.forceError
	}
	return m.categories, nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id int) (*model.Category, error) {
	if m.forceError != nil {
		return nil, m.forceError
	}
	for _, c := range m.categories {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil // как и настоящий репозиторий — nil, nil означает "не найдено"
}

func TestCategoryService_CreateCategory(t *testing.T) {
	t.Run("успешное создание", func(t *testing.T) {
		repo := newMockCategoryRepository()
		svc := NewCategoryService(repo)

		category, err := svc.CreateCategory(context.Background(), "Электроника")

		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, "Электроника", category.Name)
		assert.Equal(t, "электроника", category.Slug)
	})

	t.Run("пустое название отклоняется", func(t *testing.T) {
		repo := newMockCategoryRepository()
		svc := NewCategoryService(repo)

		category, err := svc.CreateCategory(context.Background(), "   ")

		assert.Error(t, err)
		assert.Nil(t, category)
	})

	t.Run("ошибка репозитория пробрасывается наверх", func(t *testing.T) {
		repo := newMockCategoryRepository()
		repo.forceError = context.DeadlineExceeded // любая произвольная ошибка для примера
		svc := NewCategoryService(repo)

		category, err := svc.CreateCategory(context.Background(), "Одежда")

		assert.Error(t, err)
		assert.Nil(t, category)
	})
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	t.Run("категория найдена", func(t *testing.T) {
		repo := newMockCategoryRepository()
		repo.categories = []model.Category{{ID: 1, Name: "Книги", Slug: "книги"}}
		svc := NewCategoryService(repo)

		category, err := svc.GetCategoryByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, "Книги", category.Name)
	})

	t.Run("категория не найдена — ErrCategoryNotFound", func(t *testing.T) {
		repo := newMockCategoryRepository()
		svc := NewCategoryService(repo)

		category, err := svc.GetCategoryByID(context.Background(), 999)

		assert.Nil(t, category)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}
