package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
)

var (
	ErrEmptyCart         = errors.New("корзина пуста")
	ErrInsufficientStock = errors.New("недостаточно товара на складе")
	ErrOrderNotFound     = errors.New("заказ не найден")
	ErrOrderAccessDenied = errors.New("у вас нет доступа к этому заказу")
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	cartRepo *repository.CartRepository,
	productRepo *repository.ProductRepository,
) *OrderService {
	return &OrderService{orderRepo: orderRepo, cartRepo: cartRepo, productRepo: productRepo}
}

type OrderWithItems struct {
	Order model.Order       `json:"order"`
	Items []model.OrderItem `json:"items"`
}

// Checkout оформляет заказ из текущей корзины пользователя.
// Вся операция выполняется в ОДНОЙ транзакции: либо весь заказ создаётся
// успешно (заказ + позиции + списание остатков + очистка корзины),
// либо не происходит НИЧЕГО — база остаётся ровно в том состоянии, что была
// до вызова этого метода.
func (s *OrderService) Checkout(ctx context.Context, userID int) (*OrderWithItems, error) {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItems, err := s.cartRepo.ListItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	// Сортируем товары по product_id ПЕРЕД блокировкой строк — это защита от
	// deadlock
	sort.Slice(cartItems, func(i, j int) bool {
		return cartItems[i].ProductID < cartItems[j].ProductID
	})

	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: begin transaction: %w", err)
	}
	// defer tx.Rollback() — стандартный, безопасный паттерн работы с транзакциями в Go.
	defer tx.Rollback()

	totalAmount := 0

	for _, item := range cartItems {
		product, err := s.productRepo.GetForUpdate(ctx, tx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, fmt.Errorf("товар с id %d больше не существует", item.ProductID)
		}
		if product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("%w: товар «%s», доступно %d, запрошено %d",
				ErrInsufficientStock, product.Name, product.StockQuantity, item.Quantity)
		}
		totalAmount += item.Price * item.Quantity
	}

	// Шаг 2: создаём сам заказ
	order, err := s.orderRepo.CreateOrderTx(ctx, tx, userID, totalAmount)
	if err != nil {
		return nil, err
	}

	// Шаг 3: создаём позиции заказа и списываем остатки товара
	for _, item := range cartItems {
		if err := s.orderRepo.CreateOrderItemTx(ctx, tx, order.ID, item.ProductID, item.Quantity, item.Price); err != nil {
			return nil, err
		}
		if err := s.productRepo.DecrementStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
			return nil, err
		}
	}

	// Шаг 4: очищаем корзину — она принадлежит уже оформленному заказу, не старой "предкорзине"
	if err := s.cartRepo.ClearCartTx(ctx, tx, cart.ID); err != nil {
		return nil, err
	}

	// Шаг 5: ФИКСИРУЕМ транзакцию — только теперь все изменения реально применяются
	// к базе данных, одним неделимым блоком. До этого момента снаружи никто (даже
	// параллельные запросы) не видел ни одного из этих промежуточных изменений.
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("service: commit transaction: %w", err)
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	return &OrderWithItems{Order: *order, Items: items}, nil
}

// ListMyOrders возвращает список заказов пользователя (без детализации по товарам).
func (s *OrderService) ListMyOrders(ctx context.Context, userID int) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

// GetOrderByID возвращает заказ с деталями, только если он принадлежит пользователю
func (s *OrderService) GetOrderByID(ctx context.Context, orderID, userID int, isAdmin bool) (*OrderWithItems, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	if order.UserID != userID && !isAdmin {
		return nil, ErrOrderAccessDenied
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &OrderWithItems{Order: *order, Items: items}, nil
}
