package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// SourceConfig is the top-level source JSON structure.
type SourceConfig struct {
	Comment    string               `json:"_comment,omitempty"`
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer is a single server entry in the source file.
type MCPServer struct {
	Enabled *bool             `json:"enabled,omitempty"`
	Comment string            `json:"_comment,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// CleanServer is what goes into the real Claude Desktop config.
type CleanServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// DestConfig is the structure written to the Claude Desktop config.
type DestConfig struct {
	MCPServers map[string]CleanServer `json:"mcpServers"`
}

// IsEnabled reports whether a server should be included in output.
// A nil enabled pointer (field absent) is treated as enabled.
func IsEnabled(s MCPServer) bool {
	return s.Enabled == nil || *s.Enabled
}

// Load reads and unmarshals the source JSON file.
func Load(path string) (*SourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading source file: %w", err)
	}
	var cfg SourceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing source file: %w", err)
	}
	return &cfg, nil
}

// Save marshals cfg and writes it to path with 2-space indentation.
func Save(path string, cfg *SourceConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing source file: %w", err)
	}
	return nil
}

// Filter returns a DestConfig containing only the enabled servers,
// with the enabled and _comment fields stripped.
func Filter(cfg *SourceConfig) *DestConfig {
	dest := &DestConfig{
		MCPServers: make(map[string]CleanServer, len(cfg.MCPServers)),
	}
	for name, srv := range cfg.MCPServers {
		if IsEnabled(srv) {
			dest.MCPServers[name] = CleanServer{
				Command: srv.Command,
				Args:    srv.Args,
				Env:     srv.Env,
			}
		}
	}
	return dest
}

// Counts returns the total and enabled server counts from a SourceConfig.
func Counts(cfg *SourceConfig) (total, enabled int) {
	total = len(cfg.MCPServers)
	for _, srv := range cfg.MCPServers {
		if IsEnabled(srv) {
			enabled++
		}
	}
	return total, enabled
}
