package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/damianoneill/claude-desktop-config/internal/config"
)

func TestDestPath_ReturnsNonEmptyPath(t *testing.T) {
	path, err := config.DestPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestDestPath_EndsWithConfigFilename(t *testing.T) {
	path, err := config.DestPath()
	require.NoError(t, err)
	assert.Equal(t, "claude_desktop_config.json", filepath.Base(path))
}

func TestDestPath_ContainsClaudeDirectory(t *testing.T) {
	path, err := config.DestPath()
	require.NoError(t, err)
	assert.Contains(t, path, "Claude")
}

func TestDestPath_IsAbsolute(t *testing.T) {
	path, err := config.DestPath()
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(path), "DestPath should return an absolute path, got: %s", path)
}

func TestDestPath_DarwinPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	path, err := config.DestPath()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	assert.Equal(t, expected, path)
}

func TestDestPath_LinuxPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-specific test")
	}
	path, err := config.DestPath()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	assert.Equal(t, expected, path)
}

func TestDestPath_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific test")
	}
	path, err := config.DestPath()
	require.NoError(t, err)

	appData := os.Getenv("APPDATA")
	require.NotEmpty(t, appData, "APPDATA must be set on Windows")

	expected := filepath.Join(appData, "Claude", "claude_desktop_config.json")
	assert.Equal(t, expected, path)
}

func TestDestPath_WindowsMissingAPPDATA(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific test")
	}

	original := os.Getenv("APPDATA")
	t.Cleanup(func() { require.NoError(t, os.Setenv("APPDATA", original)) })
	require.NoError(t, os.Unsetenv("APPDATA"))

	_, err := config.DestPath()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "APPDATA")
}

func TestDestPath_UnderHomeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("home-dir check not applicable on Windows (uses APPDATA)")
	}

	path, err := config.DestPath()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(path, home),
		"DestPath %q should be under home directory %q", path, home)
}
