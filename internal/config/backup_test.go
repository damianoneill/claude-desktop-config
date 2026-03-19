package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/damianoneill/claude-desktop-config/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeBackup creates a fake backup file with the standard naming convention.
func makeBackup(t *testing.T, destPath, timestamp string) string {
	t.Helper()
	path := destPath + "." + timestamp + ".bak"
	require.NoError(t, os.WriteFile(path, []byte("backup "+timestamp), 0o644))
	return path
}

// backupCount returns the number of .bak files alongside destPath.
func backupCount(t *testing.T, destPath string) int {
	t.Helper()
	dir := filepath.Dir(destPath)
	base := filepath.Base(destPath)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	n := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), base+".") && strings.HasSuffix(e.Name(), ".bak") {
			n++
		}
	}
	return n
}

// ── PruneBackups ──────────────────────────────────────────────────────────────

func TestPruneBackups_RemovesOldest(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Create 5 backups with ascending timestamps so lexical sort = time sort.
	for i := range 5 {
		ts := time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC).Format("20060102-150405")
		makeBackup(t, dest, ts)
	}
	require.Equal(t, 5, backupCount(t, dest))

	require.NoError(t, config.PruneBackups(dest, 3))

	assert.Equal(t, 3, backupCount(t, dest))
}

func TestPruneBackups_KeepsNewest(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	timestamps := []string{
		"20250101-000000",
		"20250102-000000",
		"20250103-000000",
		"20250104-000000",
		"20250105-000000",
	}
	for _, ts := range timestamps {
		makeBackup(t, dest, ts)
	}

	require.NoError(t, config.PruneBackups(dest, 3))

	// The three newest should survive.
	for _, ts := range timestamps[2:] {
		path := dest + "." + ts + ".bak"
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected %s to exist", filepath.Base(path))
	}
	// The two oldest should be gone.
	for _, ts := range timestamps[:2] {
		path := dest + "." + ts + ".bak"
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), "expected %s to be deleted", filepath.Base(path))
	}
}

func TestPruneBackups_NoDeletionWhenUnderLimit(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	makeBackup(t, dest, "20250101-000000")
	makeBackup(t, dest, "20250102-000000")

	require.NoError(t, config.PruneBackups(dest, 3))

	assert.Equal(t, 2, backupCount(t, dest))
}

func TestPruneBackups_ExactlyAtLimit(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	for i := range 3 {
		ts := time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC).Format("20060102-150405")
		makeBackup(t, dest, ts)
	}

	require.NoError(t, config.PruneBackups(dest, 3))

	assert.Equal(t, 3, backupCount(t, dest))
}

func TestPruneBackups_NoBackupsExist(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Should not error when there are no backups at all.
	require.NoError(t, config.PruneBackups(dest, 3))
}

func TestPruneBackups_ZeroKeepUsesDefault(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Create more backups than the default (3).
	for i := range 6 {
		ts := time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC).Format("20060102-150405")
		makeBackup(t, dest, ts)
	}

	require.NoError(t, config.PruneBackups(dest, 0))

	assert.Equal(t, config.DefaultKeepBackups, backupCount(t, dest))
}

func TestPruneBackups_IgnoresUnrelatedFiles(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Unrelated files that should never be touched.
	unrelated := []string{
		"other_file.bak",
		"claude_desktop_config.json.notabak",
		"claude_desktop_config.json",
	}
	for _, name := range unrelated {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644))
	}

	makeBackup(t, dest, "20250101-000000")
	makeBackup(t, dest, "20250102-000000")
	makeBackup(t, dest, "20250103-000000")
	makeBackup(t, dest, "20250104-000000")

	require.NoError(t, config.PruneBackups(dest, 3))

	// Unrelated files must still exist.
	for _, name := range unrelated {
		_, err := os.Stat(filepath.Join(dir, name))
		assert.NoError(t, err, "unrelated file %s should not have been deleted", name)
	}

	assert.Equal(t, 3, backupCount(t, dest))
}

// ── WriteConfig ───────────────────────────────────────────────────────────────

func TestWriteConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")
	data := []byte(`{"mcpServers":{}}` + "\n")

	require.NoError(t, config.WriteConfig(dest, data, 3))

	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestWriteConfig_CreatesDestDir(t *testing.T) {
	dir := t.TempDir()
	// Nested directory that doesn't exist yet.
	dest := filepath.Join(dir, "Claude", "claude_desktop_config.json")

	require.NoError(t, config.WriteConfig(dest, []byte("{}"), 3))

	_, err := os.Stat(dest)
	assert.NoError(t, err)
}

func TestWriteConfig_BacksUpExistingFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	original := []byte(`{"mcpServers":{"old":{}}}`)
	require.NoError(t, os.WriteFile(dest, original, 0o644))

	require.NoError(t, config.WriteConfig(dest, []byte(`{"mcpServers":{}}`), 3))

	// Exactly one backup should have been created.
	assert.Equal(t, 1, backupCount(t, dest))
}

func TestWriteConfig_BackupContainsOriginalContent(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	original := []byte(`{"mcpServers":{"preserved":{}}}`)
	require.NoError(t, os.WriteFile(dest, original, 0o644))

	require.NoError(t, config.WriteConfig(dest, []byte(`{"mcpServers":{}}`), 3))

	// Find the backup and verify its content matches the original.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var backupPath string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bak") {
			backupPath = filepath.Join(dir, e.Name())
		}
	}
	require.NotEmpty(t, backupPath, "expected a backup file")

	got, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestWriteConfig_PrunesOldBackups(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Pre-create 3 existing backups (at the keep limit).
	for i := range 3 {
		ts := time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC).Format("20060102-150405")
		makeBackup(t, dest, ts)
	}

	// Write the live file so WriteConfig can back it up (adding a 4th backup).
	require.NoError(t, os.WriteFile(dest, []byte(`{"mcpServers":{}}`), 0o644))
	require.NoError(t, config.WriteConfig(dest, []byte(`{"mcpServers":{"new":{}}}`), 3))

	// After pruning, still only 3 backups.
	assert.Equal(t, 3, backupCount(t, dest))
}

func TestWriteConfig_NoBackupWhenDestAbsent(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "claude_desktop_config.json")

	// Dest does not exist yet — no backup should be created.
	require.NoError(t, config.WriteConfig(dest, []byte(`{}`), 3))

	assert.Equal(t, 0, backupCount(t, dest))
}
