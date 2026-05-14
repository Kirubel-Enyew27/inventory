package repository

import (
	"context"
	"fmt"
	"inventory/internal/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
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

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	var products []model.Product
	if err := r.db.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
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
