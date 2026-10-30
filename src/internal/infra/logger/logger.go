package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	userIDKey    ctxKey = "user_id"
)

var (
	fileMu  sync.Mutex
	logFile *os.File
)

// ContextHandler adalah wrapper slog.Handler yang otomatis mengekstrak
// request_id dan user_id dari context.Context ke setiap baris log tanpa alokasi logger baru.
type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
			r.AddAttrs(slog.String("request_id", reqID))
		}
		if usrID, ok := ctx.Value(userIDKey).(string); ok && usrID != "" {
			r.AddAttrs(slog.String("user_id", usrID))
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}

// Setup menginisialisasi global slog logger berdasarkan level ("debug", "info", "warn", "error")
// dan format ("json" atau "text"), serta opsional menulis ke file log.
func Setup(levelStr, format string, filePath ...string) {
	level := parseLevel(levelStr)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var writer io.Writer = os.Stdout
	if len(filePath) > 0 && strings.TrimSpace(filePath[0]) != "" {
		path := strings.TrimSpace(filePath[0])
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to create log directory %q: %v\n", filepath.Dir(path), err)
		} else if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to open log file %q: %v\n", path, err)
		} else {
			fileMu.Lock()
			if logFile != nil {
				_ = logFile.Close()
			}
			logFile = f
			fileMu.Unlock()
			writer = io.MultiWriter(os.Stdout, f)
		}
	}

	var baseHandler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text", "console":
		baseHandler = slog.NewTextHandler(writer, opts)
	default:
		baseHandler = slog.NewJSONHandler(writer, opts)
	}

	slog.SetDefault(slog.New(&ContextHandler{Handler: baseHandler}))
}

// Close menutup file handle log jika LOG_FILE digunakan.
func Close() error {
	fileMu.Lock()
	defer fileMu.Unlock()
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}

// WithRequestID menyimpan request_id ke dalam context.Context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext mengambil request_id dari context.Context.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val, _ := ctx.Value(requestIDKey).(string)
	return val
}

// WithUserID menyimpan user_id ke dalam context.Context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext mengambil user_id dari context.Context.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val, _ := ctx.Value(userIDKey).(string)
	return val
}

// Debug mencatat log level DEBUG dengan context (otomatis menyertakan request_id/user_id via ContextHandler).
func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

// Info mencatat log level INFO dengan context (otomatis menyertakan request_id/user_id via ContextHandler).
func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

// Warn mencatat log level WARN dengan context (otomatis menyertakan request_id/user_id via ContextHandler).
func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

// Error mencatat log level ERROR dengan context (otomatis menyertakan request_id/user_id via ContextHandler).
func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
