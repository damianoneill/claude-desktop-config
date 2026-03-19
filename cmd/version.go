package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

var versionCmd = &cobra.Command{
	Use:               "version",
	Short:             "Print the claude-desktop-config version",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("claude-desktop-config version", resolveVersion())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
