// Package cli constructs and executes the sbxflow command tree.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/jamessawle/sbxflow/internal/buildinfo"
	"github.com/jamessawle/sbxflow/internal/doctor"
	"github.com/jamessawle/sbxflow/internal/validation"
	"github.com/spf13/cobra"
)

const shortCommitLength = 7

// Streams contains the process streams used by the command tree.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// NewRootCommand returns a fresh sbxflow root command.
func NewRootCommand(streams Streams, doctorRunner DoctorRunner, validateRunners ...ValidateRunner) *cobra.Command {
	root := &cobra.Command{
		Use:   "sbxflow",
		Short: "Apply a repository's Docker Sandbox configuration and lifecycle",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)
	root.CompletionOptions.DisableDefaultCmd = true
	root.SilenceUsage = true

	help := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about sbxflow",
		Args:  cobra.MaximumNArgs(1),
	}
	help.RunE = func(_ *cobra.Command, args []string) error {
		if len(args) == 0 {
			root.InitDefaultHelpFlag()
			root.InitDefaultVersionFlag()
			return root.Help()
		}

		target, remaining, err := root.Find(args)
		if err != nil || target == root || len(remaining) != 0 {
			return fmt.Errorf("unknown help topic %q", args[0])
		}

		target.InitDefaultHelpFlag()
		target.InitDefaultVersionFlag()
		return target.Help()
	}
	root.SetHelpCommand(help)
	validateRunner := ValidateRunner(validation.NewDefaultRunner())
	if len(validateRunners) > 0 {
		validateRunner = validateRunners[0]
	}
	root.AddCommand(help, newDoctorCommand(doctorRunner), newValidateCommand(validateRunner))

	return root
}

// Execute runs a fresh root command.
func Execute(ctx context.Context, args []string, streams Streams, info buildinfo.Info) error {
	root := NewRootCommand(streams, doctor.NewDefaultRunner(), validation.NewDefaultRunner())
	root.Version = formatVersion(info)
	root.SetArgs(args)

	return root.ExecuteContext(ctx)
}

func formatVersion(info buildinfo.Info) string {
	version := info.Version
	if version == "" {
		version = "development"
	}
	if len(info.Commit) >= shortCommitLength {
		version += " (" + info.Commit[:shortCommitLength] + ")"
	}

	return version
}
