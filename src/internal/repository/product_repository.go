package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/model"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id uint64) (*model.Product, error)
	FindBySKU(ctx context.Context, sku string) (*model.Product, error)
	List(ctx context.Context, filter servicedto.ListProductFilter) ([]model.Product, int64, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uint64) (bool, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindBySKU(ctx context.Context, sku string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) List(ctx context.Context, f servicedto.ListProductFilter) ([]model.Product, int64, error) {
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Product{})
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		tx = tx.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ?", like, like)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Short-circuit: jika tidak ada data yang cocok, hindari query SELECT kedua ke database.
	if total == 0 {
		return []model.Product{}, 0, nil
	}

	var products []model.Product
	offset := (f.Page - 1) * f.PerPage
	if err := tx.Order("id DESC").Limit(f.PerPage).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete menghapus data dalam 1x query database dan mengembalikan true jika baris ditemukan & terhapus.
func (r *productRepository) Delete(ctx context.Context, id uint64) (bool, error) {
	res := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
