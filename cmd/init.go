package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create source file from example (skips if already exists)",
	// init does not require the source file to already exist — PersistentPreRunE only
	// resolves the path, it does not open or validate the file, so this is safe.
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		source := ac.SourceFile
		example := source + ".example"

		if _, err := os.Stat(source); err == nil {
			ac.Out.Info(fmt.Sprintf("%s already exists — edit it directly.", source))
			return nil
		}

		if _, err := os.Stat(example); os.IsNotExist(err) {
			return fmt.Errorf("example file not found: %s", example)
		}

		data, err := os.ReadFile(example)
		if err != nil {
			return fmt.Errorf("reading example file: %w", err)
		}
		if err := os.WriteFile(source, data, 0o644); err != nil {
			return fmt.Errorf("writing source file: %w", err)
		}

		ac.Out.Success(fmt.Sprintf("Created %s from example.", source))
		ac.Out.Info("Edit it with your credentials, then run: claude-desktop-config apply")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
