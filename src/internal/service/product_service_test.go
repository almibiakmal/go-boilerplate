package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/service"
)

// mockProductRepo adalah in-memory / stub implementasi dari repository.ProductRepository untuk unit testing.
type mockProductRepo struct {
	createFn    func(ctx context.Context, product *model.Product) error
	findByIDFn  func(ctx context.Context, id uint64) (*model.Product, error)
	findBySKUFn func(ctx context.Context, sku string) (*model.Product, error)
	listFn      func(ctx context.Context, filter servicedto.ListProductFilter) ([]model.Product, int64, error)
	updateFn    func(ctx context.Context, product *model.Product) error
	deleteFn    func(ctx context.Context, id uint64) (bool, error)
}

func (m *mockProductRepo) Create(ctx context.Context, product *model.Product) error {
	return m.createFn(ctx, product)
}

func (m *mockProductRepo) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockProductRepo) FindBySKU(ctx context.Context, sku string) (*model.Product, error) {
	return m.findBySKUFn(ctx, sku)
}

func (m *mockProductRepo) List(ctx context.Context, filter servicedto.ListProductFilter) ([]model.Product, int64, error) {
	return m.listFn(ctx, filter)
}

func (m *mockProductRepo) Update(ctx context.Context, product *model.Product) error {
	return m.updateFn(ctx, product)
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint64) (bool, error) {
	return m.deleteFn(ctx, id)
}

func TestProductService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success create new product", func(t *testing.T) {
		repo := &mockProductRepo{
			findBySKUFn: func(ctx context.Context, sku string) (*model.Product, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, product *model.Product) error {
				product.ID = 1
				return nil
			},
		}

		svc := service.NewProductService(repo)
		res, err := svc.Create(ctx, servicedto.CreateProductInput{
			SKU:   "SKU-100",
			Name:  "Matcha Latte",
			Price: 25000,
		})

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, uint64(1), res.ID)
		assert.Equal(t, "SKU-100", res.SKU)
		assert.Equal(t, "Matcha Latte", res.Name)
		assert.Equal(t, int64(25000), res.Price)
	})

	t.Run("duplicate SKU returns CodeDuplicateEntry", func(t *testing.T) {
		repo := &mockProductRepo{
			findBySKUFn: func(ctx context.Context, sku string) (*model.Product, error) {
				return &model.Product{ID: 1, SKU: sku}, nil
			},
		}

		svc := service.NewProductService(repo)
		res, err := svc.Create(ctx, servicedto.CreateProductInput{
			SKU:   "SKU-100",
			Name:  "Matcha Latte",
			Price: 25000,
		})

		assert.Nil(t, res)
		require.Error(t, err)

		var appErr *response.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, response.CodeDuplicateEntry, appErr.Code)
	})
}

func TestProductService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("not found returns CodeNotFound", func(t *testing.T) {
		repo := &mockProductRepo{
			findByIDFn: func(ctx context.Context, id uint64) (*model.Product, error) {
				return nil, nil
			},
		}

		svc := service.NewProductService(repo)
		res, err := svc.GetByID(ctx, 99)

		assert.Nil(t, res)
		require.Error(t, err)

		var appErr *response.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, response.CodeNotFound, appErr.Code)
	})

	t.Run("found returns product response", func(t *testing.T) {
		repo := &mockProductRepo{
			findByIDFn: func(ctx context.Context, id uint64) (*model.Product, error) {
				return &model.Product{
					ID:    id,
					SKU:   "SKU-200",
					Name:  "Americano",
					Price: 15000,
				}, nil
			},
		}

		svc := service.NewProductService(repo)
		res, err := svc.GetByID(ctx, 10)

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, uint64(10), res.ID)
		assert.Equal(t, "SKU-200", res.SKU)
	})
}

func TestProductService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("not found returns CodeNotFound in single DB call", func(t *testing.T) {
		repo := &mockProductRepo{
			deleteFn: func(ctx context.Context, id uint64) (bool, error) {
				return false, nil
			},
		}

		svc := service.NewProductService(repo)
		err := svc.Delete(ctx, 404)
		require.Error(t, err)

		var appErr *response.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, response.CodeNotFound, appErr.Code)
	})

	t.Run("deleted returns nil", func(t *testing.T) {
		repo := &mockProductRepo{
			deleteFn: func(ctx context.Context, id uint64) (bool, error) {
				return true, nil
			},
		}

		svc := service.NewProductService(repo)
		err := svc.Delete(ctx, 1)
		require.NoError(t, err)
	})
}
