package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Options controls output behavior
type Options struct {
	JSON    bool
	Quiet   bool
	NoColor bool
	Debug   bool
	Writer  io.Writer
	ErrWriter io.Writer
}

// Output handles all CLI output
type Output struct {
	opts Options
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
	return o.isTTY && !o.opts.NoColor
}

// color wraps text in ANSI color codes
func (o *Output) color(code, text string) string {
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
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.color("0;34", "[INFO]"), msg)
}

// Success prints success message
func (o *Output) Success(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.color("0;32", "✓"), msg)
}

// Error prints error message
func (o *Output) Error(msg string) {
	fmt.Fprintf(o.opts.ErrWriter, "%s %s\n", o.color("0;31", "✗"), msg)
}

// Warning prints warning message
func (o *Output) Warning(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, "%s %s\n", o.color("1;33", "!"), msg)
}

// Debug prints debug message if debug enabled
func (o *Output) Debug(msg string) {
	if !o.opts.Debug {
		return
	}
	fmt.Fprintf(o.opts.ErrWriter, "%s %s\n", o.color("0;36", "[DEBUG]"), msg)
}

// Print prints plain text
func (o *Output) Print(msg string) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintln(o.opts.Writer, msg)
}

// Printf prints formatted text
func (o *Output) Printf(format string, args ...interface{}) {
	if o.opts.Quiet {
		return
	}
	fmt.Fprintf(o.opts.Writer, format, args...)
}

// JSON outputs data as JSON
func (o *Output) JSON(data interface{}) error {
	enc := json.NewEncoder(o.opts.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// IsJSON returns true if JSON output is enabled
func (o *Output) IsJSON() bool {
	return o.opts.JSON
}

// IsQuiet returns true if quiet mode is enabled
func (o *Output) IsQuiet() bool {
	return o.opts.Quiet
}
