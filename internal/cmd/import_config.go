package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mynameismaxz/switchy/internal/config"
	"github.com/mynameismaxz/switchy/internal/profile"
	"github.com/mynameismaxz/switchy/internal/validate"
)

func newImportConfigCmd() *cobra.Command {
	var replace bool

	cmd := &cobra.Command{
		Use:   "import-config <input-file>",
		Short: "Import profiles from a JSON file",
		Long: `Import profiles from a JSON file created by 'swy export-config'.

By default, profiles are merged: existing profiles with the same name
are updated, and new profiles are added. Use --replace to remove all
existing profiles before importing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			data, err := os.ReadFile(path) // #nosec G304 -- path comes from CLI arg, user-intended file
			if err != nil {
				return fmt.Errorf("could not read %s: %w", path, err)
			}

			var pf config.ProfilesFile
			if err := json.Unmarshal(data, &pf); err != nil {
				return fmt.Errorf("invalid import file: %w", err)
			}

			if pf.Profiles == nil {
				return fmt.Errorf("invalid import file: no profiles found")
			}

			// Validate all profiles before importing anything
			for name, vars := range pf.Profiles {
				if err := validate.ProfileName(name); err != nil {
					return fmt.Errorf("invalid profile name %q: %w", name, err)
				}
				for k := range vars {
					if err := validate.EnvKey(k); err != nil {
						return fmt.Errorf("invalid env key %q in profile %q: %w", k, name, err)
					}
				}
			}

			if replace {
				existing, err := config.LoadProfiles()
				if err != nil {
					return fmt.Errorf("could not load existing profiles: %w", err)
				}
				// Delete all existing profiles first
				for name := range existing.Profiles {
					if err := profile.DeleteProfile(name, true); err != nil {
						return fmt.Errorf("could not delete profile %q: %w", name, err)
					}
				}
			}

			added := 0
			updated := 0
			for name, vars := range pf.Profiles {
				_, err := profile.GetProfile(name)
				exists := err == nil
				if err := profile.UpsertProfile(name, vars); err != nil {
					return fmt.Errorf("could not import profile %q: %w", name, err)
				}
				if exists {
					updated++
				} else {
					added++
				}
			}

			fmt.Printf("Imported %d profile(s) from %s (%d added, %d updated)\n",
				len(pf.Profiles), path, added, updated)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&replace, "replace", "r", false, "Replace all existing profiles before importing")
	return cmd
}
