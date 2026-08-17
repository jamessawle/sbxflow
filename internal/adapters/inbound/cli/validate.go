package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jamessawle/sbxflow/internal/application/validation"
	"github.com/spf13/cobra"
)

var errValidationFailed = errors.New("configuration validation failed")

// ValidateRunner validates the nearest repository declaration.
type ValidateRunner interface {
	Run(ctx context.Context, start string) validation.Report
}

// NewValidateCommand returns a command that validates a repository declaration.
func NewValidateCommand(runner ValidateRunner) *cobra.Command {
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
				cmd.Root().SilenceErrors = true
				return errValidationFailed
			}
			return nil
		},
	}
}

func renderValidationReport(command *cobra.Command, report validation.Report) {
	writer := io.Writer(command.OutOrStdout())
	if !report.Valid() {
		writer = command.ErrOrStderr()
	}

	declaration := report.Declaration
	if declaration == "" {
		declaration = "unavailable"
	}
	_, _ = fmt.Fprintf(writer, "Declaration: %s\n\n", declaration)

	if report.Linked.Configuration.Version == 0 {
		_, _ = fmt.Fprintln(writer, "Derived State: unavailable")
	} else {
		_, _ = fmt.Fprintln(writer, "Derived State:")
		_, _ = fmt.Fprintln(writer, "  Kit:")
		if len(report.Linked.Trust.AllowedSources) == 0 {
			_, _ = fmt.Fprintln(writer, "    Allowed Sources: []")
		} else {
			_, _ = fmt.Fprintln(writer, "    Allowed Sources:")
			for _, source := range report.Linked.Trust.AllowedSources {
				_, _ = fmt.Fprintf(writer, "      - %s\n", source)
			}
		}
		_, _ = fmt.Fprintf(writer, "    Local Kits Allowed: %t\n", report.Linked.Trust.AllowLocalKits)
	}

	state := "pass"
	if !report.Valid() {
		state = "fail"
	}
	_, _ = fmt.Fprintf(writer, "\nValidation:\n  State: %s\n", state)
	findings := validationFindings(report)
	if len(findings) == 0 {
		_, _ = fmt.Fprintln(writer, "  Findings: []")
		return
	}
	_, _ = fmt.Fprintln(writer, "  Findings:")
	for _, finding := range findings {
		lines := strings.Split(finding, "\n")
		_, _ = fmt.Fprintf(writer, "    - %s\n", lines[0])
		for _, line := range lines[1:] {
			_, _ = fmt.Fprintf(writer, "      %s\n", line)
		}
	}
}

func validationFindings(report validation.Report) []string {
	findings := make([]string, 0, len(report.Errors))
	for _, err := range report.Errors {
		findings = append(findings, err.Error())
	}

	findingIndex := 0
	for _, result := range report.LocalKits {
		if result.Err == nil {
			continue
		}
		if findingIndex < len(findings) && result.Diagnostics != "" {
			findings[findingIndex] += ": " + result.Diagnostics
		}
		findingIndex++
	}
	return findings
}
