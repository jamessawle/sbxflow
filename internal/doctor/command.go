package doctor

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// CommandOutput preserves the observable result of one subprocess invocation.
type CommandOutput struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Err      error
}

// CommandRunner locates and invokes executables without a shell.
type CommandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, executable string, args ...string) CommandOutput
}

// ExecCommandRunner runs commands through os/exec with a per-command timeout.
type ExecCommandRunner struct {
	Timeout time.Duration
}

// LookPath resolves file using the process PATH.
func (ExecCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Run invokes executable directly and keeps stdout and stderr separate.
func (r ExecCommandRunner) Run(ctx context.Context, executable string, args ...string) CommandOutput {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	var stdout, stderr bytes.Buffer
	command := exec.CommandContext(ctx, executable, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	output := CommandOutput{
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
