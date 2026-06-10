package output

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Options controls output behavior
type Options struct {
	Quiet     bool
	NoColor   bool
	Writer    io.Writer
	ErrWriter io.Writer
}

// Output handles all CLI output
type Output struct {
	opts  Options
	isTTY bool
}

// New creates a new Output with options
func New(opts Options) *Output {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.ErrWriter == nil {
		opts.ErrWriter = os.Stderr
	}

	// Check if stdout is a TTY
	isTTY := false
	if f, ok := opts.Writer.(*os.File); ok {
		isTTY = term.IsTerminal(int(f.Fd()))
	}

	return &Output{
		opts:  opts,
		isTTY: isTTY,
	}
}

// Default creates output with default options
func Default() *Output {
	return New(Options{})
}

// colorEnabled returns true if colors should be used
func (o *Output) colorEnabled() bool {
	return o.isTTY && !o.opts.NoColor && os.Getenv("NO_COLOR") == ""
}

// Colorize wraps text in ANSI color codes when color is enabled
func (o *Output) Colorize(code, text string) string {
	if !o.colorEnabled() {
		return text
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
}

// Info prints informational message
func (o *Output) Info(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.Colorize("0;34", "[INFO]"), msg)
}

// Success prints success message
func (o *Output) Success(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.Colorize("0;32", "✓"), msg)
}

// Error prints error message
func (o *Output) Error(msg string) {
	fmt.Fprintf(o.opts.ErrWriter, "%s %s\n", o.Colorize("0;31", "✗"), msg)
}

// Warning prints warning message
func (o *Output) Warning(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.Colorize("1;33", "!"), msg)
}
