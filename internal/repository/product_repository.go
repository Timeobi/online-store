package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// CreateProductParams — входные параметры для создания товара.
type CreateProductParams struct {
	CategoryID    *int
	Name          string
	Slug          string
	Description   string
	Price         int
	StockQuantity int
}

func (r *ProductRepository) Create(ctx context.Context, p CreateProductParams) (*model.Product, error) {
	query := `
		INSERT INTO products (category_id, name, slug, description, price, stock_quantity)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, category_id, name, slug, description, price, stock_quantity, created_at, updated_at
	`

	var prod model.Product
	err := r.db.QueryRowContext(ctx, query,
		p.CategoryID, p.Name, p.Slug, p.Description, p.Price, p.StockQuantity,
	).Scan(
		&prod.ID, &prod.CategoryID, &prod.Name, &prod.Slug, &prod.Description,
		&prod.Price, &prod.StockQuantity, &prod.CreatedAt, &prod.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create product: %w", err)
	}

	return &prod, nil
}

// ListParams — параметры для получения списка товаров с фильтрацией и пагинацией.
type ListParams struct {
	CategoryID *int
	Search     string
	Limit      int
	Offset     int
}

// List возвращает товары с учётом фильтров, плюс общее количество (для пагинации на фронте).
func (r *ProductRepository) List(ctx context.Context, params ListParams) ([]model.Product, int, error) {
	// Строим WHERE-условие динамически, в зависимости от того, какие фильтры переданы
	var conditions []string
	var args []interface{}
	argIdx := 1

	if params.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *params.CategoryID)
		argIdx++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Сначала считаем общее количество подходящих товаров (без LIMIT/OFFSET)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository: count products: %w", err)
	}

	// Теперь получаем сам список с пагинацией
	listArgs := append(args, params.Limit, params.Offset)
	listQuery := fmt.Sprintf(`
		SELECT id, category_id, name, slug, description, price, stock_quantity, created_at, updated_at
		FROM products
		%s
		ORDER BY id
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
			&p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scan product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: rows iteration error: %w", err)
	}

	return products, total, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*model.Product, error) {
	query := `
		SELECT id, category_id, name, slug, description, price, stock_quantity, created_at, updated_at
		FROM products WHERE id = $1
	`

	var p model.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
		&p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get product by id: %w", err)
	}

	return &p, nil
}

// UpdateProductParams — поля, которые можно обновить.
type UpdateProductParams struct {
	Name          string
	Description   string
	Price         int
	StockQuantity int
}

func (r *ProductRepository) Update(ctx context.Context, id int, p UpdateProductParams) (*model.Product, error) {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock_quantity = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, category_id, name, slug, description, price, stock_quantity, created_at, updated_at
	`

	var prod model.Product
	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.Price, p.StockQuantity, id).Scan(
		&prod.ID, &prod.CategoryID, &prod.Name, &prod.Slug, &prod.Description,
		&prod.Price, &prod.StockQuantity, &prod.CreatedAt, &prod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: update product: %w", err)
	}

	return &prod, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) (bool, error) {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("repository: delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repository: rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetForUpdate возвращает товар по ID и БЛОКИРУЕТ его строку до конца транзакции
// (используется параметр tx, а не обычный db — блокировка имеет смысл только внутри транзакции).
// Применяется при оформлении заказа, чтобы не допустить одновременной продажи
// последних единиц товара двумя разными покупателями.
func (r *ProductRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, id int) (*model.Product, error) {
	query := `
		SELECT id, category_id, name, slug, description, price, stock_quantity, created_at, updated_at
		FROM products
		WHERE id = $1
		FOR UPDATE
	`

	var p model.Product
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
		&p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get product for update: %w", err)
	}

	return &p, nil
}

// DecrementStock уменьшает остаток товара на складе.
// Условие "AND stock_quantity >= $1" в WHERE — это дополнительная защита
// (defense in depth): даже если блокировка строки почему-то не сработала как ожидалось,
// сам UPDATE физически не сможет увести остаток в минус — просто не затронет ни одной строки.
func (r *ProductRepository) DecrementStock(ctx context.Context, tx *sql.Tx, id, quantity int) error {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity - $1, updated_at = NOW()
		WHERE id = $2 AND stock_quantity >= $1
	`
	result, err := tx.ExecContext(ctx, query, quantity, id)
	if err != nil {
		return fmt.Errorf("repository: decrement stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository: insufficient stock for product %d", id)
	}

	return nil
}
