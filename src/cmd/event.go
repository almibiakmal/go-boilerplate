package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"maj-core-srvc/src/internal/bootstrap"
)

// Group: `app event`
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Event related commands",
}

// `app event serve`
var eventServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run event consumers and outbox relay",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApp(func(ctx context.Context, app *bootstrap.App) error {
			return app.RunEvent(ctx)
		})
	},
}

func init() {
	eventCmd.AddCommand(eventServeCmd)
}