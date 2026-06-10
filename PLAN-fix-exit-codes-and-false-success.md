# Plan: fix - aidb reports success on failure (exit codes + false positives)

Handoff plan for an AI agent. Self-contained. Go project (cobra CLI). aidb is driven by
unattended agents (see AGENTS.md); an exit code of 0 on failure means the caller cannot detect
breakage. This happened in a real session: `aidb add` (failed: "already tracked") and
`aidb commit` (failed: nothing staged) both exited 0 and work was silently not persisted.

## Symptoms (all verified against source)

1. `aidb add nonexistent.md` prints a red error, exits 0. `runAdd` prints per-file errors and
   `continue`s, then returns nil unconditionally - `cmd/aidb/cmd/add.go:67-75`.
2. `aidb seen nope.md` prints a red error, exits 0. Same swallow-and-continue pattern -
   `cmd/aidb/cmd/seen.go:42-70` (count stays 0, returns nil).
3. `aidb unseen anything-at-all` prints SUCCESS for paths that exist nowhere. `runUnseen` does no
   existence check at all (`cmd/aidb/cmd/unseen.go:52-61`) and `metadata.MarkUnseen` silently
   no-ops on unknown paths (`internal/metadata/metadata.go:74-78`). False success, exit 0.
4. `addDirectory` swallows walk errors (`add.go:140-142` returns nil on err) and ignores git-add
   errors (`add.go:169` bare `gitCmd.Run()`). Partial directory adds report nothing.
5. `addFile` downgrades a git-add failure to a warning (`add.go:130-133`). The file is then
   tracked-but-unstaged - combined with the commit bug (see PLAN-fix-staging-tracked-files.md)
   it becomes uncommittable through aidb.

## Fix policy (small, mechanical)

- Per-item errors keep printing per-item (current UX is right), but the command returns a
  non-nil error if ANY item failed: count failures in the loop, return
  `fmt.Errorf("%d of %d failed", nFail, nTotal)` at the end. Applies to add, seen, unseen.
- `unseen`: before printing success, require the path to exist in metadata
  (`meta.GetInfo(relPath) != nil`) OR on disk under the store. Unknown path = per-item error.
- `addDirectory`: propagate walk errors and git-add errors (collect, report, non-zero exit).
- `addFile` git-add failure: keep the file move + symlink (do not roll back), but count it as a
  failure, not a warning - the "stage in git" half of the contract was not delivered.

## Optional (same surface, do only if trivial after the above)

- Path-argument UX: seen/unseen take store-relative paths while add/remove take cwd-relative
  (`seen.go:43` joins to `cfg.DBDir`). Accept a cwd symlink too: if the arg resolves to a symlink
  whose target is under the store, operate on that target. Keeps AGENTS.md's `aidb seen <file>`
  honest. Skip if it grows beyond ~20 lines.

## Acceptance criteria

- [ ] `aidb add nonexistent.md` exits non-zero.
- [ ] `aidb add good.md nonexistent.md` adds good.md AND exits non-zero (partial failure visible).
- [ ] `aidb seen nonexistent` exits non-zero.
- [ ] `aidb unseen nonexistent` prints an error (not success) and exits non-zero.
- [ ] `aidb unseen <known-path>` still works and exits 0.
- [ ] A directory add with one unmovable file reports it and exits non-zero.
- [ ] All currently-passing tests still pass (none assert exit 0 on failure - verified; the
      closest is add_test.go:87-91 which asserts nothing and is being fixed by the staging plan).

## Tests (fix without a failing-on-revert test is not done)

- Extend `cmd/aidb/cmd/add_test.go`: nonexistent-file case asserting `rootCmd.Execute()` returns
  an error. Must FAIL on current code (returns nil today).
- ADD `cmd/aidb/cmd/seen_test.go` and cover seen + unseen (both directions: false success today,
  error after). No seen/unseen tests exist at all today.
- Reuse `internal/testutil` scaffolding (pattern in existing `*_test.go`).
- Note: cobra prints the returned error and sets exit code via `Execute()`; assert on the
  returned error, not os.Exit.

## Out of scope (do NOT bundle)

- The staging bug itself - `PLAN-fix-staging-tracked-files.md`.
- Path mangling (absolute paths, `..` escapes) - `PLAN-fix-path-handling.md`.
- backup-run failure handling - framework plan Phase 0b.

## Verify

```
cd ~/Code/aidb
go build ./... && go test ./...
./aidb add /tmp/does-not-exist.md; echo "exit=$?"   # must be non-zero
```
