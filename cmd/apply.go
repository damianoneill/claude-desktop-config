package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Generate and write Claude Desktop config from source file",
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

		merged, err := config.MergeConfig(destPath, dest)
		if err != nil {
			return err
		}

		if err := config.WriteConfig(destPath, merged, ac.KeepBackups); err != nil {
			return err
		}

		out.Info("")
		out.Success("Claude Desktop config generated successfully.")
		out.Info("")
		out.Info(fmt.Sprintf("  Source  : %s", ac.SourceFile))
		out.Info(fmt.Sprintf("  Output  : %s", destPath))
		out.Info(fmt.Sprintf("  Enabled : %d server(s)", enabled))
		out.Info(fmt.Sprintf("  Disabled: %d server(s)", disabled))
		out.Info("")
		out.Info("Active MCP servers:")

		names := sortedKeys(dest.MCPServers)
		for _, name := range names {
			srv := dest.MCPServers[name]
			url := "(stdio)"
			if len(srv.Args) > 1 {
				url = srv.Args[1]
			}
			out.Info(fmt.Sprintf("  - %s  →  %s", name, url))
		}

		out.Info("")
		out.Info("Restart Claude Desktop to pick up the new configuration.")
		return nil
	},
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
