package cmd

import (
	"fmt"

	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/validate"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <profile> KEY=VALUE [KEY=VALUE...]",
		Short: "Create or update a profile with the given environment variables",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := validate.ProfileName(name); err != nil {
				exitError("%v", err)
			}
			kvPairs, err := validate.ParseKVPairs(args[1:])
			if err != nil {
				exitError("%v", err)
			}
			if err := profile.UpsertProfile(name, kvPairs); err != nil {
				exitError("could not save profile: %v", err)
			}
			fmt.Printf("Profile %q updated.\n", name)
			return nil
		},
	}
}
