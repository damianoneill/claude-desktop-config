package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

var disableCmd = &cobra.Command{
	Use:   "disable <server-name>",
	Short: "Disable a server by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		name := args[0]

		src, err := config.Load(ac.SourceFile)
		if err != nil {
			return err
		}

		srv, ok := src.MCPServers[name]
		if !ok {
			return fmt.Errorf("server not found: %s (run 'list' to see available servers)", name)
		}

		f := false
		srv.Enabled = &f
		src.MCPServers[name] = srv

		if err := config.Save(ac.SourceFile, src); err != nil {
			return err
		}

		ac.Out.Success(fmt.Sprintf("Disabled %s — run 'apply' to write the config", name))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}
