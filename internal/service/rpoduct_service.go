package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
)

var ErrProductNotFound = errors.New("product not found")

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

type CreateProductInput struct {
	CategoryID    *int
	Name          string
	Description   string
	Price         int
	StockQuantity int
}

func (s *ProductService) CreateProduct(ctx context.Context, input CreateProductInput) (*model.Product, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("название товара обязательно")
	}
	if input.Price < 0 {
		return nil, errors.New("цена не может быть отрицательной")
	}
	if input.StockQuantity < 0 {
		return nil, errors.New("остаток не может быть отрицательным")
	}

	slug := generateSlug(name)

	return s.repo.Create(ctx, repository.CreateProductParams{
		CategoryID:    input.CategoryID,
		Name:          name,
		Slug:          slug,
		Description:   input.Description,
		Price:         input.Price,
		StockQuantity: input.StockQuantity,
	})
}

type ListProductsInput struct {
	CategoryID *int
	Search     string
	Page       int
	PageSize   int
}

type ListProductsOutput struct {
	Products []model.Product
	Total    int
	Page     int
	PageSize int
}

func (s *ProductService) ListProducts(ctx context.Context, input ListProductsInput) (*ListProductsOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = defaultLimit
	}
	if pageSize > maxLimit {
		pageSize = maxLimit
	}

	offset := (page - 1) * pageSize

	products, total, err := s.repo.List(ctx, repository.ListParams{
		CategoryID: input.CategoryID,
		Search:     input.Search,
		Limit:      pageSize,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	return &ListProductsOutput{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

type UpdateProductInput struct {
	Name          string
	Description   string
	Price         int
	StockQuantity int
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int, input UpdateProductInput) (*model.Product, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("название товара обязательно")
	}
	if input.Price < 0 {
		return nil, errors.New("цена не может быть отрицательной")
	}

	product, err := s.repo.Update(ctx, id, repository.UpdateProductParams{
		Name:          name,
		Description:   input.Description,
		Price:         input.Price,
		StockQuantity: input.StockQuantity,
	})
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrProductNotFound
	}
	return nil
}
