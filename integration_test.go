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
	allArgs := make([]string, 0, 2+len(args))
	allArgs = append(allArgs, "--source", source)
	allArgs = append(allArgs, args...)
	return run(t, bin, allArgs...)
}

// ── version ───────────────────────────────────────────────────────────────────

func TestVersion(t *testing.T) {
	bin := binary(t)
	stdout, _, code := run(t, bin, "version")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "claude-desktop-config version")
}

// ── help ──────────────────────────────────────────────────────────────────────

func TestHelp(t *testing.T) {
	bin := binary(t)
	// Use --help so we get the help text without launching the TUI.
	stdout, _, _ := run(t, bin, "--help")
	assert.Contains(t, stdout, "claude-desktop-config")
	assert.Contains(t, stdout, "apply")
	assert.Contains(t, stdout, "dry-run")
	assert.Contains(t, stdout, "tui")
	assert.Contains(t, stdout, "init")
	assert.Contains(t, stdout, "version")
}

// ── dry-run ───────────────────────────────────────────────────────────────────

func TestDryRun(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	stdout, _, code := runWithSource(t, bin, source, "dry-run")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "Dry run")
	assert.Contains(t, stdout, "mcpServers")
	assert.Contains(t, stdout, "Enabled")
	assert.Contains(t, stdout, "Disabled")
	// local-server-a is enabled in the example — should appear in output
	assert.Contains(t, stdout, "local-server-a")
	// prod-server-a is disabled — should not appear in the generated config block
	assert.NotContains(t, stdout, "prod-server-a")
}

func TestDryRunJSON(t *testing.T) {
	bin := binary(t)
	source := filepath.Join("testdata", "source.example.json")
	stdout, _, code := runWithSource(t, bin, source, "dry-run")
	assert.Equal(t, 0, code)

	// Extract the JSON block between the header line and === Summary ===
	const header = "=== Dry run"
	const summary = "=== Summary ==="
	headerIdx := strings.Index(stdout, header)
	summaryIdx := strings.Index(stdout, summary)
	require.Greater(t, headerIdx, -1, "missing dry-run header")
	require.Greater(t, summaryIdx, -1, "missing summary header")

	afterHeader := stdout[headerIdx:]
	nl := strings.Index(afterHeader, "\n")
	require.Greater(t, nl, -1)
	afterHeader = afterHeader[nl+1:]
	sumStart := strings.Index(afterHeader, summary)
	require.Greater(t, sumStart, -1)

	jsonStr := strings.TrimSpace(afterHeader[:sumStart])
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed), "dry-run output should contain valid JSON")

	servers, ok := parsed["mcpServers"].(map[string]interface{})
	require.True(t, ok, "mcpServers key should be present")
	assert.Contains(t, servers, "local-server-a")
	assert.NotContains(t, servers, "prod-server-a")
}

// ── apply ─────────────────────────────────────────────────────────────────────

func TestApplyWritesDestFile(t *testing.T) {
	bin := binary(t)

	src, err := os.ReadFile(filepath.Join("testdata", "source.example.json"))
	require.NoError(t, err)

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "source.json")
	require.NoError(t, os.WriteFile(sourceFile, src, 0o644))

	// apply writes to the OS-detected Claude Desktop path, which we can't
	// override from outside, so we verify the command exits cleanly and
	// reports the correct summary — the config package unit tests cover the
	// actual file write logic.
	stdout, _, code := runWithSource(t, bin, sourceFile, "apply")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "generated successfully")
	assert.Contains(t, stdout, "Enabled")
	assert.Contains(t, stdout, "Disabled")
}

// ── init ──────────────────────────────────────────────────────────────────────

func TestInitCreatesFromExample(t *testing.T) {
	bin := binary(t)

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
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "already exists")
}

// ── tui ───────────────────────────────────────────────────────────────────────

func TestTUIHelp(t *testing.T) {
	bin := binary(t)
	// --help on tui subcommand should print usage without launching the program.
	stdout, _, code := run(t, bin, "tui", "--help")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "tui")
	assert.Contains(t, stdout, "space")
	assert.Contains(t, stdout, "apply")
}

// ── error cases ───────────────────────────────────────────────────────────────

func TestExitCode_MissingSourceFile(t *testing.T) {
	bin := binary(t)
	_, stderr, code := runWithSource(t, bin, "/tmp/does-not-exist-ever.json", "dry-run")
	assert.NotEqual(t, 0, code)
	assert.Contains(t, stderr, "error")
}

func TestExitCode_UnknownSubcommand(t *testing.T) {
	bin := binary(t)
	_, _, code := run(t, bin, "doesnotexist")
	assert.NotEqual(t, 0, code)
}
