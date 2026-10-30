package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/config"
)

// PoolStats merepresentasikan metrik status connection pool database saat runtime.
type PoolStats struct {
	MaxOpenConnections int   `json:"max_open_connections"`
	OpenConnections    int   `json:"open_connections"`
	InUse              int   `json:"in_use"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDurationMs     int64 `json:"wait_duration_ms"`
}

// New membuat koneksi GORM secara dinamis berdasarkan DB_DRIVER ("postgres", "mysql", "sqlite")
// dan mengonfigurasi database/sql connection pool.
func New(ctx context.Context, cfg config.DatabaseConfig) (*gorm.DB, error) {
	dialector, err := resolveDialector(cfg)
	if err != nil {
		return nil, err
	}

	gormCfg := &gorm.Config{
		Logger:                 NewSlogGormLogger(200 * time.Millisecond),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open gorm connection (%s): %w", cfg.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB pool: %w", err)
	}

	// Setup Connection Pool dengan fallback nilai aman jika env bernilai <= 0
	configureConnectionPool(sqlDB, cfg)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database pool (%s): %w", cfg.Driver, err)
	}

	slog.Info("database connection pool initialized",
		"driver", cfg.Driver,
		"host", cfg.Host,
		"db", cfg.Name,
		"max_idle_conns", cfg.MaxIdleConns,
		"max_open_conns", cfg.MaxOpenConns,
		"conn_max_lifetime", cfg.ConnMaxLifetime.String(),
		"conn_max_idle_time", cfg.ConnMaxIdleTime.String(),
	)

	return db, nil
}

// configureConnectionPool menerapkan konfigurasi connection pooling pada underlying *sql.DB.
func configureConnectionPool(sqlDB *sql.DB, cfg config.DatabaseConfig) {
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}

	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 50
	}

	maxLifetime := cfg.ConnMaxLifetime
	if maxLifetime <= 0 {
		maxLifetime = 30 * time.Minute
	}

	maxIdleTime := cfg.ConnMaxIdleTime
	if maxIdleTime <= 0 {
		maxIdleTime = 10 * time.Minute
	}

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)
}

// Stats mengembalikan metrik connection pool saat ini dari instance *gorm.DB.
func Stats(db *gorm.DB) (*PoolStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	s := sqlDB.Stats()
	return &PoolStats{
		MaxOpenConnections: s.MaxOpenConnections,
		OpenConnections:    s.OpenConnections,
		InUse:              s.InUse,
		Idle:               s.Idle,
		WaitCount:          s.WaitCount,
		WaitDurationMs:     s.WaitDuration.Milliseconds(),
	}, nil
}

func resolveDialector(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	switch driver {
	case "postgres", "postgresql":
		return postgres.Open(cfg.URL), nil
	case "mysql":
		dsn := strings.TrimPrefix(cfg.URL, "mysql://")
		return mysql.Open(dsn), nil
	case "sqlite", "sqlite3":
		return sqlite.Open(cfg.URL), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q (supported: postgres, mysql, sqlite)", cfg.Driver)
	}
}
