package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"maj-core-srvc/src/internal/bootstrap"
)

// Group: `app http`
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP related commands",
}

// `app http serve`
var httpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run REST API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApp(func(ctx context.Context, app *bootstrap.App) error {
			return app.RunHTTP(ctx)
		})
	},
}

func init() {
	httpCmd.AddCommand(httpServeCmd)
}