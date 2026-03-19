package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers and their enabled status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		out := ac.Out

		src, err := config.Load(ac.SourceFile)
		if err != nil {
			return err
		}

		names := sortedKeys(src.MCPServers)
		for _, name := range names {
			srv := src.MCPServers[name]
			status := "[off]"
			if config.IsEnabled(srv) {
				status = "[on] "
			}
			url := ""
			if len(srv.Args) > 1 {
				url = srv.Args[1]
			}
			out.Info(fmt.Sprintf("  %s %-45s %s", status, name, url))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
