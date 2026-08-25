package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// GetOrCreateCart возвращает корзину пользователя, создавая новую, если её ещё нет.
// У каждого пользователя ровно одна корзина.
func (r *CartRepository) GetOrCreateCart(ctx context.Context, userID int) (*model.Cart, error) {
	var cart model.Cart

	query := `SELECT id, user_id, created_at FROM carts WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt)
	if err == nil {
		return &cart, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("repository: get cart: %w", err)
	}

	// Корзины ещё нет — создаём новую
	insertQuery := `INSERT INTO carts (user_id) VALUES ($1) RETURNING id, user_id, created_at`
	err = r.db.QueryRowContext(ctx, insertQuery, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository: create cart: %w", err)
	}

	return &cart, nil
}

// AddItem добавляет товар в корзину. Если товар уже там есть — увеличивает количество,
// а не создаёт вторую строку (за это отвечает ON CONFLICT — используем UNIQUE(cart_id, product_id)
func (r *CartRepository) AddItem(ctx context.Context, cartID, productID, quantity int) error {
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`
	_, err := r.db.ExecContext(ctx, query, cartID, productID, quantity)
	if err != nil {
		return fmt.Errorf("repository: add cart item: %w", err)
	}
	return nil
}

// SetItemQuantity устанавливает ТОЧНОЕ количество товара в корзине (не прибавляет, а заменяет).
func (r *CartRepository) SetItemQuantity(ctx context.Context, cartID, productID, quantity int) (bool, error) {
	query := `UPDATE cart_items SET quantity = $1 WHERE cart_id = $2 AND product_id = $3`
	result, err := r.db.ExecContext(ctx, query, quantity, cartID, productID)
	if err != nil {
		return false, fmt.Errorf("repository: set cart item quantity: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repository: rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

// RemoveItem удаляет один товар из корзины.
func (r *CartRepository) RemoveItem(ctx context.Context, cartID, productID int) (bool, error) {
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`
	result, err := r.db.ExecContext(ctx, query, cartID, productID)
	if err != nil {
		return false, fmt.Errorf("repository: remove cart item: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repository: rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

// ListItems возвращает товары в корзине ВМЕСТЕ с данными о продукте — через JOIN.
func (r *CartRepository) ListItems(ctx context.Context, cartID int) ([]model.CartItemDetail, error) {
	query := `
		SELECT p.id, p.name, p.price, ci.quantity
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id
	`
	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, fmt.Errorf("repository: list cart items: %w", err)
	}
	defer rows.Close()

	var items []model.CartItemDetail
	for rows.Next() {
		var item model.CartItemDetail
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity); err != nil {
			return nil, fmt.Errorf("repository: scan cart item: %w", err)
		}
		item.Subtotal = item.Price * item.Quantity
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}

	return items, nil
}

// ClearCartTx удаляет все товары из корзины В РАМКАХ переданной транзакции.
// Используется при оформлении заказа — обрати внимание на параметр *sql.Tx
// вместо обычного db: этот метод предназначен для вызова ВНУТРИ транзакции Checkout.
func (r *CartRepository) ClearCartTx(ctx context.Context, tx *sql.Tx, cartID int) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := tx.ExecContext(ctx, query, cartID)
	if err != nil {
		return fmt.Errorf("repository: clear cart: %w", err)
	}
	return nil
}
