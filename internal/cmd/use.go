package cmd

import (
	"fmt"
	"maps"
	"sort"

	"github.com/mynameismaxz/switchy/internal/clipboard"
	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/mynameismaxz/switchy/internal/validate"
	"github.com/spf13/cobra"
)

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
				if err := config.SaveState(&config.StateFile{
					CurrentProfile:     name,
					LastActivationMode: "persistent",
				}); err != nil {
					exitError("could not save state: %v", err)
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
				if err := config.SaveState(&config.StateFile{
					CurrentProfile:     name,
					LastActivationMode: "persistent",
				}); err != nil {
					exitError("could not save state: %v", err)
				}
				fmt.Printf("Profile %q activated in %s.\n", name, sh.RCFile)
				fmt.Printf("Run: source %s\n", sh.RCFile)
			} else {
				if err := config.SaveState(&config.StateFile{
					CurrentProfile:     name,
					LastActivationMode: "session",
				}); err != nil {
					exitError("could not save state: %v", err)
				}
				evalCmd := fmt.Sprintf(`eval "$(swy export %s)"`, name)
				fmt.Printf("Profile %q selected.\n", name)
				if clipboard.Copy(evalCmd) {
					fmt.Println("Copied to clipboard — paste and press Enter to apply.")
				} else {
					fmt.Printf("To apply now:  %s\n", evalCmd)
				}
				fmt.Printf("To persist:    swy use %s --persistent\n", name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&persistent, "persistent", false, "Write profile to shell rc file for future sessions")
	cmd.Flags().BoolVar(&inline, "inline", false, "Edit export statements directly in the shell rc file (update existing or append new)")
	return cmd
}
