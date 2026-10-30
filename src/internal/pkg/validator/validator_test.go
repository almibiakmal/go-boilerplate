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
	req := requestdto.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "secretpassword",
	}

	err := validator.ValidateStruct(req)
	assert.NoError(t, err)
}

func TestValidateStruct_ValidationErrors(t *testing.T) {
	req := requestdto.RegisterRequest{
		Name:     "B",
		Email:    "not-an-email",
		Password: "123",
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

	assert.Contains(t, fields["name"], "name must be at least 2 characters")
	assert.Contains(t, fields["email"], "email must be a valid email address")
	assert.Contains(t, fields["password"], "password must be at least 6 characters")
}
