package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/damianoneill/claude-desktop-config/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for managing MCP server configurations",
	Long: `Launch an interactive terminal UI to browse, toggle, and apply
MCP server configurations without leaving the terminal.

  ↑↓ / jk   navigate servers
  space      toggle enabled/disabled (staged)
  s          save pending changes to source file
  a          save + apply to Claude Desktop config
  d          dry-run preview of enabled servers
  q          quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ac := appCtx(cmd)
		return runTUI(ac.SourceFile, ac.KeepBackups)
	},
}

func runTUI(sourceFile string, keepBackups int) error {
	m, err := tui.New(sourceFile, keepBackups)
	if err != nil {
		return fmt.Errorf("loading source file: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
