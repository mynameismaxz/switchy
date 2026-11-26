package cmd

import (
	"fmt"
	"maps"
	"os"
	"sort"

	"github.com/mynameismaxz/switchy/internal/clipboard"
	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/mynameismaxz/switchy/internal/validate"
	"github.com/spf13/cobra"
)

// hookEnv marks that `swy use` was invoked by the shell hook wrapper. In this
// mode the hook itself applies the exports, so the command skips the clipboard
// flow and prints only a one-line status to stderr.
const hookEnv = "SWITCHY_HOOK"

func newUseCmd() *cobra.Command {
	var persistent bool
	var inline bool

	cmd := &cobra.Command{
		Use:               "use <profile>",
		Short:             "Select a profile (optionally persist it to the shell rc file)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: profileCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := validate.ProfileName(name); err != nil {
				exitError("%v", err)
			}
			vars, err := profile.GetProfile(name)
			if err != nil {
				exitError("%v", err)
			}

			if inline {
				sh, err := shell.Detect(shellOverride)
				if err != nil {
					exitError("%v", err)
				}
				upsertVars := make(map[string]string, len(vars)+1)
				maps.Copy(upsertVars, vars)
				upsertVars["SWITCHY_PROFILE"] = name
				if err := shell.UpsertExports(sh, upsertVars); err != nil {
					exitError("could not update %s: %v", sh.RCFile, err)
				}
				if err := selectProfile(name, "persistent"); err != nil {
					exitError("%v", err)
				}
				fmt.Printf("Profile %q applied to %s.\n", name, sh.RCFile)
				fmt.Printf("Run: source %s\n", sh.RCFile)
			} else if persistent {
				sh, err := shell.Detect(shellOverride)
				if err != nil {
					exitError("%v", err)
				}
				keys := make([]string, 0, len(vars))
				for k := range vars {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				exports := make([]string, 0, len(keys)+1)
				for _, k := range keys {
					exports = append(exports, shell.FormatExport(k, vars[k]))
				}
				exports = append(exports, shell.FormatExport("SWITCHY_PROFILE", name))
				if err := shell.WriteActivationBlock(sh, exports); err != nil {
					exitError("could not update %s: %v", sh.RCFile, err)
				}
				if err := selectProfile(name, "persistent"); err != nil {
					exitError("%v", err)
				}
				fmt.Printf("Profile %q activated in %s.\n", name, sh.RCFile)
				if hookMode() {
					fmt.Fprintf(os.Stderr, "Applied to this shell — no source needed.\n")
				} else {
					fmt.Printf("Run: source %s\n", sh.RCFile)
				}
			} else {
				if err := selectProfile(name, "session"); err != nil {
					exitError("%v", err)
				}
				if hookMode() {
					// The wrapper evals `swy hook-apply` next; stdout must stay
					// clean because the hook pipes only hook-apply into eval.
					fmt.Fprintf(os.Stderr, "⚡ applied %q to this shell\n", name)
					return nil
				}
				evalCmd := fmt.Sprintf(`eval "$(swy export %s)"`, name)
				fmt.Printf("Profile %q selected.\n", name)
				if clipboard.Copy(evalCmd) {
					fmt.Println("Copied to clipboard — paste and press Enter to apply.")
				} else {
					fmt.Printf("To apply now:  %s\n", evalCmd)
				}
				fmt.Printf("To persist:    swy use %s --persistent\n", name)
				hookTip()
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&persistent, "persistent", false, "Write profile to shell rc file for future sessions")
	cmd.Flags().BoolVar(&inline, "inline", false, "Edit export statements directly in the shell rc file (update existing or append new)")
	return cmd
}

func hookMode() bool { return os.Getenv(hookEnv) == "1" }

// selectProfile records the selection in state and bumps the generation so
// hooked shells pick the change up on their next prompt.
func selectProfile(name, mode string) error {
	st, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("could not load state: %w", err)
	}
	st.CurrentProfile = name
	st.LastActivationMode = mode
	if err := st.BumpGeneration(); err != nil {
		return fmt.Errorf("could not save state: %w", err)
	}
	return nil
}

// hookTip suggests the one-time hook installation when the session-mode
// clipboard flow is being used without the hook installed.
func hookTip() {
	sh, err := shell.Detect(shellOverride)
	if err != nil {
		return
	}
	installed, err := shell.HasInitBlock(sh)
	if err != nil || installed {
		return
	}
	fmt.Printf("Tip: run `swy init` once so `swy use` applies instantly — no clipboard needed.\n")
}
