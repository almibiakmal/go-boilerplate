package validator_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	requestdto "maj-core-srvc/src/internal/dto/request"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
)

func TestValidateStruct_Success(t *testing.T) {
	req := requestdto.CreateProductRequest{
		SKU:   "SKU-001",
		Name:  "Kopi Susu Aren",
		Price: 18000,
	}

	err := validator.ValidateStruct(req)
	assert.NoError(t, err)
}

func TestValidateStruct_ValidationErrors(t *testing.T) {
	req := requestdto.CreateProductRequest{
		SKU:   "",
		Name:  "A",
		Price: -500,
	}

	err := validator.ValidateStruct(req)
	require.Error(t, err)

	var appErr *response.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeValidationError, appErr.Code)
	require.Len(t, appErr.Details, 3)

	fields := map[string]string{}
	for _, d := range appErr.Details {
		fields[d.Field] = d.Message
	}

	assert.Contains(t, fields["sku"], "sku is required")
	assert.Contains(t, fields["name"], "name must be at least 2 characters")
	assert.Contains(t, fields["price"], "price must be greater than or equal to 0")
}

