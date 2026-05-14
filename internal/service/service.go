package service

import (
	"context"
	"errors"
	"fmt"

	"inventory/internal/model"
	"inventory/internal/repository"

	"gorm.io/gorm"
)

// ProductService contains business logic for products.
type ProductService struct {
	repo *repository.ProductRepository
}

// NewProductService creates a ProductService.
func NewProductService(r *repository.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

var (
	ErrNotFound          = errors.New("product not found")
	ErrInvalidPrice      = errors.New("price must be greater than 0")
	ErrInvalidQuantity   = errors.New("quantity cannot be negative")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// CreateProduct validates and creates a product.
func (s *ProductService) CreateProduct(ctx context.Context, p *model.Product) error {
	if p.Price <= 0 {
		return fmt.Errorf("create product: %w", ErrInvalidPrice)
	}
	if p.Quantity < 0 {
		return fmt.Errorf("create product: %w", ErrInvalidQuantity)
	}
	if err := s.repo.CreateProduct(ctx, p); err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

// GetProductByID returns a product or ErrNotFound.
func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

// GetAllProducts returns products with optional category filter, low-stock flag, and pagination.
func (s *ProductService) GetAllProducts(ctx context.Context, category string, lowStock bool, search string, limit, offset int) ([]model.Product, int64, error) {
	opts := repository.ListOptions{
		Category: category,
		LowStock: lowStock,
		Search:   search,
		Limit:    limit,
		Offset:   offset,
	}
	products, err := s.repo.GetAllProducts(ctx, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	total, err := s.repo.CountProducts(ctx, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	return products, total, nil
}

// UpdateProduct validates and updates a product.
func (s *ProductService) UpdateProduct(ctx context.Context, p *model.Product) error {
	if p.Price <= 0 {
		return fmt.Errorf("update product: %w", ErrInvalidPrice)
	}
	if p.Quantity < 0 {
		return fmt.Errorf("update product: %w", ErrInvalidQuantity)
	}
	if err := s.repo.UpdateProduct(ctx, p); err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

// DeleteProduct removes a product by ID.
func (s *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

// AdjustStock safely updates product stock by delta (can be negative).
func (s *ProductService) AdjustStock(ctx context.Context, id uint, delta int) (*model.Product, error) {
	p, err := s.repo.AdjustStock(ctx, id, delta)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, fmt.Errorf("adjust stock: %w", err)
	}
	return p, nil
}
