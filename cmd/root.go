package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/config"
	"github.com/damianoneill/claude-desktop-config/internal/output"
)

// AppContext is threaded through all subcommands via cobra's context.
type AppContext struct {
	SourceFile  string
	Out         *output.Writer
	KeepBackups int
}

type ctxKey struct{}

func appCtx(cmd *cobra.Command) *AppContext {
	return cmd.Context().Value(ctxKey{}).(*AppContext)
}

var flagSource string
var flagKeepBackups int

var rootCmd = &cobra.Command{
	Use:   "claude-desktop-config",
	Short: "Manage Claude Desktop MCP server configurations",
	Long: `claude-desktop-config manages your Claude Desktop MCP server configuration.

Maintain a source file with per-server enable/disable flags and generate
the real Claude Desktop config from it.

Run without arguments to launch the interactive TUI.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		source := flagSource
		if source == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			source = filepath.Join(cwd, "claude_desktop_config.source.json")
		}

		out := output.New(false)

		ac := &AppContext{
			SourceFile:  source,
			Out:         out,
			KeepBackups: flagKeepBackups,
		}
		ctx := context.WithValue(cmd.Context(), ctxKey{}, ac)
		cmd.SetContext(ctx)
		return nil
	},
	// Launch the TUI when invoked with no subcommand.
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		return runTUI(ac.SourceFile, ac.KeepBackups)
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagSource, "source", "", "path to source JSON file (default: ./claude_desktop_config.source.json)")
	rootCmd.PersistentFlags().IntVar(&flagKeepBackups, "keep-backups", config.DefaultKeepBackups, "number of backup files to keep")
}
