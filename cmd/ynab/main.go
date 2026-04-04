package main

import (
	"fmt"
	"os"

	"github.com/bwishan/ynab-cli/internal/cli"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	app := cli.New(version, commit, buildDate)
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
