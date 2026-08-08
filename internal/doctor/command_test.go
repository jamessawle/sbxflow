package doctor

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestExecCommandRunnerPreservesStreamsAndExitStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture uses POSIX sh")
	}

	runner := ExecCommandRunner{Timeout: time.Second}
	output := runner.Run(context.Background(), "sh", "-c", "printf stdout; printf stderr >&2; exit 7")

	if string(output.Stdout) != "stdout" {
		t.Errorf("stdout = %q, want %q", output.Stdout, "stdout")
	}
	if string(output.Stderr) != "stderr" {
		t.Errorf("stderr = %q, want %q", output.Stderr, "stderr")
	}
	if output.ExitCode != 7 {
		t.Errorf("exit code = %d, want 7", output.ExitCode)
	}
	if output.Err == nil {
		t.Error("error = nil, want exit error")
	}
}

func TestExecCommandRunnerTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture uses POSIX sh")
	}

	runner := ExecCommandRunner{Timeout: 10 * time.Millisecond}
	output := runner.Run(context.Background(), "sh", "-c", "sleep 1")

	if !errors.Is(output.Err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context deadline exceeded", output.Err)
	}
	if output.ExitCode == 0 {
		t.Fatal("exit code = 0, want non-zero")
	}
}

func TestExecCommandRunnerLookPath(t *testing.T) {
	runner := ExecCommandRunner{}
	path, err := runner.LookPath(os.Args[0])
	if err != nil {
		t.Fatalf("LookPath(test executable): %v", err)
	}
	if path == "" {
		t.Fatal("LookPath(test executable) returned an empty path")
	}
}
