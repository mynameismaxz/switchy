package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mynameismaxz/switchy/internal/tui"
)

var shellOverride string

func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "swy",
		Short: "Switchy — manage and switch environment variable profiles",
		Long: `Switchy (swy) lets you create named sets of environment variables
and switch between them without manually editing shell config files.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.Run(); err != nil {
				exitError("TUI error: %v", err)
			}
		},
	}

	root.PersistentFlags().StringVar(&shellOverride, "shell", "", "Override shell detection (zsh or bash)")

	root.AddCommand(
		newSetCmd(),
		newListCmd(),
		newShowCmd(),
		newExportCmd(),
		newUseCmd(),
		newCurrentCmd(),
		newDeleteCmd(),
		newInitCmd(),
		newVersionCmd(version),
	)

	return root
}

func exitError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
