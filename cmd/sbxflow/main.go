package main

import (
	"context"
	"os"

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
		os.Exit(1)
	}
}
