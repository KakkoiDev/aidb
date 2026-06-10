package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestConfigCommand_DBPathWired(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	custom := filepath.Join(env.TempDir, "customdb")
	rootCmd.SetArgs([]string{"config", "db.path", custom})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	cfg, err := config.New()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBDir != custom {
		t.Errorf("DBDir = %q, want %q (db.path must be wired, not a placebo)", cfg.DBDir, custom)
	}
}

func TestConfigCommand_BackupEnabledRejected(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	rootCmd.SetArgs([]string{"config", "backup.enabled", "true"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("setting backup.enabled must be rejected (managed by 'aidb backup')")
	}
	if !strings.Contains(err.Error(), "aidb backup") {
		t.Errorf("rejection should point at 'aidb backup enable/disable', got: %v", err)
	}
}
