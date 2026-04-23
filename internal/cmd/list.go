package cmd

import (
	"fmt"

	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			names, current, err := profile.ListProfiles()
			if err != nil {
				exitError("could not list profiles: %v", err)
			}
			if len(names) == 0 {
				fmt.Println("No profiles found.")
				fmt.Println("Create one with:")
				fmt.Println("  swy set local ANTHROPIC_BASE_URL=http://localhost:8080")
				return nil
			}
			for _, name := range names {
				if name == current {
					fmt.Printf("* %s\n", name)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}
			return nil
		},
	}
}
