package main

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/jamessawle/sbxflow/internal/buildinfo"
	"github.com/jamessawle/sbxflow/internal/cli"
)

func main() {
	err := cli.Execute(
		context.Background(),
		os.Args[1:],
		cli.Streams{In: os.Stdin, Out: os.Stdout, Err: os.Stderr},
		buildinfo.Current(),
	)
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}
