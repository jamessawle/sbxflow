package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jamessawle/sbxflow/internal/validation"
	"github.com/spf13/cobra"
)

var errValidationFailed = errors.New("configuration validation failed")

// ValidateRunner validates the nearest repository declaration.
type ValidateRunner interface {
	Run(ctx context.Context, start string) validation.Report
}

func newValidateCommand(runner ValidateRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the repository's Docker Sandbox configuration",
		Long: "Discover and validate the nearest sbxflow.yaml repository declaration.\n" +
			"Remote Git and OCI references are checked offline; selected local kits are validated by Docker Sandboxes.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine working directory: %w", err)
			}
			report := runner.Run(cmd.Context(), workingDirectory)
			renderValidationReport(cmd, report)
			if !report.Valid() {
				return errValidationFailed
			}
			return nil
		},
	}
}

func renderValidationReport(command *cobra.Command, report validation.Report) {
	if report.Declaration != "" {
		fmt.Fprintf(command.OutOrStdout(), "Declaration: %s\n", report.Declaration)
	}
	for _, result := range report.LocalKits {
		if result.Target.Path == "" {
			continue
		}
		status := "VALID"
		writer := command.OutOrStdout()
		if !result.Valid {
			status = "INVALID"
			writer = command.ErrOrStderr()
		}
		fmt.Fprintf(writer, "[%s] local kit %s/%s: %s\n", status, result.Target.Source, result.Target.Kit, result.Target.Path)
		if result.Diagnostics != "" {
			fmt.Fprintf(writer, "  %s\n", result.Diagnostics)
		}
	}
	if report.Linked.Configuration.Version != 0 {
		fmt.Fprintf(command.OutOrStdout(), "kit.allowedSources: [%s]\n", strings.Join(report.Linked.Trust.AllowedSources, ", "))
		fmt.Fprintf(command.OutOrStdout(), "kit.allowLocalKits: %s\n", strconv.FormatBool(report.Linked.Trust.AllowLocalKits))
	}
	for _, err := range report.Errors {
		fmt.Fprintf(command.ErrOrStderr(), "validation error: %s\n", err)
	}
}
