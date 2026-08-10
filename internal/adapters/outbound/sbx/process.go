package sbx

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// Output is the captured sandbox port result.
type Output = sandboxport.Output

// Runner locates and invokes executables without a shell.
type Runner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, executable string, args ...string) Output
}

// ExecRunner runs commands through os/exec with a per-command timeout.
type ExecRunner struct {
	Timeout time.Duration
}

// LookPath resolves file using the process PATH.
func (ExecRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Run invokes executable directly and keeps stdout and stderr separate.
func (r ExecRunner) Run(ctx context.Context, executable string, args ...string) Output {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	var stdout, stderr bytes.Buffer
	invocation := exec.CommandContext(ctx, executable, args...)
	invocation.Stdout = &stdout
	invocation.Stderr = &stderr

	err := invocation.Run()
	output := Output{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: 0,
		Err:      err,
	}
	if err == nil {
		return output
	}

	output.ExitCode = -1
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		output.ExitCode = exitError.ExitCode()
	}
	if ctx.Err() != nil {
		output.Err = ctx.Err()
	}

	return output
}
