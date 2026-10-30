package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	responsedto "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/handler"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/router"
)

type mockProductService struct {
	createFn  func(ctx context.Context, input servicedto.CreateProductInput) (*responsedto.ProductResponse, error)
	getByIDFn func(ctx context.Context, id uint64) (*responsedto.ProductResponse, error)
	listFn    func(ctx context.Context, filter servicedto.ListProductFilter) ([]responsedto.ProductResponse, int64, error)
	updateFn  func(ctx context.Context, input servicedto.UpdateProductInput) (*responsedto.ProductResponse, error)
	deleteFn  func(ctx context.Context, id uint64) error
}

func (m *mockProductService) Create(ctx context.Context, input servicedto.CreateProductInput) (*responsedto.ProductResponse, error) {
	return m.createFn(ctx, input)
}

func (m *mockProductService) GetByID(ctx context.Context, id uint64) (*responsedto.ProductResponse, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockProductService) List(ctx context.Context, filter servicedto.ListProductFilter) ([]responsedto.ProductResponse, int64, error) {
	return m.listFn(ctx, filter)
}

func (m *mockProductService) Update(ctx context.Context, input servicedto.UpdateProductInput) (*responsedto.ProductResponse, error) {
	return m.updateFn(ctx, input)
}

func (m *mockProductService) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func setupTestApp(svc *mockProductService) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	router.RegisterHTTP(app, router.Dependencies{
		HealthHandler:  handler.NewHealthHandler(nil),
		ProductHandler: handler.NewProductHandler(svc),
	})
	return app
}

func TestProductHandler_Create(t *testing.T) {
	t.Run("returns 422 VALIDATION_ERROR on invalid payload", func(t *testing.T) {
		app := setupTestApp(&mockProductService{})

		body := []byte(`{"sku":"","name":"A","price":-10}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		var env response.Envelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
		assert.False(t, env.Success)
		assert.Equal(t, response.CodeValidationError, env.Code)
		require.NotNil(t, env.Error)
		assert.Len(t, env.Error.Details, 3)
	})

	t.Run("returns 201 CREATED on valid payload", func(t *testing.T) {
		svc := &mockProductService{
			createFn: func(ctx context.Context, input servicedto.CreateProductInput) (*responsedto.ProductResponse, error) {
				return &responsedto.ProductResponse{
					ID:    1,
					SKU:   input.SKU,
					Name:  input.Name,
					Price: input.Price,
				}, nil
			},
		}
		app := setupTestApp(svc)

		body := []byte(`{"sku":"SKU-001","name":"Espresso","price":15000}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var env response.Envelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
		assert.True(t, env.Success)
		assert.Equal(t, response.CodeCreated, env.Code)
	})
}

func TestProductHandler_GetByID(t *testing.T) {
	t.Run("returns 404 NOT_FOUND when product does not exist", func(t *testing.T) {
		svc := &mockProductService{
			getByIDFn: func(ctx context.Context, id uint64) (*responsedto.ProductResponse, error) {
				return nil, response.NewError(response.CodeNotFound, "Product not found")
			},
		}
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/404", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var env response.Envelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
		assert.False(t, env.Success)
		assert.Equal(t, response.CodeNotFound, env.Code)
	})
}

