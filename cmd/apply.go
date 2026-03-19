package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

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

		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}

		// Back up existing config
		if _, err := os.Stat(destPath); err == nil {
			timestamp := time.Now().Format("20060102-150405")
			backup := destPath + "." + timestamp + ".bak"
			if err := copyFile(destPath, backup); err != nil {
				return fmt.Errorf("backing up existing config: %w", err)
			}
			out.Info("Backed up existing config to: " + backup)
		}

		// Merge: preserve non-mcpServers keys from existing dest config
		merged, err := mergeConfig(destPath, dest)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destPath, merged, 0o644); err != nil {
			return fmt.Errorf("writing config: %w", err)
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

// mergeConfig merges the generated mcpServers into the existing dest config,
// preserving all other top-level keys. If the dest file doesn't exist or is
// invalid JSON, returns the generated config as-is.
func mergeConfig(destPath string, dest *config.DestConfig) ([]byte, error) {
	existing, err := os.ReadFile(destPath)
	if err == nil {
		var raw map[string]json.RawMessage
		if json.Unmarshal(existing, &raw) == nil {
			servers, err := json.Marshal(dest.MCPServers)
			if err != nil {
				return nil, fmt.Errorf("marshalling servers: %w", err)
			}
			raw["mcpServers"] = servers
			data, err := json.MarshalIndent(raw, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("marshalling merged config: %w", err)
			}
			return append(data, '\n'), nil
		}
	}
	data, err := json.MarshalIndent(dest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling config: %w", err)
	}
	return append(data, '\n'), nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
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
