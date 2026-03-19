package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DestPath returns the OS-appropriate path to the Claude Desktop config file.
//
//   - darwin  → ~/Library/Application Support/Claude/claude_desktop_config.json
//   - linux   → ~/.config/Claude/claude_desktop_config.json
//   - windows → %APPDATA%/Claude/claude_desktop_config.json
func DestPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil

	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil

	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% environment variable is not set")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil

	default:
		// Fallback to XDG-style for unknown OSes.
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
}
