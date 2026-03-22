package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/razobeckett/goco/internal/cli"
)

var (
	version = ""
	commit  = ""
)

func main() {
	if err := fang.Execute(
		context.Background(),
		cli.NewRootCmd(),
		fang.WithVersion(version),
		fang.WithCommit(commit),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
