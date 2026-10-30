package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	govalidator "github.com/go-playground/validator/v10"

	"maj-core-srvc/src/internal/pkg/response"
)

var (
	once     sync.Once
	instance *govalidator.Validate
)

// getValidator mengembalikan singleton instance go-playground/validator
// yang sudah dikonfigurasi untuk membaca nama field dari tag `json` atau `query`.
func getValidator() *govalidator.Validate {
	once.Do(func() {
		v := govalidator.New()
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			for _, tag := range []string{"json", "query"} {
				name := strings.SplitN(fld.Tag.Get(tag), ",", 2)[0]
				if name != "" && name != "-" {
					return name
				}
			}
			return strings.ToLower(fld.Name)
		})
		instance = v
	})
	return instance
}

// ValidateStruct memvalidasi struct berdasarkan tag `validate:"..."`.
// Jika gagal, mengembalikan *response.AppError dengan CodeValidationError (422) beserta daftar FieldError.
func ValidateStruct(s any) error {
	err := getValidator().Struct(s)
	if err == nil {
		return nil
	}

	var valErrs govalidator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return response.NewError(response.CodeBadRequest, "Invalid request data")
	}

	details := make([]response.FieldError, 0, len(valErrs))
	for _, fe := range valErrs {
		details = append(details, response.FieldError{
			Field:   fe.Field(),
			Message: formatFieldError(fe),
		})
	}

	return response.NewValidationError(details)
}

func formatFieldError(fe govalidator.FieldError) string {
	field := fe.Field()
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, param)
		}
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, param)
		}
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s failed on %s validation", field, fe.Tag())
	}
}

