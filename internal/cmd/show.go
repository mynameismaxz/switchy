package cmd

import (
	"fmt"
	"sort"

	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/validate"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <profile>",
		Short: "Display environment variables for a profile",
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

			fmt.Printf("Profile: %s\n", name)
			for _, k := range keys {
				v := vars[k]
				if validate.IsSensitiveKey(k) {
					v = validate.MaskValue(v)
				}
				fmt.Printf("  %-30s = %s\n", k, v)
			}
			return nil
		},
	}
}
