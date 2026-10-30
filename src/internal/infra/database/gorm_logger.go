package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SlogGormLogger mengimplementasikan gormlogger.Interface menggunakan log/slog
// agar seluruh query database, slow query, dan error tercatat terstruktur beserta request_id.
type SlogGormLogger struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  gormlogger.LogLevel
}

func NewSlogGormLogger(slowThreshold time.Duration) *SlogGormLogger {
	if slowThreshold <= 0 {
		slowThreshold = 200 * time.Millisecond
	}
	return &SlogGormLogger{
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  gormlogger.Info,
	}
}

func (l *SlogGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *SlogGormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Info {
		slog.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogGormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Warn {
		slog.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogGormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Error {
		slog.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	isError := err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError)
	isSlow := elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn
	isDebug := l.LogLevel == gormlogger.Info && slog.Default().Enabled(ctx, slog.LevelDebug)

	// Hindari pemanggilan fc() (interpolasi string SQL yang mahal) jika tidak ada log yang perlu ditulis.
	if !isError && !isSlow && !isDebug {
		return
	}

	sql, rows := fc()
	attrs := []any{
		"elapsed_ms", elapsed.Milliseconds(),
		"rows", rows,
		"sql", sql,
	}

	switch {
	case isError:
		slog.ErrorContext(ctx, "gorm query error", append(attrs, "err", err)...)
	case isSlow:
		slog.WarnContext(ctx, "gorm slow query", append(attrs, "threshold_ms", l.SlowThreshold.Milliseconds())...)
	case isDebug:
		slog.DebugContext(ctx, "gorm query", attrs...)
	}
}
