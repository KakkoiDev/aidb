package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KakkoiDev/aidb/internal/metadata"
	"github.com/KakkoiDev/aidb/internal/testutil"
)

// trackFile adds a file via aidb and returns its store-relative path
func trackFile(t *testing.T, env *testutil.TestEnv, repoDir, name, content string) string {
	t.Helper()

	env.CreateFile(filepath.Join(repoDir, name), content)
	rootCmd.SetArgs([]string{"add", name})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	target, err := os.Readlink(filepath.Join(repoDir, name))
	if err != nil {
		t.Fatal(err)
	}
	relPath, err := filepath.Rel(env.DBDir, target)
	if err != nil {
		t.Fatal(err)
	}
	return relPath
}

func seenTestEnv(t *testing.T) (*testutil.TestEnv, string) {
	t.Helper()

	env := testutil.New(t)
	t.Cleanup(env.Cleanup)
	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()
	return env, repoDir
}

func TestSeenCommand_KnownPath(t *testing.T) {
	env, repoDir := seenTestEnv(t)
	relPath := trackFile(t, env, repoDir, "TASK.md", "# Task")

	rootCmd.SetArgs([]string{"seen", relPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("seen failed: %v", err)
	}

	meta, err := metadata.New(env.DBDir)
	if err != nil {
		t.Fatal(err)
	}
	info := meta.GetInfo(relPath)
	if info == nil || !info.Seen {
		t.Errorf("%s should be marked seen", relPath)
	}
}

func TestSeenCommand_Nonexistent(t *testing.T) {
	seenTestEnv(t)

	rootCmd.SetArgs([]string{"seen", "does/not/exist.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("seen on nonexistent path should fail")
	}
}

func TestSeenCommand_CwdSymlink(t *testing.T) {
	env, repoDir := seenTestEnv(t)
	relPath := trackFile(t, env, repoDir, "TASK.md", "# Task")

	// Store-relative is the documented form; a cwd-relative tracked symlink resolves too
	rootCmd.SetArgs([]string{"seen", "TASK.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("seen on tracked symlink failed: %v", err)
	}

	meta, err := metadata.New(env.DBDir)
	if err != nil {
		t.Fatal(err)
	}
	info := meta.GetInfo(relPath)
	if info == nil || !info.Seen {
		t.Errorf("%s should be marked seen via cwd symlink", relPath)
	}
}

func TestUnseenCommand_KnownPath(t *testing.T) {
	env, repoDir := seenTestEnv(t)
	relPath := trackFile(t, env, repoDir, "TASK.md", "# Task")

	rootCmd.SetArgs([]string{"seen", relPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("seen failed: %v", err)
	}
	rootCmd.SetArgs([]string{"unseen", relPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unseen failed: %v", err)
	}

	meta, err := metadata.New(env.DBDir)
	if err != nil {
		t.Fatal(err)
	}
	info := meta.GetInfo(relPath)
	if info == nil || info.Seen {
		t.Errorf("%s should be marked unseen", relPath)
	}
}

func TestUnseenCommand_Nonexistent(t *testing.T) {
	seenTestEnv(t)

	rootCmd.SetArgs([]string{"unseen", "does/not/exist.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("unseen on nonexistent path should fail, not report success")
	}
}
