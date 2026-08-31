//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB поднимает временный контейнер PostgreSQL специально для теста
// и применяет к нему нашу схему БД. Возвращает подключение и функцию очистки.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	// Создаём минимальную схему прямо здесь, для теста —
	// достаточно только той таблицы, которую тестируем.
	_, err = db.ExecContext(ctx, `
		CREATE TABLE categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			slug VARCHAR(120) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestCategoryRepository_Create_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewCategoryRepository(db)

	category, err := repo.Create(context.Background(), "Электроника", "elektronika")

	assert.NoError(t, err)
	assert.NotZero(t, category.ID)
	assert.Equal(t, "Электроника", category.Name)
	assert.Equal(t, "elektronika", category.Slug)
	assert.NotZero(t, category.CreatedAt)
}

func TestCategoryRepository_GetAll_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewCategoryRepository(db)

	_, err := repo.Create(context.Background(), "Книги", "knigi")
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), "Одежда", "odezhda")
	require.NoError(t, err)

	categories, err := repo.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
}

func TestCategoryRepository_GetByID_NotFound_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewCategoryRepository(db)

	category, err := repo.GetByID(context.Background(), 999)

	assert.NoError(t, err)
	assert.Nil(t, category)
}
