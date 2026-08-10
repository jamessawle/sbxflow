package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jamessawle/sbxflow/internal/command"
	"github.com/jamessawle/sbxflow/internal/validation"
)

const defaultSandboxLookupTimeout = 30 * time.Second

// ErrValidationFailed identifies a lifecycle request stopped by repository
// validation.
var ErrValidationFailed = errors.New("configuration validation failed")

// Validator runs the complete repository validation pipeline.
type Validator interface {
	Run(ctx context.Context, start string) validation.Report
}

// Streams are attached to the interactive Docker agent process.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Runner validates, inspects, and enters the declared sandbox.
type Runner struct {
	Validation  Validator
	Commands    command.Runner
	Interactive command.InteractiveRunner
}

// NewDefaultRunner constructs production lifecycle dependencies.
func NewDefaultRunner() Runner {
	return Runner{
		Validation:  validation.NewDefaultRunner(),
		Commands:    command.ExecRunner{Timeout: defaultSandboxLookupTimeout},
		Interactive: command.InteractiveExecRunner{},
	}
}

// AttachedProcessError identifies a failure already rendered by the attached
// Docker process on the CLI error stream.
type AttachedProcessError struct {
	Err error
}

func (e AttachedProcessError) Error() string { return e.Err.Error() }
func (e AttachedProcessError) Unwrap() error { return e.Err }

// Run validates the repository, chooses the creation or existing-sandbox
// invocation, and remains attached until Docker exits.
func (r Runner) Run(ctx context.Context, start string, streams Streams) (validation.Report, error) {
	report := r.Validation.Run(ctx, start)
	if !report.Valid() {
		return report, ErrValidationFailed
	}
	if streams.Err != nil {
		fmt.Fprintf(streams.Err, "Configuration valid: %s\n", report.Declaration)
	}

	plan, err := NewPlan(report)
	if err != nil {
		return report, err
	}
	executable, err := r.Commands.LookPath("sbx")
	if err != nil {
		return report, fmt.Errorf("locate sbx for sandbox lookup: %w", err)
	}
	exists, err := sandboxExists(ctx, r.Commands, executable, plan.Name)
	if err != nil {
		return report, err
	}

	args := existingSandboxArguments(plan)
	if !exists {
		args = missingSandboxArguments(plan)
	}
	err = r.Interactive.Run(ctx, command.InteractiveInvocation{
		Executable:  executable,
		Args:        args,
		Environment: plan.Environment,
		Stdin:       streams.In,
		Stdout:      streams.Out,
		Stderr:      streams.Err,
	})
	if err != nil {
		return report, AttachedProcessError{Err: err}
	}
	return report, nil
}

func sandboxExists(ctx context.Context, runner command.Runner, executable, name string) (bool, error) {
	output := runner.Run(ctx, executable, "ls", "--quiet")
	if output.Err != nil {
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), ": "))
		if diagnostics != "" {
			return false, fmt.Errorf("list Docker Sandboxes with `sbx ls --quiet`: %s: %w", diagnostics, output.Err)
		}
		return false, fmt.Errorf("list Docker Sandboxes with `sbx ls --quiet`: %w", output.Err)
	}
	for _, existing := range strings.Split(string(output.Stdout), "\n") {
		if strings.TrimSpace(existing) != "" && existing == name {
			return true, nil
		}
	}
	return false, nil
}

func missingSandboxArguments(plan Plan) []string {
	args := []string{"run", "--name", plan.Name}
	for _, kit := range plan.Kits {
		args = append(args, "--kit", kit)
	}
	return append(args, plan.Agent, plan.Workspace)
}

func existingSandboxArguments(plan Plan) []string {
	return []string{"run", plan.Agent, "--name", plan.Name}
}

func nonempty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}
