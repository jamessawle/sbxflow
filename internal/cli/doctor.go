package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jamessawle/sbxflow/internal/doctor"
	"github.com/spf13/cobra"
)

var errDoctorFailed = errors.New("one or more required doctor checks failed")

// DoctorRunner runs the ordered system diagnostic checks.
type DoctorRunner interface {
	Run(ctx context.Context) doctor.Report
}

func newDoctorCommand(runner DoctorRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check Docker Sandboxes system health and configuration",
		Long: "Check whether Docker Sandboxes is compatible, healthy, and safely configured.\n" +
			"Doctor inspects system-level state without reading sbxflow.yaml or changing the system.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			report := runner.Run(cmd.Context())
			renderDoctorReport(cmd, report)
			if report.Failed() {
				return errDoctorFailed
			}
			return nil
		},
	}
}

func renderDoctorReport(command *cobra.Command, report doctor.Report) {
	for _, result := range report.Results {
		fmt.Fprintf(
			command.OutOrStdout(),
			"[%s] %s: %s\n",
			strings.ToUpper(string(result.Status)),
			result.ID,
			result.Summary,
		)
		if result.Guidance != "" {
			fmt.Fprintf(command.OutOrStdout(), "  %s\n", result.Guidance)
		}
	}
}
