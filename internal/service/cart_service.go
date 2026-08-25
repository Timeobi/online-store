package service

import (
	"context"
	"errors"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
)

var (
	ErrInvalidQuantity  = errors.New("количество должно быть больше нуля")
	ErrProductNotInCart = errors.New("товар не найден в корзине")
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo}
}

type CartView struct {
	Items []model.CartItemDetail `json:"items"`
	Total int                    `json:"total"`
}

// calculateTotal — ЧИСТАЯ функция (pure function): не обращается к БД, не имеет
// побочных эффектов, зависит только от своих аргументов. Такие функции проще
// всего тестировать — не нужна ни БД, ни сеть, ни какие-либо "заглушки" (моки).
func calculateTotal(items []model.CartItemDetail) int {
	total := 0
	for _, item := range items {
		total += item.Subtotal
	}
	return total
}

// validateQuantity — тоже чистая функция, простое правило валидации.
func validateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

// GetCart возвращает текущую корзину пользователя с товарами и итоговой суммой.
func (s *CartService) GetCart(ctx context.Context, userID int) (*CartView, error) {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.cartRepo.ListItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}

	return &CartView{Items: items, Total: calculateTotal(items)}, nil
}

// AddItem добавляет товар в корзину, проверив, что он существует и количество корректно.
func (s *CartService) AddItem(ctx context.Context, userID, productID, quantity int) error {
	if err := validateQuantity(quantity); err != nil {
		return err
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	return s.cartRepo.AddItem(ctx, cart.ID, productID, quantity)
}

// UpdateItemQuantity меняет количество конкретного товара в корзине.
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID, quantity int) error {
	if err := validateQuantity(quantity); err != nil {
		return err
	}

	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	updated, err := s.cartRepo.SetItemQuantity(ctx, cart.ID, productID, quantity)
	if err != nil {
		return err
	}
	if !updated {
		return ErrProductNotInCart
	}

	return nil
}

// RemoveItem удаляет товар из корзины.
func (s *CartService) RemoveItem(ctx context.Context, userID, productID int) error {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	removed, err := s.cartRepo.RemoveItem(ctx, cart.ID, productID)
	if err != nil {
		return err
	}
	if !removed {
		return ErrProductNotInCart
	}

	return nil
}
