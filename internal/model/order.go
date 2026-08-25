package model

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          int         `json:"id"`
	UserID      int         `json:"user_id"`
	Status      OrderStatus `json:"status"`
	TotalAmount int         `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem — позиция заказа с названием товара, подтянутым через JOIN
// (название удобно иметь прямо в ответе API, чтобы не делать отдельный запрос за товаром).
type OrderItem struct {
	ID              int       `json:"id"`
	OrderID         int       `json:"order_id"`
	ProductID       int       `json:"product_id"`
	ProductName     string    `json:"product_name"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase int       `json:"price_at_purchase"`
	CreatedAt       time.Time `json:"created_at"`
}
