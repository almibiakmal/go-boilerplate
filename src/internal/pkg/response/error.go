package response

import "fmt"

// FieldError merepresentasikan detail error per-field (misal untuk validasi form/JSON).
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AppError adalah custom error aplikasi yang membawa Code standar, pesan user-facing,
// detail validasi opsional, serta underlying error untuk kebutuhan logging.
type AppError struct {
	Code    Code         `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
	Err     error        `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError membuat AppError baru dengan Code tertentu. Jika message kosong, akan memakai DefaultMessage dari Code.
func NewError(code Code, message string) *AppError {
	if message == "" {
		message = code.DefaultMessage()
	}
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WrapError membungkus error asli (misal dari DB/library) ke dalam AppError.
func WrapError(code Code, message string, err error) *AppError {
	if message == "" {
		message = code.DefaultMessage()
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewValidationError membuat AppError khusus validasi beserta daftar field yang gagal.
func NewValidationError(details []FieldError) *AppError {
	return &AppError{
		Code:    CodeValidationError,
		Message: CodeValidationError.DefaultMessage(),
		Details: details,
	}
}

