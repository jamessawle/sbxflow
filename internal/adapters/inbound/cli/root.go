// Package cli adapts Cobra commands and terminal streams to application and
// domain services.
package cli

import (
	"context"
	"fmt"
	"io"

	buildinfo "github.com/jamessawle/sbxflow/internal/ports/buildInfo"
	"github.com/spf13/cobra"
)

const shortCommitLength = 7

// Streams contains the process streams used by the command tree.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Invocation contains the process inputs used to execute the command tree.
type Invocation struct {
	Args      []string
	Streams   Streams
	BuildInfo buildinfo.Info
}

// NewRootCommand returns a fresh sbxflow root command.
func NewRootCommand(streams Streams, commands ...*cobra.Command) *cobra.Command {
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
	root.AddCommand(help)
	root.AddCommand(commands...)

	return root
}

// Execute runs a fresh root command.
func Execute(ctx context.Context, invocation Invocation, commands ...*cobra.Command) error {
	root := NewRootCommand(invocation.Streams, commands...)
	root.Version = formatVersion(invocation.BuildInfo)
	root.SetArgs(invocation.Args)

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
