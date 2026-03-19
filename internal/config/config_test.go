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

func TestLoadSave_EnabledRoundtrip(t *testing.T) {
	// Verify _enabled is written and read back correctly.
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"on":  {Enabled: boolPtr(true), Command: "npx"},
			"off": {Enabled: boolPtr(false), Command: "npx"},
		},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "source.json")

	require.NoError(t, config.Save(path, src))

	// Verify the raw JSON uses _enabled, not enabled.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"_enabled"`)
	assert.NotContains(t, string(raw), `"enabled"`)

	loaded, err := config.Load(path)
	require.NoError(t, err)
	assert.True(t, *loaded.MCPServers["on"].Enabled)
	assert.False(t, *loaded.MCPServers["off"].Enabled)
}

func TestLoadSave_CommentRoundtrip(t *testing.T) {
	// Verify _comment survives a save/load cycle.
	src := &config.SourceConfig{
		MCPServers: map[string]config.MCPServer{
			"srv": {
				Enabled: boolPtr(true),
				Comment: "my description",
				Command: "npx",
			},
		},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "source.json")

	require.NoError(t, config.Save(path, src))

	loaded, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "my description", loaded.MCPServers["srv"].Comment)
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
			"c": {}, // nil _enabled = treated as enabled
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
	// Marshal to JSON and verify neither _enabled nor _comment appear in output.
	data, err := json.Marshal(dest)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "_enabled")
	assert.NotContains(t, string(data), "enabled")
	assert.NotContains(t, string(data), "_comment")
	assert.Contains(t, string(data), "npx")
}
