package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestListCommand_AidbFlag_ExcludesAidbByDefault(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create regular files in database
	projectDir := filepath.Join(env.DBDir, "myproject", "main")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "MEMO.md"), []byte("# Memo"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "TASK.md"), []byte("# Task"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create _aidb/ knowledge files
	aidbDir := filepath.Join(projectDir, "_aidb")
	if err := os.MkdirAll(aidbDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aidbDir, "patterns.md"), []byte("# Patterns"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run list --json (without --aidb flag)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"list", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	var entries []FileEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Should include MEMO.md and TASK.md but NOT _aidb/patterns.md
	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Path)
	}

	hasRegular := false
	hasAidb := false
	for _, p := range paths {
		if p == "myproject/main/MEMO.md" || p == "myproject/main/TASK.md" {
			hasRegular = true
		}
		if p == "myproject/main/_aidb/patterns.md" {
			hasAidb = true
		}
	}

	if !hasRegular {
		t.Error("list should include regular files (MEMO.md, TASK.md)")
	}
	if hasAidb {
		t.Error("list without --aidb should NOT include _aidb/ files")
	}
}

func TestListCommand_AidbFlag_ShowsOnlyAidbFiles(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create regular files in database
	projectDir := filepath.Join(env.DBDir, "myproject", "main")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "MEMO.md"), []byte("# Memo"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create _aidb/ knowledge files
	aidbDir := filepath.Join(projectDir, "_aidb")
	if err := os.MkdirAll(aidbDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aidbDir, "patterns.md"), []byte("# Patterns"), 0644); err != nil {
		t.Fatal(err)
	}

	// Also create global _aidb at root
	globalAidb := filepath.Join(env.DBDir, "_aidb")
	if err := os.MkdirAll(globalAidb, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalAidb, "global.md"), []byte("# Global"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run list --json --aidb
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"list", "--json", "--aidb"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	var entries []FileEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Should include ONLY _aidb files
	hasRegular := false
	hasAidb := false
	for _, e := range entries {
		if e.Path == "myproject/main/MEMO.md" {
			hasRegular = true
		}
		if e.Path == "myproject/main/_aidb/patterns.md" || e.Path == "_aidb/global.md" {
			hasAidb = true
		}
	}

	if hasRegular {
		t.Error("list --aidb should NOT include regular files")
	}
	if !hasAidb {
		t.Error("list --aidb should include _aidb/ files")
	}
}

func TestListCommand_AidbFlag_WithUnseen(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create _aidb/ knowledge files
	projectDir := filepath.Join(env.DBDir, "myproject", "main")
	aidbDir := filepath.Join(projectDir, "_aidb")
	if err := os.MkdirAll(aidbDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aidbDir, "patterns.md"), []byte("# Patterns"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run list --unseen --aidb --json
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"list", "--unseen", "--aidb", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	var entries []FileEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Should include _aidb files that are unseen
	if len(entries) == 0 {
		t.Error("list --unseen --aidb should show unseen _aidb files")
	}

	for _, e := range entries {
		if e.Seen {
			t.Errorf("list --unseen should only show unseen files, got seen: %s", e.Path)
		}
	}
}
