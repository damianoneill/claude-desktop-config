package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/output"
)

// AppContext is threaded through all subcommands via cobra's context.
type AppContext struct {
	SourceFile string
	Out        *output.Writer
}

type ctxKey struct{}

func appCtx(cmd *cobra.Command) *AppContext {
	return cmd.Context().Value(ctxKey{}).(*AppContext)
}

var flagSource string

var rootCmd = &cobra.Command{
	Use:   "claude-desktop-config",
	Short: "Manage Claude Desktop MCP server configurations",
	Long: `claude-desktop-config manages your Claude Desktop MCP server configuration.

Maintain a source file with per-server enable/disable flags and generate
the real Claude Desktop config from it.`,
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
			SourceFile: source,
			Out:        out,
		}
		ctx := context.WithValue(cmd.Context(), ctxKey{}, ac)
		cmd.SetContext(ctx)
		return nil
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
}
