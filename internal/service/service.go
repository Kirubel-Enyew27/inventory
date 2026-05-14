package service

import (
	"context"
	"errors"
	"fmt"
	"inventory/internal/model"
	"inventory/internal/repository"
	"strings"

	"gorm.io/gorm"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(r *repository.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

var (
	ErrNotFound          = errors.New("product not found")
	ErrInvalidPrice      = errors.New("price must be greater than 0")
	ErrInvalidQuantity   = errors.New("quantity cannot be negative")
	ErrInvalidSKU        = errors.New("sku is required")
	ErrDuplicateSKU      = errors.New("sku already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
)

func (s *ProductService) CreateProduct(ctx context.Context, p *model.Product) error {
	p.SKU = strings.TrimSpace(p.SKU)
	if p.SKU == "" {
		return fmt.Errorf("create product: %w", ErrInvalidSKU)
	}
	if p.Price <= 0 {
		return fmt.Errorf("create product: %w", ErrInvalidPrice)
	}
	if p.Quantity < 0 {
		return fmt.Errorf("create product: %w", ErrInvalidQuantity)
	}
	existing, err := s.repo.GetProductBySKU(ctx, p.SKU)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("create product: %w", ErrDuplicateSKU)
	}
	if err := s.repo.CreateProduct(ctx, p); err != nil {
		if errors.Is(err, repository.ErrDuplicateSKU) {
			return fmt.Errorf("create product: %w", ErrDuplicateSKU)
		}
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {

	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, category string, lowStock bool, search string, sku string, limit, offset int) ([]model.Product, int64, error) {
	opts := repository.ListOptions{
		Category: category,
		LowStock: lowStock,
		Search:   search,
		SKU:      sku,
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

func (s *ProductService) UpdateProduct(ctx context.Context, p *model.Product) error {
	p.SKU = strings.TrimSpace(p.SKU)
	if p.SKU == "" {
		return fmt.Errorf("update product: %w", ErrInvalidSKU)
	}
	if p.Price <= 0 {
		return fmt.Errorf("update product: %w", ErrInvalidPrice)
	}
	if p.Quantity < 0 {
		return fmt.Errorf("update product: %w", ErrInvalidQuantity)
	}
	existing, err := s.repo.GetProductBySKU(ctx, p.SKU)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if existing != nil && existing.ID != p.ID {
		return fmt.Errorf("update product: %w", ErrDuplicateSKU)
	}
	if err := s.repo.UpdateProduct(ctx, p); err != nil {
		if errors.Is(err, repository.ErrDuplicateSKU) {
			return fmt.Errorf("update product: %w", ErrDuplicateSKU)
		}
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

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
