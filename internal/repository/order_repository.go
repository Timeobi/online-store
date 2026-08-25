package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// BeginTx открывает новую транзакцию БД. Сервис будет использовать её
// для всей логики оформления заказа как единого атомарного блока.
func (r *OrderRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// CreateOrderTx создаёт запись заказа В РАМКАХ транзакции.
func (r *OrderRepository) CreateOrderTx(ctx context.Context, tx *sql.Tx, userID, totalAmount int) (*model.Order, error) {
	query := `
		INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, status, total_amount, created_at, updated_at
	`

	var o model.Order
	err := tx.QueryRowContext(ctx, query, userID, model.OrderStatusPending, totalAmount).Scan(
		&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create order: %w", err)
	}

	return &o, nil
}

// CreateOrderItemTx добавляет одну позицию заказа В РАМКАХ транзакции.
func (r *OrderRepository) CreateOrderItemTx(ctx context.Context, tx *sql.Tx, orderID, productID, quantity, priceAtPurchase int) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, orderID, productID, quantity, priceAtPurchase)
	if err != nil {
		return fmt.Errorf("repository: create order item: %w", err)
	}
	return nil
}

// ListByUser возвращает все заказы пользователя, от новых к старым.
func (r *OrderRepository) ListByUser(ctx context.Context, userID int) ([]model.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}

	return orders, nil
}

// GetByID возвращает заказ по ID (без позиций — их получаем отдельным методом ниже).
func (r *OrderRepository) GetByID(ctx context.Context, id int) (*model.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders WHERE id = $1
	`

	var o model.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get order by id: %w", err)
	}

	return &o, nil
}

// GetItemsByOrderID возвращает позиции заказа вместе с ТЕКУЩИМ названием товара (JOIN).
func (r *OrderRepository) GetItemsByOrderID(ctx context.Context, orderID int) ([]model.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, p.name, oi.quantity, oi.price_at_purchase, oi.created_at
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("repository: get order items: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.PriceAtPurchase, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}

	return items, nil
}
