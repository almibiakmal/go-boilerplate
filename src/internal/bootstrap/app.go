package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/config"
	"maj-core-srvc/src/internal/handler"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/repository"
	"maj-core-srvc/src/internal/router"
	"maj-core-srvc/src/internal/service"
)

// App membungkus semua komponen utama aplikasi (config, koneksi DB, dan server HTTP).
type App struct {
	cfg  *config.Config
	db   *gorm.DB
	http *fiber.App
}

// New menginisialisasi koneksi database, transaction manager, repository, service, handler, middleware, dan router.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	// Setup Fiber instance
	fiberApp := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ReadTimeout:           cfg.HTTP.ReadTimeout,
		WriteTimeout:          cfg.HTTP.WriteTimeout,
		IdleTimeout:           cfg.HTTP.IdleTimeout,
		DisableStartupMessage: true,
		ErrorHandler:          middleware.ErrorHandler,
	})

	// Register global middlewares (RequestID, Recover, CORS, Logger, RateLimiter)
	middleware.Setup(fiberApp, cfg)

	// Wiring 3-Tier Dependencies: TxManager & Repositories -> Services -> Handlers
	txManager := database.NewTxManager(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.Auth)
	postSvc := service.NewPostService(txManager, postRepo, commentRepo)
	commentSvc := service.NewCommentService(txManager, postRepo, commentRepo)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authSvc)
	postHandler := handler.NewPostHandler(postSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)

	// Register HTTP Routes
	router.RegisterHTTP(fiberApp, router.Dependencies{
		Config:         cfg,
		HealthHandler:  healthHandler,
		AuthHandler:    authHandler,
		PostHandler:    postHandler,
		CommentHandler: commentHandler,
	})

	return &App{
		cfg:  cfg,
		db:   db,
		http: fiberApp,
	}, nil
}

// RunHTTP menjalankan server Fiber dan menangani graceful shutdown ketika context dibatalkan.
func (a *App) RunHTTP(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)
	errCh := make(chan error, 1)

	go func() {
		slog.Info("starting HTTP server", "addr", addr, "env", a.cfg.App.Env)
		if err := a.http.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down HTTP server gracefully...")
		if err := a.http.ShutdownWithTimeout(a.cfg.HTTP.ShutdownTimeout); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		slog.Info("HTTP server stopped")
		return nil
	case err := <-errCh:
		return fmt.Errorf("listen HTTP: %w", err)
	}
}

// RunEvent adalah placeholder untuk event consumer / outbox relay yang akan diimplementasikan nanti.
func (a *App) RunEvent(ctx context.Context) error {
	slog.Info("event worker started (no message broker configured yet; waiting for signal...)")
	<-ctx.Done()
	slog.Info("shutting down event worker gracefully...")
	return nil
}

// Close menutup resource aplikasi seperti connection pool database.
func (a *App) Close() error {
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err != nil {
			return err
		}
		slog.Info("closing database connection")
		return sqlDB.Close()
	}
	return nil
}
