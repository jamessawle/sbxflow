package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
	"github.com/spf13/cobra"
)

// UpRunner validates and enters the repository's declared sandbox.
type UpRunner interface {
	Run(ctx context.Context, start string, options lifecycle.UpOptions, streams lifecycle.Streams) (lifecycle.ValidationReport, error)
}

// NewUpCommand returns a command that creates or enters a sandbox.
func NewUpCommand(runner UpRunner) *cobra.Command {
	var recreate bool
	var force bool
	command := &cobra.Command{
		Use:   "up",
		Short: "Create or enter the repository's Docker Sandbox",
		Long: "Discover and validate the nearest sbxflow.yaml, then interactively create or enter its declared Docker Sandbox.\n" +
			"An existing named sandbox is entered without reconciling its workspace or kits with the current declaration.\n" +
			"With --recreate, an existing sandbox, its persisted state, and work stored only inside it are force-removed before its replacement is created and entered.\n" +
			"Recreating a running sandbox requires confirmation because it can terminate other attached terminal sessions.\n" +
			"With --recreate, --force bypasses that confirmation despite the permanent state loss and risk to attached sessions.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if force && !recreate {
				return fmt.Errorf("--force requires --recreate")
			}
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine working directory: %w", err)
			}
			streams := lifecycle.Streams{
				In:  cmd.InOrStdin(),
				Out: cmd.OutOrStdout(),
				Err: cmd.ErrOrStderr(),
			}
			report, err := runner.Run(cmd.Context(), workingDirectory, lifecycle.UpOptions{Recreate: recreate, Force: force, Confirmer: recreationConfirmer{}}, streams)
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
	command.Flags().BoolVar(&recreate, "recreate", false, "Force-recreate the sandbox, confirming first if it is running")
	command.Flags().BoolVar(&force, "force", false, "With --recreate, skip running-sandbox confirmation despite permanent state loss and risk to attached sessions")
	return command
}

type recreationConfirmer struct{}

func (recreationConfirmer) ConfirmRunningSandboxRecreation(name string, streams lifecycle.Streams) (bool, error) {
	if streams.In == nil || streams.Err == nil {
		return false, errors.New("confirmation input or error stream is unavailable")
	}
	_, _ = fmt.Fprintf(streams.Err, "Warning: recreating running sandbox %q permanently removes its persisted state and work stored only inside it, and can terminate other attached terminal sessions.\nContinue? [y/N] ", name)
	response, err := readConfirmationLine(streams.In)
	if err != nil {
		return false, fmt.Errorf("read confirmation response: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func readConfirmationLine(input io.Reader) (string, error) {
	var response strings.Builder
	var buffer [1]byte
	for {
		count, err := input.Read(buffer[:])
		if count == 1 {
			if buffer[0] != '\n' {
				response.WriteByte(buffer[0])
			}
		}
		if count == 1 && buffer[0] == '\n' && (err == nil || errors.Is(err, io.EOF)) {
			return response.String(), nil
		}
		if err != nil {
			if errors.Is(err, io.EOF) && response.Len() > 0 {
				return response.String(), nil
			}
			return "", err
		}
		if count == 0 {
			return "", io.ErrNoProgress
		}
	}
}
