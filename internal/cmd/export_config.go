package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mynameismaxz/switchy/internal/config"
)

func newExportConfigCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export-config [output-file]",
		Short: "Export all profiles to a JSON file",
		Long: `Export all profiles to a portable JSON file for backup or transfer.

If no output file is specified, the JSON is printed to stdout.
The exported format is compatible with 'swy import-config'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := outputPath
			if len(args) == 1 {
				path = args[0]
			}

			pf, err := config.LoadProfiles()
			if err != nil {
				return fmt.Errorf("could not load profiles: %w", err)
			}

			data, err := json.MarshalIndent(pf, "", "  ")
			if err != nil {
				return fmt.Errorf("could not marshal profiles: %w", err)
			}

			if path == "" {
				fmt.Print(string(data))
				return nil
			}

			if err := os.WriteFile(path, data, 0600); err != nil {
				return fmt.Errorf("could not write %s: %w", path, err)
			}
			fmt.Printf("Exported %d profile(s) to %s\n", len(pf.Profiles), path)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")
	return cmd
}
