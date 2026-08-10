package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
	"github.com/spf13/cobra"
)

// DestroyRunner permanently removes the repository's declared sandbox.
type DestroyRunner interface {
	Run(ctx context.Context, start string, force bool, streams lifecycle.Streams) error
}

func newDestroyCommand(runner DestroyRunner) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Permanently remove the repository's Docker Sandbox",
		Long: "Discover the nearest sbxflow.yaml and permanently remove its declared Docker Sandbox and persisted state.\n" +
			"Unlike down, destroy cannot be undone. Docker asks for confirmation by default; --force skips confirmation and permits removal with an active session.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine working directory: %w", err)
			}
			err = runner.Run(cmd.Context(), workingDirectory, force, lifecycle.Streams{
				In:  cmd.InOrStdin(),
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
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation and remove a sandbox with an active session")
	return cmd
}
