package cmd

import (
	"fmt"

	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Install the swyuse() shell helper into your rc file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sh, err := shell.Detect(shellOverride)
			if err != nil {
				exitError("%v", err)
			}
			already, err := shell.HasInitBlock(sh)
			if err != nil {
				exitError("could not check %s: %v", sh.RCFile, err)
			}
			if already {
				fmt.Printf("Shell helper already installed in %s. No changes made.\n", sh.RCFile)
				return nil
			}
			if err := shell.WriteInitBlock(sh); err != nil {
				exitError("could not update %s: %v", sh.RCFile, err)
			}
			fmt.Printf("Shell helper installed in %s.\n", sh.RCFile)
			fmt.Println("Installed function: swyuse()")
			fmt.Printf("Run: source %s\n", sh.RCFile)
			fmt.Println("Then use: swyuse <profile>")
			return nil
		},
	}
}
