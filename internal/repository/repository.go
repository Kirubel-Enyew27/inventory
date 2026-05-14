package repository

import (
	"context"
	"errors"
	"fmt"
	"inventory/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

type ListOptions struct {
	Category string
	LowStock bool
	Search   string
	SKU      string
	Limit    int
	Offset   int
}

const LowStockThreshold = 5

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

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

func (r *ProductRepository) GetAllProducts(ctx context.Context, opts ListOptions) ([]model.Product, error) {
	var products []model.Product
	q := r.db.WithContext(ctx).Model(&model.Product{})

	if opts.Category != "" {
		q = q.Where("category = ?", opts.Category)
	}
	if opts.LowStock {
		q = q.Where("quantity <= ?", LowStockThreshold)
	}
	if opts.Search != "" {
		pattern := "%" + opts.Search + "%"
		q = q.Where("sku ILIKE ? OR name ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}
	if opts.SKU != "" {
		q = q.Where("sku = ?", opts.SKU)
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

func (r *ProductRepository) CountProducts(ctx context.Context, opts ListOptions) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Product{})

	if opts.Category != "" {
		q = q.Where("category = ?", opts.Category)
	}
	if opts.Search != "" {
		pattern := "%" + opts.Search + "%"
		q = q.Where("sku ILIKIE ? OR name ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}
	if opts.SKU != "" {
		q = q.Where("sku = ?", opts.SKU)
	}

	if err := q.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return total, nil
}

func (r *ProductRepository) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).Where("sku=?", sku).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get product by sku: %w", err)
	}
	return &p, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r ProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete product: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepository) AdjustStock(ctx context.Context, id uint, delta int) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, id).Error; err != nil {
			return err
		}

		product.Quantity += delta
		if product.Quantity < 0 {
			return ErrInsufficientStock
		}

		if err := tx.Save(&product).Error; err != nil {
			return fmt.Errorf("save adjusted stock: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &product, nil
}

var (
	ErrDuplicateSKU      = errors.New("sku already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
)

func isUniqueViolation(err error) bool {
	var sqlState interface {
		SQLState() string
	}
	return errors.As(err, &sqlState) && sqlState.SQLState() == "23505"
}
