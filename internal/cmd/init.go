package cmd

import (
	"fmt"

	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Install the swy shell hook into your rc file",
		Long: `Install the swy shell hook into your rc file (idempotent — safe to run multiple times).

After sourcing your rc file once, "swy use <profile>" and "swy off" apply to
the current shell immediately — no clipboard, no pasting, no eval — and other
hooked shells follow along on their next prompt.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sh, err := shell.Detect(shellOverride)
			if err != nil {
				exitError("%v", err)
			}
			upToDate, err := shell.InitBlockUpToDate(sh)
			if err != nil {
				exitError("could not check %s: %v", sh.RCFile, err)
			}
			if upToDate {
				fmt.Printf("Shell hook already installed in %s. No changes made.\n", sh.RCFile)
				return nil
			}
			already, err := shell.HasInitBlock(sh)
			if err != nil {
				exitError("could not check %s: %v", sh.RCFile, err)
			}
			if err := shell.WriteInitBlock(sh); err != nil {
				exitError("could not update %s: %v", sh.RCFile, err)
			}
			if already {
				fmt.Printf("Shell hook upgraded in %s (older helper replaced).\n", sh.RCFile)
			} else {
				fmt.Printf("Shell hook installed in %s.\n", sh.RCFile)
			}
			fmt.Println("Installed: swy() wrapper, swyuse(), and a prompt hook")
			fmt.Printf("Run: source %s   # once — after this, switchy applies instantly\n", sh.RCFile)
			fmt.Println("Then: swy use <profile>   # applies immediately, no clipboard")
			return nil
		},
	}
}
