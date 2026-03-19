package output

import (
	"fmt"
	"io"
	"os"
)

// Writer is a minimal structured output helper.
type Writer struct {
	out     io.Writer
	err     io.Writer
	verbose bool
	color   bool
}

// New returns a Writer using os.Stdout and os.Stderr.
func New(verbose bool) *Writer {
	return &Writer{
		out:     os.Stdout,
		err:     os.Stderr,
		verbose: verbose,
		color:   isTerminal(os.Stdout),
	}
}

// NewWithWriters returns a Writer using the provided io.Writers (useful in tests).
func NewWithWriters(out, err io.Writer, verbose bool) *Writer {
	return &Writer{out: out, err: err, verbose: verbose}
}

func (w *Writer) Info(msg string)    { _, _ = fmt.Fprintln(w.out, msg) }
func (w *Writer) Success(msg string) { _, _ = fmt.Fprintln(w.out, w.green("✓ ")+msg) }
func (w *Writer) Warn(msg string)    { _, _ = fmt.Fprintln(w.err, w.yellow("warn: ")+msg) }
func (w *Writer) Error(msg string)   { _, _ = fmt.Fprintln(w.err, w.red("error: ")+msg) }

func (w *Writer) Verbose(msg string) {
	if w.verbose {
		_, _ = fmt.Fprintln(w.out, w.dim("  "+msg))
	}
}

func (w *Writer) Command(cmd string) {
	_, _ = fmt.Fprintln(w.out, w.dim("$ "+cmd))
}

func (w *Writer) green(s string) string  { return w.ansi("\033[32m", s) }
func (w *Writer) yellow(s string) string { return w.ansi("\033[33m", s) }
func (w *Writer) red(s string) string    { return w.ansi("\033[31m", s) }
func (w *Writer) dim(s string) string    { return w.ansi("\033[2m", s) }

func (w *Writer) ansi(code, s string) string {
	if !w.color {
		return s
	}
	return code + s + "\033[0m"
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
