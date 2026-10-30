package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Envelope adalah struktur standar JSON response untuk seluruh endpoint HTTP.
type Envelope struct {
	Success bool          `json:"success"`
	Code    Code          `json:"code"`
	Message string        `json:"message"`
	Data    any           `json:"data,omitempty"`
	Meta    *PageMeta     `json:"meta,omitempty"`
	Error   *ErrorPayload `json:"error,omitempty"`
}

// PageMeta menyimpan informasi pagination untuk endpoint list.
type PageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// ErrorPayload menyimpan detail error pada response gagal.
type ErrorPayload struct {
	Code    Code         `json:"code"`
	Details []FieldError `json:"details,omitempty"`
}

// JSON mengirim response dengan Code yang otomatis di-mapping ke HTTP status code.
func JSON(c *fiber.Ctx, code Code, message string, data any) error {
	meta := code.Resolve()
	if message == "" {
		message = meta.DefaultMessage
	}

	return c.Status(meta.HTTPStatus).JSON(Envelope{
		Success: meta.HTTPStatus >= 200 && meta.HTTPStatus < 300,
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// OK mengirim response 200 OK dengan CodeSuccess.
func OK(c *fiber.Ctx, message string, data any) error {
	return JSON(c, CodeSuccess, message, data)
}

// Created mengirim response 201 Created dengan CodeCreated.
func Created(c *fiber.Ctx, message string, data any) error {
	return JSON(c, CodeCreated, message, data)
}

// Paginated mengirim response 200 OK beserta metadata pagination.
func Paginated(c *fiber.Ctx, message string, data any, page, perPage int, totalItems int64) error {
	meta := CodeSuccess.Resolve()
	if message == "" {
		message = meta.DefaultMessage
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int((totalItems + int64(perPage) - 1) / int64(perPage))
	}

	return c.Status(meta.HTTPStatus).JSON(Envelope{
		Success: true,
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
		Meta: &PageMeta{
			Page:       page,
			PerPage:    perPage,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	})
}

// Error menerjemahkan error (baik *AppError, *fiber.Error, maupun error biasa)
// ke format JSON Envelope standar beserta HTTP status code dari mapping Code.
func Error(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		meta := appErr.Code.Resolve()
		msg := appErr.Message
		if msg == "" {
			msg = meta.DefaultMessage
		}
		return c.Status(meta.HTTPStatus).JSON(Envelope{
			Success: false,
			Code:    appErr.Code,
			Message: msg,
			Error: &ErrorPayload{
				Code:    appErr.Code,
				Details: appErr.Details,
			},
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code := mapHTTPStatusToCode(fiberErr.Code)
		return c.Status(fiberErr.Code).JSON(Envelope{
			Success: false,
			Code:    code,
			Message: fiberErr.Message,
			Error: &ErrorPayload{
				Code: code,
			},
		})
	}

	// Default fallback: 500 Internal Server Error
	meta := CodeInternalError.Resolve()
	return c.Status(meta.HTTPStatus).JSON(Envelope{
		Success: false,
		Code:    CodeInternalError,
		Message: meta.DefaultMessage,
		Error: &ErrorPayload{
			Code: CodeInternalError,
		},
	})
}

func mapHTTPStatusToCode(status int) Code {
	switch status {
	case fiber.StatusBadRequest:
		return CodeBadRequest
	case fiber.StatusUnauthorized:
		return CodeUnauthorized
	case fiber.StatusForbidden:
		return CodeForbidden
	case fiber.StatusNotFound:
		return CodeNotFound
	case fiber.StatusConflict:
		return CodeConflict
	case fiber.StatusUnprocessableEntity:
		return CodeValidationError
	case fiber.StatusTooManyRequests:
		return CodeTooManyRequests
	case fiber.StatusServiceUnavailable:
		return CodeServiceUnavailable
	default:
		return CodeInternalError
	}
}

