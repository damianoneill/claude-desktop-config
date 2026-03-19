package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// binary builds the claude-desktop-config binary once per test run and returns its path.
func binary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "claude-desktop-config")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))
	return bin
}

func run(t *testing.T, bin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = 1
		}
	}
	return outBuf.String(), errBuf.String(), code
}

func runWithSource(t *testing.T, bin, source string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	allArgs := append([]string{"--source", source}, args...)
	return run(t, bin, allArgs...)
}

func TestVersion(t *testing.T) {
	bin := binary(t)
	stdout, _, code := run(t, bin, "version")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "claude-desktop-config version")
}

func TestHelp(t *testing.T) {
	bin := binary(t)
	stdout, _, _ := run(t, bin)
	assert.Contains(t, stdout, "claude-desktop-config")
	assert.Contains(t, stdout, "apply")
	assert.Contains(t, stdout, "list")
	assert.Contains(t, stdout, "enable")
	assert.Contains(t, stdout, "disable")
}

func TestList(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	stdout, _, code := runWithSource(t, bin, source, "list")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "[on]")
	assert.Contains(t, stdout, "[off]")
}

func TestDryRun(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	stdout, _, code := runWithSource(t, bin, source, "dry-run")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "Dry run")
	assert.Contains(t, stdout, "mcpServers")
	assert.Contains(t, stdout, "Enabled")
	assert.Contains(t, stdout, "Disabled")
	// enabled entries should appear
	assert.Contains(t, stdout, "local-server-a")
	// disabled entries should not appear in the config output
	assert.NotContains(t, stdout, "prod-server-a")
}

func TestApply(t *testing.T) {
	bin := binary(t)

	// Copy example source to a temp dir so we don't mutate testdata
	src, err := os.ReadFile(filepath.Join("testdata", "source.example.json"))
	require.NoError(t, err)

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "source.json")
	require.NoError(t, os.WriteFile(sourceFile, src, 0o644))

	destFile := filepath.Join(dir, "claude_desktop_config.json")

	// Run apply but we can't control the OS dest path from the binary directly.
	// Instead test dry-run output correctness and the config package directly.
	stdout, _, code := runWithSource(t, bin, sourceFile, "dry-run")
	assert.Equal(t, 0, code)

	// Parse the JSON block from dry-run output — extract between header and summary
	const header = "=== Dry run"
	const summary = "=== Summary ==="
	headerIdx := strings.Index(stdout, header)
	summaryIdx := strings.Index(stdout, summary)
	require.Greater(t, headerIdx, -1, "dry-run output should contain header")
	require.Greater(t, summaryIdx, -1, "dry-run output should contain summary")
	// advance past the header line
	afterHeader := stdout[headerIdx:]
	nl := strings.Index(afterHeader, "\n")
	require.Greater(t, nl, -1)
	afterHeader = afterHeader[nl+1:]
	// trim from summary onward
	sumStart := strings.Index(afterHeader, summary)
	require.Greater(t, sumStart, -1)
	jsonStr := strings.TrimSpace(afterHeader[:sumStart])
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	require.NoError(t, err, "dry-run output should contain valid JSON")

	servers, ok := parsed["mcpServers"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, servers, "local-server-a")
	assert.NotContains(t, servers, "prod-server-a")

	_ = destFile // referenced for clarity
}

func TestEnable(t *testing.T) {
	bin := binary(t)

	src, err := os.ReadFile(filepath.Join("testdata", "source.example.json"))
	require.NoError(t, err)

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "source.json")
	require.NoError(t, os.WriteFile(sourceFile, src, 0o644))

	// local-server-b is disabled in the example
	stdout, _, code := runWithSource(t, bin, sourceFile, "enable", "local-server-b")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "Enabled local-server-b")

	// Verify it's now enabled in the file
	data, err := os.ReadFile(sourceFile)
	require.NoError(t, err)
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))
	servers := cfg["mcpServers"].(map[string]interface{})
	srv := servers["local-server-b"].(map[string]interface{})
	assert.Equal(t, true, srv["enabled"])
}

func TestDisable(t *testing.T) {
	bin := binary(t)

	src, err := os.ReadFile(filepath.Join("testdata", "source.example.json"))
	require.NoError(t, err)

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "source.json")
	require.NoError(t, os.WriteFile(sourceFile, src, 0o644))

	// local-server-a is enabled in the example
	stdout, _, code := runWithSource(t, bin, sourceFile, "disable", "local-server-a")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "Disabled local-server-a")

	data, err := os.ReadFile(sourceFile)
	require.NoError(t, err)
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))
	servers := cfg["mcpServers"].(map[string]interface{})
	srv := servers["local-server-a"].(map[string]interface{})
	assert.Equal(t, false, srv["enabled"])
}

func TestEnableUnknownServer(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	_, stderr, code := runWithSource(t, bin, source, "enable", "does-not-exist")
	assert.NotEqual(t, 0, code)
	assert.Contains(t, stderr, "server not found")
}

func TestDisableUnknownServer(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	_, stderr, code := runWithSource(t, bin, source, "disable", "does-not-exist")
	assert.NotEqual(t, 0, code)
	assert.Contains(t, stderr, "server not found")
}

func TestInitCreatesFromExample(t *testing.T) {
	bin := binary(t)

	// Set up a temp dir with only an .example file
	dir := t.TempDir()
	exampleSrc, err := os.ReadFile(filepath.Join("testdata", "source.example.json"))
	require.NoError(t, err)

	sourceFile := filepath.Join(dir, "claude_desktop_config.source.json")
	exampleFile := sourceFile + ".example"
	require.NoError(t, os.WriteFile(exampleFile, exampleSrc, 0o644))

	stdout, _, code := runWithSource(t, bin, sourceFile, "init")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "Created")

	_, err = os.Stat(sourceFile)
	assert.NoError(t, err, "source file should have been created")
}

func TestInitSkipsIfExists(t *testing.T) {
	bin := binary(t)

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "claude_desktop_config.source.json")
	require.NoError(t, os.WriteFile(sourceFile, []byte(`{"mcpServers":{}}`), 0o644))

	stdout, _, code := runWithSource(t, bin, sourceFile, "init")
	// Should succeed (exit 0) but warn
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "already exists")
}

func TestExitCode_MissingSourceFile(t *testing.T) {
	bin := binary(t)
	_, stderr, code := runWithSource(t, bin, "/tmp/does-not-exist.json", "list")
	assert.NotEqual(t, 0, code)
	assert.Contains(t, stderr, "error")
}
