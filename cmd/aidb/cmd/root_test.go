package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

// captureStdout swaps os.Stdout for a pipe (non-TTY) while fn runs
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestPipedOutputHasNoANSI(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	got := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"status"})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("status failed: %v", err)
		}
	})

	if strings.Contains(got, "\033") {
		t.Errorf("piped output must not contain ANSI escapes, got %q", got)
	}
}

func TestGlobalJSONFlagRemoved(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	rootCmd.SetArgs([]string{"status", "--json"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("status --json should be an unknown flag (global --json removed; only list implements JSON)")
	}
}
