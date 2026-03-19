package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DefaultKeepBackups = 3

// WriteConfig backs up the existing dest file (if present), writes data to
// destPath, then prunes old backups so at most keepBackups are retained.
func WriteConfig(destPath string, data []byte, keepBackups int) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if _, err := os.Stat(destPath); err == nil {
		ts := time.Now().Format("20060102-150405")
		backup := destPath + "." + ts + ".bak"
		if err := copyFile(destPath, backup); err != nil {
			return fmt.Errorf("backing up existing config: %w", err)
		}
	}

	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	if err := PruneBackups(destPath, keepBackups); err != nil {
		// Non-fatal — pruning failure should not fail the overall apply.
		_, _ = fmt.Fprintf(os.Stderr, "warn: pruning backups: %v\n", err)
	}

	return nil
}

// PruneBackups removes old backup files for destPath, keeping the most recent
// keepN. Backup files are matched by the pattern "<destPath>.<timestamp>.bak".
func PruneBackups(destPath string, keepN int) error {
	if keepN <= 0 {
		keepN = DefaultKeepBackups
	}

	dir := filepath.Dir(destPath)
	base := filepath.Base(destPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	// Collect backup filenames that match <base>.<timestamp>.bak
	var backups []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, base+".") && strings.HasSuffix(name, ".bak") {
			backups = append(backups, filepath.Join(dir, name))
		}
	}

	// Sort ascending by name — timestamp format YYYYMMDD-HHMMSS sorts lexically.
	sort.Strings(backups)

	// Remove oldest entries beyond keepN.
	if len(backups) > keepN {
		for _, old := range backups[:len(backups)-keepN] {
			if err := os.Remove(old); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("removing %s: %w", old, err)
			}
		}
	}

	return nil
}

// MergeConfig merges the generated mcpServers into the existing dest config,
// preserving all other top-level keys. If the dest file doesn't exist or is
// invalid JSON, returns the generated config as-is.
func MergeConfig(destPath string, dest *DestConfig) ([]byte, error) {
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
