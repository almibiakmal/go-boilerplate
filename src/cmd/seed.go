package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"maj-core-srvc/src/internal/infra/database"
)

// `app seed`
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with initial sample data (users, posts, comments)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := database.New(ctx, cfg.Database)
		if err != nil {
			return fmt.Errorf("connect database for seeding: %w", err)
		}
		defer func() {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}()

		if err := database.SeedAll(ctx, db); err != nil {
			return fmt.Errorf("seed database: %w", err)
		}

		slog.Info("database seeding completed successfully")
		fmt.Fprintln(cmd.OutOrStdout(), "seed: done")
		return nil
	},
}

