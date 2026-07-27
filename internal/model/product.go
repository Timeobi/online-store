package model

import "time"

// Product представляет товар в каталоге магазина.
type Product struct {
	ID            int       `json:"id"`
	CategoryID    *int      `json:"category_id"` // указатель, потому что может быть NULL
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	Price         int       `json:"price"` // цена в копейках
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
