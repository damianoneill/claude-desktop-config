package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/damianoneill/claude-desktop-config/internal/output"
)

func TestWriter_Info(t *testing.T) {
	var out bytes.Buffer
	w := output.NewWithWriters(&out, &bytes.Buffer{}, false)
	w.Info("hello")
	assert.Equal(t, "hello\n", out.String())
}

func TestWriter_Success(t *testing.T) {
	var out bytes.Buffer
	w := output.NewWithWriters(&out, &bytes.Buffer{}, false)
	w.Success("done")
	assert.Contains(t, out.String(), "done")
	assert.Contains(t, out.String(), "✓")
}

func TestWriter_Warn(t *testing.T) {
	var errBuf bytes.Buffer
	w := output.NewWithWriters(&bytes.Buffer{}, &errBuf, false)
	w.Warn("careful")
	assert.Contains(t, errBuf.String(), "careful")
	assert.Contains(t, errBuf.String(), "warn:")
}

func TestWriter_Error(t *testing.T) {
	var errBuf bytes.Buffer
	w := output.NewWithWriters(&bytes.Buffer{}, &errBuf, false)
	w.Error("boom")
	assert.Contains(t, errBuf.String(), "boom")
	assert.Contains(t, errBuf.String(), "error:")
}

func TestWriter_Verbose_WhenEnabled(t *testing.T) {
	var out bytes.Buffer
	w := output.NewWithWriters(&out, &bytes.Buffer{}, true)
	w.Verbose("detail")
	assert.Contains(t, out.String(), "detail")
}

func TestWriter_Verbose_WhenDisabled(t *testing.T) {
	var out bytes.Buffer
	w := output.NewWithWriters(&out, &bytes.Buffer{}, false)
	w.Verbose("detail")
	assert.Empty(t, out.String())
}

func TestWriter_Command(t *testing.T) {
	var out bytes.Buffer
	w := output.NewWithWriters(&out, &bytes.Buffer{}, false)
	w.Command("go build ./...")
	assert.Contains(t, out.String(), "go build ./...")
	assert.Contains(t, out.String(), "$")
}
