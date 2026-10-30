package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maj-core-srvc/src/internal/infra/logger"
)

func TestContextHandler_InjectsRequestIDAndUserID(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(&logger.ContextHandler{Handler: baseHandler}))

	ctx := context.Background()
	ctx = logger.WithRequestID(ctx, "req-abc-123")
	ctx = logger.WithUserID(ctx, "usr-999")

	logger.Info(ctx, "order processed", "order_id", 42)

	var entry map[string]any
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "order processed", entry["msg"])
	assert.Equal(t, "req-abc-123", entry["request_id"])
	assert.Equal(t, "usr-999", entry["user_id"])
	assert.Equal(t, float64(42), entry["order_id"])
}

