package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutput_NonTTYHasNoANSI(t *testing.T) {
	var buf bytes.Buffer
	o := New(Options{Writer: &buf, ErrWriter: &buf})

	o.Info("info")
	o.Success("success")
	o.Warning("warning")
	o.Error("error")

	if strings.Contains(buf.String(), "\033") {
		t.Errorf("non-TTY output must not contain ANSI escapes, got %q", buf.String())
	}
}

func TestOutput_NoColorEnvDisablesColorOnTTY(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	o := New(Options{Writer: &buf, ErrWriter: &buf})
	o.isTTY = true

	o.Info("info")

	if strings.Contains(buf.String(), "\033") {
		t.Errorf("NO_COLOR must disable ANSI escapes, got %q", buf.String())
	}
}

func TestOutput_NoColorOptionDisablesColorOnTTY(t *testing.T) {
	var buf bytes.Buffer
	o := New(Options{Writer: &buf, ErrWriter: &buf, NoColor: true})
	o.isTTY = true

	o.Info("info")

	if strings.Contains(buf.String(), "\033") {
		t.Errorf("NoColor must disable ANSI escapes, got %q", buf.String())
	}
}

func TestOutput_QuietSuppressesAllButErrors(t *testing.T) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	o := New(Options{Writer: &buf, ErrWriter: &errBuf, Quiet: true})

	o.Info("info")
	o.Success("success")
	o.Warning("warning")
	o.Error("error")

	if buf.String() != "" {
		t.Errorf("quiet mode must suppress non-error output, got %q", buf.String())
	}
	if !strings.Contains(errBuf.String(), "error") {
		t.Error("quiet mode must not suppress errors")
	}
}
