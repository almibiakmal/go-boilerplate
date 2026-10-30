package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"maj-core-srvc/src/internal/bootstrap"
	"maj-core-srvc/src/internal/config"
	"maj-core-srvc/src/internal/infra/logger"
)

// cfg diisi oleh PersistentPreRunE sebelum command apa pun dijalankan.
var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:           "app",
	Short:         "My service",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg = c

		logger.Setup(cfg.Log.Level, cfg.Log.Format, cfg.Log.File)
		return nil
	},
}

func Execute() {
	defer func() {
		_ = logger.Close()
	}()
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}

// runApp membuat bootstrap App, menangani signal (SIGINT/SIGTERM),
// lalu memastikan resource ditutup setelah run selesai.
func runApp(run func(ctx context.Context, app *bootstrap.App) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			slog.Error("close app", "err", err)
		}
	}()

	return run(ctx, app)
}

func init() {
	rootCmd.AddCommand(httpCmd, eventCmd, migrateCmd, seedCmd)
}