package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/damianoneill/claude-desktop-config/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestIsEnabled(t *testing.T) {
	assert.True(t, config.IsEnabled(config.MCPServer{Enabled: nil}))
	assert.True(t, config.IsEnabled(config.MCPServer{Enabled: boolPtr(true)}))
	assert.False(t, config.IsEnabled(config.MCPServer{Enabled: boolPtr(false)}))
}

func TestFilter(t *testing.T) {
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"on":  {Enabled: boolPtr(true), Command: "npx", Args: []string{"a", "http://on"}},
			"off": {Enabled: boolPtr(false), Command: "npx", Args: []string{"a", "http://off"}},
			"nil": {Command: "npx", Args: []string{"a", "http://nil"}},
		},
	}
	dest := config.Filter(src)
	assert.Len(t, dest.MCPServers, 2)
	assert.Contains(t, dest.MCPServers, "on")
	assert.Contains(t, dest.MCPServers, "nil")
	assert.NotContains(t, dest.MCPServers, "off")
}

func TestLoadSave(t *testing.T) {
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"test": {Enabled: boolPtr(true), Command: "npx"},
		},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "source.json")

	require.NoError(t, config.Save(path, src))

	loaded, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, src.MCPServers["test"].Command, loaded.MCPServers["test"].Command)
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o644))
	_, err := config.Load(path)
	assert.Error(t, err)
}

func TestCounts(t *testing.T) {
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"a": {Enabled: boolPtr(true)},
			"b": {Enabled: boolPtr(false)},
			"c": {}, // nil = enabled
		},
	}
	total, enabled := config.Counts(src)
	assert.Equal(t, 3, total)
	assert.Equal(t, 2, enabled)
}

func TestFilterStripsMetaFields(t *testing.T) {
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"srv": {
				Enabled: boolPtr(true),
				Comment: "should be stripped",
				Command: "npx",
				Args:    []string{"mcp-remote", "http://example.com"},
			},
		},
	}
	dest := config.Filter(src)
	// Marshal to JSON and verify no "enabled" or "_comment" keys
	data, err := json.Marshal(dest)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "enabled")
	assert.NotContains(t, string(data), "_comment")
	assert.Contains(t, string(data), "npx")
}
