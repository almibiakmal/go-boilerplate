package service

import (
	"context"
	"strings"

	responsedto "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/repository"
)

type ProductService interface {
	Create(ctx context.Context, input servicedto.CreateProductInput) (*responsedto.ProductResponse, error)
	GetByID(ctx context.Context, id uint64) (*responsedto.ProductResponse, error)
	List(ctx context.Context, filter servicedto.ListProductFilter) ([]responsedto.ProductResponse, int64, error)
	Update(ctx context.Context, input servicedto.UpdateProductInput) (*responsedto.ProductResponse, error)
	Delete(ctx context.Context, id uint64) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, input servicedto.CreateProductInput) (*responsedto.ProductResponse, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	input.Name = strings.TrimSpace(input.Name)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindBySKU(ctx, input.SKU)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to check existing SKU", err)
	}
	if existing != nil {
		return nil, response.NewError(response.CodeDuplicateEntry, "Product with this SKU already exists")
	}

	product := &model.Product{
		SKU:   input.SKU,
		Name:  input.Name,
		Price: input.Price,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to create product", err)
	}

	logger.Info(ctx, "product created", "product_id", product.ID, "sku", product.SKU)
	return responsedto.ToProductResponse(product), nil
}

func (s *productService) GetByID(ctx context.Context, id uint64) (*responsedto.ProductResponse, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch product", err)
	}
	if product == nil {
		return nil, response.NewError(response.CodeNotFound, "Product not found")
	}
	return responsedto.ToProductResponse(product), nil
}

func (s *productService) List(ctx context.Context, filter servicedto.ListProductFilter) ([]responsedto.ProductResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 10
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, response.WrapError(response.CodeDatabaseError, "Failed to list products", err)
	}
	return responsedto.ToProductListResponse(items), total, nil
}

func (s *productService) Update(ctx context.Context, input servicedto.UpdateProductInput) (*responsedto.ProductResponse, error) {
	input.Name = strings.TrimSpace(input.Name)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	product, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch product", err)
	}
	if product == nil {
		return nil, response.NewError(response.CodeNotFound, "Product not found")
	}

	product.Name = input.Name
	product.Price = input.Price

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to update product", err)
	}

	logger.Info(ctx, "product updated", "product_id", product.ID)
	return responsedto.ToProductResponse(product), nil
}

func (s *productService) Delete(ctx context.Context, id uint64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return response.WrapError(response.CodeDatabaseError, "Failed to delete product", err)
	}
	if !deleted {
		return response.NewError(response.CodeNotFound, "Product not found")
	}

	logger.Info(ctx, "product deleted", "product_id", id)
	return nil
}
