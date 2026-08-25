package model

import "time"

type Cart struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// CartItemDetail — товар в корзине вместе с данными о продукте (для показа клиенту).
// Получается через JOIN cart_items + products, поэтому это отдельная структура,
// а не просто CartItem — модели БД и модели "для API-ответа" не всегда совпадают 1:1.
type CartItemDetail struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	Subtotal    int    `json:"subtotal"`
}
