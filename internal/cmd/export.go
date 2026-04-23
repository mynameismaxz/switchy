package cmd

import (
	"fmt"
	"sort"

	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/shell"
	"github.com/mynameismaxz/switchy/internal/validate"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <profile>",
		Short: "Print shell export statements for a profile (for use with eval)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := validate.ProfileName(name); err != nil {
				exitError("%v", err)
			}
			vars, err := profile.GetProfile(name)
			if err != nil {
				exitError("%v", err)
			}
			keys := make([]string, 0, len(vars))
			for k := range vars {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Println(shell.FormatExport(k, vars[k]))
			}
			fmt.Println(shell.FormatExport("SWITCHY_PROFILE", name))
			return nil
		},
	}
}
