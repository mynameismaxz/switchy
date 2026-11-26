package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mynameismaxz/switchy/internal/activation"
	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook <zsh|bash>",
		Short: "Print the shell hook script for manual installation",
		Long: `Print the shell hook script for manual installation.

The hook makes "swy use" and "swy off" apply to the current shell immediately,
no clipboard or pasting needed. "swy init" installs it for you; use this
command only to install it by hand, for example:

  swy hook zsh >> ~/.zshrc
  swy hook bash >> ~/.bashrc`,
		ValidArgs: []string{"zsh", "bash"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			var t shell.ShellType
			switch args[0] {
			case "zsh":
				t = shell.ShellZsh
			case "bash":
				t = shell.ShellBash
			}
			fmt.Println(shell.HookScript(t))
			return nil
		},
	}
}

func newHookApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook-apply",
		Short:  "Print exports that bring the current shell in sync with switchy state (used by the shell hook)",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			sync := activation.Sync{
				AppliedGeneration: parseGenEnv(),
				LoadedProfile:     os.Getenv(activation.ProfileEnv),
			}

			// A broken config must never break the user's prompt: fail silently.
			st, err := config.LoadState()
			if err == nil {
				sync.Generation = st.Generation
				sync.CurrentProfile = st.CurrentProfile
			}
			if name := sync.CurrentProfile; name != "" {
				if vars, err := profile.GetProfile(name); err == nil {
					sync.CurrentVars = vars
				}
			}
			if name := sync.LoadedProfile; name != "" {
				if vars, err := profile.GetProfile(name); err == nil {
					sync.LoadedVars = vars
				}
			}

			for _, line := range activation.ApplyScript(sync) {
				fmt.Println(line)
			}
			return nil
		},
	}
}

func newOffCmd() *cobra.Command {
	var persistent bool

	cmd := &cobra.Command{
		Use:   "off",
		Short: "Remove the active profile from this shell and clear the selection",
		Long: `Remove the active profile from this shell and clear the selection.

Eval the printed script to unload the profile's variables from the current
shell — with the shell hook installed, plain "swy off" does this for you.
Shells with the hook unload the profile on their next prompt too.

With --persistent, also remove the managed activation block from the rc file.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded := os.Getenv(activation.ProfileEnv)
			var loadedVars map[string]string
			if loaded != "" {
				if vars, err := profile.GetProfile(loaded); err == nil {
					loadedVars = vars
				}
			}

			// Clear the selection so hooked shells unload as well.
			// A broken config must not block unloading.
			if st, err := config.LoadState(); err == nil && st.CurrentProfile != "" {
				st.CurrentProfile = ""
				st.LastActivationMode = ""
				if err := st.BumpGeneration(); err != nil {
					exitError("could not save state: %v", err)
				}
			}

			for _, line := range activation.OffScript(activation.Off{
				LoadedProfile: loaded,
				LoadedVars:    loadedVars,
			}) {
				fmt.Println(line)
			}

			if persistent {
				sh, err := shell.Detect(shellOverride)
				if err != nil {
					exitError("%v", err)
				}
				if err := shell.BackupIfNeeded(sh); err != nil {
					exitError("could not back up %s: %v", sh.RCFile, err)
				}
				if err := shell.RemoveActivationBlock(sh); err != nil {
					exitError("could not update %s: %v", sh.RCFile, err)
				}
				fmt.Fprintf(os.Stderr, "Activation block removed from %s.\n", sh.RCFile)
			} else {
				warnPersistentBlock()
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&persistent, "persistent", false, "Also remove the managed activation block from the shell rc file")
	return cmd
}

// warnPersistentBlock prints a hint when a persistent activation block would
// still load the profile into new shells. Printed to stderr so it never
// pollutes the eval'd script.
func warnPersistentBlock() {
	sh, err := shell.Detect(shellOverride)
	if err != nil {
		return
	}
	has, err := shell.HasActivationBlock(sh)
	if err != nil || !has {
		return
	}
	fmt.Fprintf(os.Stderr, "Note: a persistent block remains in %s; new shells will still load it.\nRemove it with: swy off --persistent\n", sh.RCFile)
}

// parseGenEnv reads the generation marker set by the shell hook.
func parseGenEnv() int {
	n, err := strconv.Atoi(os.Getenv(activation.AppliedGenEnv))
	if err != nil {
		return 0
	}
	return n
}
