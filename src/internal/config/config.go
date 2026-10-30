package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App            AppConfig
	HTTP           HTTPConfig
	Auth           AuthConfig
	RateLimit      RateLimitConfig
	Log            LogConfig
	Database       DatabaseConfig
	MigrationsPath string
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret      string
	JWTIssuer      string
	InternalAPIKey string
}

type RateLimitConfig struct {
	Enabled    bool
	Max        int
	Expiration time.Duration
}

type LogConfig struct {
	Level  string
	Format string
	File   string
}

type DatabaseConfig struct {
	Driver          string
	URL             string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Load membaca konfigurasi dari file .env (jika ada) dan environment variables.
func Load() (*Config, error) {
	// Abaikan error jika file .env tidak ditemukan (mis. di production container)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "maj-core-srvc"),
			Env:  getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("HTTP_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getEnvAsDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvAsDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("AUTH_JWT_SECRET", "change-me-in-production"),
			JWTIssuer:      getEnv("AUTH_JWT_ISSUER", "maj-core-srvc"),
			InternalAPIKey: getEnv("AUTH_INTERNAL_API_KEY", "internal-dev-key"),
		},
		RateLimit: RateLimitConfig{
			Enabled:    getEnvAsBool("RATE_LIMIT_ENABLED", true),
			Max:        getEnvAsInt("RATE_LIMIT_MAX", 100),
			Expiration: getEnvAsDuration("RATE_LIMIT_EXPIRATION", 1*time.Minute),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			File:   getEnv("LOG_FILE", ""),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "postgres"),
			URL:             getEnv("DB_URL", ""),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "maj_core_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 50),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://migrations"),
	}

	// Jika DB_URL belum di-set secara eksplisit, bangun otomatis berdasarkan DB_DRIVER
	if cfg.Database.URL == "" {
		url, err := buildDatabaseURL(cfg.Database)
		if err != nil {
			return nil, err
		}
		cfg.Database.URL = url
	}

	return cfg, nil
}

func buildDatabaseURL(db DatabaseConfig) (string, error) {
	switch db.Driver {
	case "postgres", "postgresql":
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
		), nil
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.User, db.Password, db.Host, db.Port, db.Name,
		), nil
	case "sqlite", "sqlite3":
		return db.Name, nil
	default:
		return "", fmt.Errorf("unsupported DB_DRIVER: %s", db.Driver)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvAsBool(key string, fallback bool) bool {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return fallback
	}
	return val
}
