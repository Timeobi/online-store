package service

import (
	"testing"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		name  string
		items []model.CartItemDetail
		want  int
	}{
		{
			name:  "пустая корзина",
			items: []model.CartItemDetail{},
			want:  0,
		},
		{
			name: "один товар",
			items: []model.CartItemDetail{
				{ProductID: 1, Price: 1000, Quantity: 2, Subtotal: 2000},
			},
			want: 2000,
		},
		{
			name: "несколько товаров",
			items: []model.CartItemDetail{
				{ProductID: 1, Price: 1000, Quantity: 2, Subtotal: 2000},
				{ProductID: 2, Price: 500, Quantity: 3, Subtotal: 1500},
			},
			want: 3500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTotal(tt.items)
			if got != tt.want {
				t.Errorf("calculateTotal() = %d, хотели %d", got, tt.want)
			}
		})
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		name     string
		quantity int
		wantErr  bool
	}{
		{name: "положительное количество", quantity: 5, wantErr: false},
		{name: "ноль — недопустимо", quantity: 0, wantErr: true},
		{name: "отрицательное — недопустимо", quantity: -3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuantity(tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQuantity(%d) error = %v, wantErr %v", tt.quantity, err, tt.wantErr)
			}
		})
	}
}
