package repository

import (
	"context"
	"fmt"

	"inventory/internal/model"

	"gorm.io/gorm"
)

// ProductRepository provides DB access for products.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// CreateProduct inserts a new product record.
func (r *ProductRepository) CreateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

// GetProductByID returns a product by its ID.
func (r *ProductRepository) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	return &p, nil
}

// GetAllProducts returns all products.
// ListOptions controls product listing queries.
type ListOptions struct {
	Category string
	LowStock bool
	Limit    int
	Offset   int
}

// LowStockThreshold is the default threshold used for low-stock queries.
const LowStockThreshold = 5

func (r *ProductRepository) GetAllProducts(ctx context.Context, opts ListOptions) ([]model.Product, error) {
	var products []model.Product
	q := r.db.WithContext(ctx).Model(&model.Product{})

	if opts.Category != "" {
		q = q.Where("category = ?", opts.Category)
	}
	if opts.LowStock {
		q = q.Where("quantity <= ?", LowStockThreshold)
	}
	if opts.Limit > 0 {
		q = q.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		q = q.Offset(opts.Offset)
	}

	if err := q.Order("id DESC").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}
	return products, nil
}

// UpdateProduct updates an existing product.
func (r *ProductRepository) UpdateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

// DeleteProduct removes a product by ID.
func (r *ProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete product: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

