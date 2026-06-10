# Plan: fix - aidb cannot commit modifications to already-tracked files

Handoff plan for an AI agent. Self-contained. Go project (cobra CLI). Fix the bug, add a regression
test that fails on revert, update the test that currently locks the buggy behavior.

## Symptom

Editing the content of an already-tracked file and trying to persist it fails through aidb:

```
# (file already tracked: it is a symlink in cwd -> ~/.aidb/<...>)
$ aidb commit "update doc"
[INFO] Nothing staged to commit          # nothing got staged

$ aidb add doc.md
[ERROR] doc.md: already tracked          # refuses, still nothing staged
```

`git -C ~/.aidb status -s` shows ` M <path>` (modified, UNSTAGED). There is no aidb command that
stages a modification to an already-tracked file, so the edit can never be committed via aidb. The
only current workaround is raw `git -C ~/.aidb add <path>` then `aidb commit`, which the project's
usage rules explicitly forbid ("always use aidb, never raw git").

This contradicts aidb's own help, which advertises updates as supported:
`commit.go` Long example -> `aidb commit "Update TASK.md with new requirements"`.

## Root cause (two gaps, both must be closed)

1. `cmd/aidb/cmd/add.go:84-91` - `addFile` short-circuits on an already-tracked symlink by returning
   an error, instead of re-staging the current content:
   ```go
   if info.Mode()&os.ModeSymlink != 0 {
       target, _ := os.Readlink(srcPath)
       if filepath.HasPrefix(target, cfg.DBDir) {
           return fmt.Errorf("already tracked")   // <- bails, never stages
       }
       return fmt.Errorf("is a symlink")
   }
   ```
2. `cmd/aidb/cmd/commit.go:45-50` - `runCommit` only checks for already-staged changes and commits
   them; it never stages anything:
   ```go
   gitCmd := exec.Command("git", "-C", cfg.DBDir, "diff", "--cached", "--quiet")
   if err := gitCmd.Run(); err == nil {
       printInfo("Nothing staged to commit")
       return nil
   }
   ```

New files work only because `addFile` stages them on first add (`add.go:130`). Modifications hit gap 1
(add refuses) and gap 2 (commit doesn't stage) -> dead end.

## Recommended fix (implement BOTH - they serve the two mental models)

### Fix A - `add` becomes idempotent: re-stage an already-tracked file instead of erroring
In `add.go:84-91`, when the symlink target is under `cfg.DBDir`, do not error. Re-stage the target
and report it. Replace the `return fmt.Errorf("already tracked")` branch with:
```go
if filepath.HasPrefix(target, cfg.DBDir) {
    gitCmd := exec.Command("git", "-C", cfg.DBDir, "add", target)
    if err := gitCmd.Run(); err != nil {
        return fmt.Errorf("failed to re-stage: %w", err)
    }
    printSuccess(fmt.Sprintf("Re-staged %s", filepath.Base(target)))
    return nil
}
```
(`git add` on a symlink target path stages the real file; the symlink in cwd is irrelevant to the
DBDir repo.)
Note: `filepath.HasPrefix` is deprecated (pure string-prefix check; `~/.aidb-other` matches the
`~/.aidb` prefix). `PLAN-fix-path-handling.md` replaces it repo-wide - keep it as-is here to stay
minimal; do not duplicate that fix.

### Fix B - `commit` auto-stages tracked modifications (matches the documented one-step workflow)
In `commit.go`, before the staged-check at line 46, stage modifications + deletions to already-tracked
files:
```go
stage := exec.Command("git", "-C", cfg.DBDir, "add", "-u")
if err := stage.Run(); err != nil {
    return fmt.Errorf("failed to stage tracked changes: %w", err)
}
```
Use `add -u` (NOT `add -A`): it stages edits/deletions to tracked files but ignores untracked files,
so brand-new files still go through `aidb add` (preserving the move-to-storage + symlink step). After
this, the existing "Nothing staged" guard still correctly handles the truly-clean case.

## Acceptance criteria

- [ ] Edit a tracked file, run `aidb commit "msg"` -> the edit is committed (no "Nothing staged").
- [ ] `aidb add <already-tracked-file>` re-stages its current content and exits 0 (no "already tracked" error).
- [ ] Adding a brand-new file still moves it to storage, symlinks, and stages (unchanged behavior).
- [ ] `aidb commit` on a genuinely clean tree still prints "Nothing staged to commit" and exits 0.
- [ ] A deleted tracked file (e.g. after `aidb remove`) is committable via `aidb commit`.
- [ ] No raw-git workaround is needed for any edit-then-commit flow.

## Tests (a fix without a failing-on-revert test is not done)

- ADD `cmd/aidb/cmd/commit_test.go` (does not exist yet). Regression test: init temp db, add a file,
  modify the stored file content, run the commit command, assert `git log`/`git show` contains the new
  content and the working tree is clean. Must FAIL against current `commit.go` and PASS after Fix B.
- UPDATE `cmd/aidb/cmd/add_test.go:87` - the case commented "Try to add - should report already
  tracked" asserts NOTHING (it runs `rootCmd.Execute()` and discards the result, add_test.go:88-90;
  a smoke test, not a behavior lock). Make it assert re-staging succeeds (Fix A): modify the stored
  file, run add, assert the change is staged in the DBDir repo.
- Reuse `internal/testutil` for temp-repo scaffolding (see existing `*_test.go` for the pattern).

## Edge cases / notes

- `.metadata.json` (seen-state) is currently tracked and committed; `add -u` will sweep its changes
  into commits too. Acceptable for this fix. (Whether seen-state should be git-tracked at all is a
  SEPARATE open decision - see "Out of scope".)
- Concurrent edits / partial commits: `add -u` stages ALL tracked edits in the store. For a dedicated
  knowledge repo this is fine; there is no partial-commit use case here. Document it in `commit.go` Long.
- Keep the change minimal. No new flags required. (If a reviewer wants opt-out, a `--no-stage` flag on
  commit is the only acceptable addition - do not add speculative options.)

## Out of scope (do NOT bundle)

- Redesign/removal of the seen/unseen feature and `.metadata.json` git-tracking. Tracked separately.
- Any change to push/pull/init.

## Verify

```
cd ~/Code/aidb
go build ./... && go test ./...
# manual: init a temp db, add a file, edit it, `aidb commit "x"`, confirm committed; `git status` clean.
```
