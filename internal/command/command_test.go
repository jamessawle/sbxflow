package command

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestExecRunnerPreservesStreamsAndExitStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture uses POSIX sh")
	}

	runner := ExecRunner{Timeout: time.Second}
	output := runner.Run(context.Background(), "sh", "-c", "printf stdout; printf stderr >&2; exit 7")

	if string(output.Stdout) != "stdout" || string(output.Stderr) != "stderr" || output.ExitCode != 7 || output.Err == nil {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestExecRunnerTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture uses POSIX sh")
	}

	output := (ExecRunner{Timeout: 10 * time.Millisecond}).Run(context.Background(), "sh", "-c", "sleep 1")
	if !errors.Is(output.Err, context.DeadlineExceeded) || output.ExitCode == 0 {
		t.Fatalf("unexpected timeout output: %#v", output)
	}
}

func TestExecRunnerLookPath(t *testing.T) {
	path, err := (ExecRunner{}).LookPath(os.Args[0])
	if err != nil || path == "" {
		t.Fatalf("LookPath(test executable) = %q, %v", path, err)
	}
}
