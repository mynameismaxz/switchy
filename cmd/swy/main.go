package main

import (
	"fmt"
	"os"

	"github.com/mynameismaxz/switchy/internal/cmd"
)

// Injected via -ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	versionStr := fmt.Sprintf("swy %s (commit %s, built %s)", version, commit, date)
	root := cmd.NewRootCmd(versionStr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
