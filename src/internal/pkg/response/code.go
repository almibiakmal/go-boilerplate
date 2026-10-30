package response

import "github.com/gofiber/fiber/v2"

// Code merepresentasikan standar kode respon aplikasi (business/system code).
type Code string

const (
	// Success codes
	CodeSuccess Code = "SUCCESS"
	CodeCreated Code = "CREATED"

	// Client error codes
	CodeBadRequest      Code = "BAD_REQUEST"
	CodeValidationError Code = "VALIDATION_ERROR"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeDuplicateEntry  Code = "DUPLICATE_ENTRY"
	CodeTooManyRequests Code = "TOO_MANY_REQUESTS"

	// Server / Infra error codes
	CodeInternalError      Code = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeDatabaseError      Code = "DATABASE_ERROR"
)

// CodeMeta menyimpan mapping HTTP status code dan pesan default untuk setiap Code.
type CodeMeta struct {
	HTTPStatus     int
	DefaultMessage string
}

var codeRegistry = map[Code]CodeMeta{
	CodeSuccess: {
		HTTPStatus:     fiber.StatusOK,
		DefaultMessage: "Request processed successfully",
	},
	CodeCreated: {
		HTTPStatus:     fiber.StatusCreated,
		DefaultMessage: "Resource created successfully",
	},
	CodeBadRequest: {
		HTTPStatus:     fiber.StatusBadRequest,
		DefaultMessage: "Invalid request parameters or payload",
	},
	CodeValidationError: {
		HTTPStatus:     fiber.StatusUnprocessableEntity,
		DefaultMessage: "Validation failed",
	},
	CodeUnauthorized: {
		HTTPStatus:     fiber.StatusUnauthorized,
		DefaultMessage: "Unauthorized access",
	},
	CodeForbidden: {
		HTTPStatus:     fiber.StatusForbidden,
		DefaultMessage: "You do not have permission to access this resource",
	},
	CodeNotFound: {
		HTTPStatus:     fiber.StatusNotFound,
		DefaultMessage: "Requested resource not found",
	},
	CodeConflict: {
		HTTPStatus:     fiber.StatusConflict,
		DefaultMessage: "Resource state conflict",
	},
	CodeDuplicateEntry: {
		HTTPStatus:     fiber.StatusConflict,
		DefaultMessage: "Resource already exists",
	},
	CodeTooManyRequests: {
		HTTPStatus:     fiber.StatusTooManyRequests,
		DefaultMessage: "Too many requests, please try again later",
	},
	CodeInternalError: {
		HTTPStatus:     fiber.StatusInternalServerError,
		DefaultMessage: "An unexpected internal server error occurred",
	},
	CodeServiceUnavailable: {
		HTTPStatus:     fiber.StatusServiceUnavailable,
		DefaultMessage: "Service is temporarily unavailable",
	},
	CodeDatabaseError: {
		HTTPStatus:     fiber.StatusInternalServerError,
		DefaultMessage: "A database error occurred",
	},
}

// Resolve mengambil metadata (HTTPStatus & DefaultMessage) berdasarkan Code.
// Jika Code belum terdaftar, fallback ke CodeInternalError (500).
func (c Code) Resolve() CodeMeta {
	if meta, ok := codeRegistry[c]; ok {
		return meta
	}
	return codeRegistry[CodeInternalError]
}

// HTTPStatus mengembalikan HTTP status code yang berkorespondensi dengan Code.
func (c Code) HTTPStatus() int {
	return c.Resolve().HTTPStatus
}

// DefaultMessage mengembalikan pesan default untuk Code tersebut.
func (c Code) DefaultMessage() string {
	return c.Resolve().DefaultMessage
}

