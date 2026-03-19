package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Preview generated config without writing to disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		out := ac.Out

		src, err := config.Load(ac.SourceFile)
		if err != nil {
			return err
		}

		dest := config.Filter(src)
		total, enabled := config.Counts(src)
		disabled := total - enabled

		destPath, err := config.DestPath()
		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(dest, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling config: %w", err)
		}

		out.Info("=== Dry run — would write the following config ===")
		out.Info("")
		out.Info(string(data))
		out.Info("")
		out.Info("=== Summary ===")
		out.Info(fmt.Sprintf("  Enabled : %d server(s)", enabled))
		out.Info(fmt.Sprintf("  Disabled: %d server(s)", disabled))
		out.Info(fmt.Sprintf("  Source  : %s", ac.SourceFile))
		out.Info(fmt.Sprintf("  Output  : %s (dry-run, nothing written)", destPath))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dryRunCmd)
}
