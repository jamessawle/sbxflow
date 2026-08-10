package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
	"github.com/spf13/cobra"
)

// UpRunner validates and enters the repository's declared sandbox.
type UpRunner interface {
	Run(ctx context.Context, start string, options lifecycle.UpOptions, streams lifecycle.Streams) (lifecycle.ValidationReport, error)
}

func newUpCommand(runner UpRunner) *cobra.Command {
	var recreate bool
	command := &cobra.Command{
		Use:   "up",
		Short: "Create or enter the repository's Docker Sandbox",
		Long: "Discover and validate the nearest sbxflow.yaml, then interactively create or enter its declared Docker Sandbox.\n" +
			"An existing named sandbox is entered without reconciling its workspace or kits with the current declaration.\n" +
			"With --recreate, an existing sandbox and its persisted state are force-removed before its replacement is created and entered.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine working directory: %w", err)
			}
			report, err := runner.Run(cmd.Context(), workingDirectory, lifecycle.UpOptions{Recreate: recreate}, lifecycle.Streams{
				In:  cmd.InOrStdin(),
				Out: cmd.OutOrStdout(),
				Err: cmd.ErrOrStderr(),
			})
			if errors.Is(err, lifecycle.ErrValidationFailed) {
				renderValidationReport(cmd, report)
				cmd.Root().SilenceErrors = true
				return errValidationFailed
			}
			var attached lifecycle.AttachedProcessError
			if errors.As(err, &attached) {
				cmd.Root().SilenceErrors = true
			}
			return err
		},
	}
	command.Flags().BoolVar(&recreate, "recreate", false, "Force-remove an existing sandbox and its persisted state before replacement")
	return command
}
