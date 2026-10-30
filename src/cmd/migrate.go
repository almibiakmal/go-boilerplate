package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// Group: `app migrate`
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

func newMigrator() (*migrate.Migrate, error) {
	dbURL := cfg.Database.URL
	if strings.EqualFold(cfg.Database.Driver, "mysql") && !strings.HasPrefix(dbURL, "mysql://") {
		dbURL = "mysql://" + dbURL
	}
	m, err := migrate.New(cfg.MigrationsPath, dbURL)
	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}
	return m, nil
}

func closeMigrator(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		slog.Warn("close migration source", "err", srcErr)
	}
	if dbErr != nil {
		slog.Warn("close migration database", "err", dbErr)
	}
}

// `app migrate up`
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrator()
		if err != nil {
			return err
		}
		defer closeMigrator(m)

		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Fprintln(cmd.OutOrStdout(), "no change")
				return nil
			}
			return fmt.Errorf("migrate up: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "migrate up: done")
		return nil
	},
}

// `app migrate down --steps 1`
var migrateDownSteps int

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back migrations (default: 1 step)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateDownSteps < 1 {
			return errors.New("--steps must be >= 1")
		}

		m, err := newMigrator()
		if err != nil {
			return err
		}
		defer closeMigrator(m)

		if err := m.Steps(-migrateDownSteps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Fprintln(cmd.OutOrStdout(), "no change")
				return nil
			}
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "migrate down: %d step(s) done\n", migrateDownSteps)
		return nil
	},
}

// `app migrate version`
var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrator()
		if err != nil {
			return err
		}
		defer closeMigrator(m)

		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Fprintln(cmd.OutOrStdout(), "version=none")
				return nil
			}
			return fmt.Errorf("get version: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%v\n", v, dirty)
		return nil
	},
}

// `app migrate force 20260101120000`
// Dipakai untuk memperbaiki state "dirty" setelah migrasi gagal di tengah jalan.
var migrateForceCmd = &cobra.Command{
	Use:   "force VERSION",
	Short: "Set version manually and clear dirty state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version %q: %w", args[0], err)
		}

		m, err := newMigrator()
		if err != nil {
			return err
		}
		defer closeMigrator(m)

		if err := m.Force(v); err != nil {
			return fmt.Errorf("force version: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "forced version to %d\n", v)
		return nil
	},
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// `app migrate create add_users_table`
var migrateCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a new pair of .up.sql and .down.sql files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(args[0]), "_"), "_")
		if name == "" {
			return errors.New("invalid migration name")
		}

		dir := strings.TrimPrefix(cfg.MigrationsPath, "file://")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create migrations dir: %w", err)
		}

		ts := time.Now().UTC().Format("20060102150405")
		for _, kind := range []string{"up", "down"} {
			path := filepath.Join(dir, fmt.Sprintf("%s_%s.%s.sql", ts, name, kind))
			if err := os.WriteFile(path, []byte("-- "+kind+"\n"), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "created", path)
		}
		return nil
	},
}

func init() {
	migrateDownCmd.Flags().IntVar(&migrateDownSteps, "steps", 1, "number of migrations to roll back")

	migrateCmd.AddCommand(
		migrateUpCmd,
		migrateDownCmd,
		migrateVersionCmd,
		migrateForceCmd,
		migrateCreateCmd,
	)
}