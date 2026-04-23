package cmd

import (
	"fmt"

	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/spf13/cobra"
)

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the last profile selected through Switchy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := config.LoadState()
			if err != nil {
				exitError("could not read state: %v", err)
			}
			if st.CurrentProfile == "" {
				fmt.Println("No profile currently selected.")
				fmt.Println("Use: swy use <profile>")
				return nil
			}
			mode := st.LastActivationMode
			if mode == "" {
				mode = "none-applied"
			}
			fmt.Printf("Current profile:  %s\n", st.CurrentProfile)
			fmt.Printf("Activation mode:  %s\n", mode)
			if mode != "persistent" {
				fmt.Printf("To use now:       eval \"$(swy export %s)\"\n", st.CurrentProfile)
				fmt.Printf("To persist:       swy use %s --persistent\n", st.CurrentProfile)
			}
			return nil
		},
	}
}
