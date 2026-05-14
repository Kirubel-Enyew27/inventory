package service

import (
	"context"
	"errors"
	"fmt"
	"inventory/internal/model"
	"inventory/internal/repository"

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
	ErrInsufficientStock = errors.New("insufficient stock")
)

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

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {

	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, category string) ([]model.Product, error) {
	products, err := s.repo.GetAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	if category == "" {
		return products, nil
	}

	var out []model.Product
	for _, p := range products {
		if p.Category == category {
			out = append(out, p)
		}
	}
	return out, nil
}

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

func (s *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (s *ProductService) AdjustStock(ctx context.Context, id uint, delta int) error {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return fmt.Errorf("adjust stock: %w", err)
	}
	if p == nil {
		return ErrNotFound
	}
	newQty := p.Quantity + delta
	if newQty < 0 {
		return ErrInsufficientStock
	}
	p.Quantity = newQty
	if err := s.repo.UpdateProduct(ctx, p); err != nil {
		return fmt.Errorf("adjust stock: %w", err)
	}
	return nil
}
