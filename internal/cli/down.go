package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jamessawle/sbxflow/internal/lifecycle"
	"github.com/spf13/cobra"
)

// DownRunner stops the repository's declared sandbox without removing it.
type DownRunner interface {
	Run(ctx context.Context, start string, streams lifecycle.Streams) error
}

func newDownCommand(runner DownRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Stop the repository's Docker Sandbox without removing it",
		Long: "Discover the nearest sbxflow.yaml and stop its declared Docker Sandbox without removing it.\n" +
			"Only the declaration version and sandbox name are resolved; persisted sandbox state is preserved for a later up.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine working directory: %w", err)
			}
			err = runner.Run(cmd.Context(), workingDirectory, lifecycle.Streams{
				Out: cmd.OutOrStdout(),
				Err: cmd.ErrOrStderr(),
			})
			var attached lifecycle.AttachedProcessError
			if errors.As(err, &attached) {
				cmd.Root().SilenceErrors = true
			}
			return err
		},
	}
}
