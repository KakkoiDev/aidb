# Plan: fix - path handling in add/remove (absolute paths, cwd-dependent layout, `..` escape)

Handoff plan for an AI agent. Self-contained. Go project (cobra CLI). Three related defects in
how add/remove turn CLI arguments into store paths. All verified against source.

## Symptoms

1. **Absolute paths are mangled.** `aidb add /abs/path/file.md` fails "file not found":
   `runAdd` does `filepath.Glob(filepath.Join(cwd, arg))` and falls back to
   `filepath.Join(cwd, arg)` (`cmd/aidb/cmd/add.go:44,50`). `filepath.Join(cwd, "/abs/x")`
   produces `cwd/abs/x`. Same bug in remove: `linkPath := filepath.Join(cwd, filename)`
   (`cmd/aidb/cmd/remove.go:46`).
2. **Storage layout depends on invocation cwd inside git repos.** The namespace comes from the
   repo TOPLEVEL (`internal/config/config.go:70-77`, `getGitRepoName` = basename of
   `rev-parse --show-toplevel`), but the path inside it comes from CWD:
   `relPath, _ = filepath.Rel(cwd, srcPath)` (`add.go:94-97`). From the repo root,
   `aidb add docs/x.md` stores `<repo>/<branch>/docs/x.md`; from `docs/`, `aidb add x.md` stores
   `<repo>/<branch>/x.md`. Same project file, two store locations, depending on where the agent
   happened to cd. Duplicate/foreign entries accumulate silently.
3. **`..` escapes the project namespace.** `aidb add ../other/x.md` joins a relPath containing
   `..` onto storageDir (`add.go:100`), landing in a sibling project's namespace; enough `..`
   lands outside `~/.aidb` entirely. No containment check exists.
4. **(minor, same surface)** `filepath.HasPrefix` (`add.go:87`) is deprecated and is a pure
   string-prefix check - a symlink into `~/.aidb-other/` matches the `~/.aidb` prefix and is
   misreported "already tracked". Same pattern with `strings.HasPrefix` in `remove.go:64`.

## Fix

1. Argument resolution (add + remove): if `filepath.IsAbs(arg)`, use it as-is (glob it as-is);
   otherwise join to cwd as today. One helper, used by both commands.
2. In-repo layout: when in a git repo, compute relPath relative to the repo toplevel, not cwd.
   `getGitRepoName` already runs `rev-parse --show-toplevel` (`config.go:131-138`) - export a
   `GetGitToplevel(dir)` (path, not basename) and use it in `addFile`. Non-git mode keeps
   cwd-relative behavior (matches the home-relative namespace, `config.go:80-96`).
3. Containment check: after computing dstPath, require
   `rel, err := filepath.Rel(storageDir, dstPath)` with `err == nil && !strings.HasPrefix(rel, "..")`.
   Refuse with a clear error ("outside project namespace") otherwise.
4. Replace both prefix checks with a path-aware containment helper (Rel + not-`..`), shared.

## Acceptance criteria

- [ ] `aidb add /abs/path/to/file.md` works from any cwd.
- [ ] `aidb remove /abs/path/to/tracked-symlink.md` works.
- [ ] Inside a git repo: adding the same file from the repo root and from a subdirectory yields
      the SAME store path (toplevel-relative).
- [ ] `aidb add ../escape.md` from a repo/project dir is refused with a clear error; nothing is
      moved.
- [ ] A symlink pointing into `~/.aidb-other/` is NOT treated as already-tracked.
- [ ] Existing store layout untouched for the common case (file added from repo toplevel) - no
      migration required.

## Tests (fail-on-revert required)

- Extend `cmd/aidb/cmd/add_test.go`: absolute-path add; subdir-vs-root add asserting identical
  store path (must FAIL today); `../` add asserting refusal and no file moved.
- Extend `cmd/aidb/cmd/remove_test.go`: absolute-path remove.
- Use `internal/testutil` (`InitGitRepoWithBranch` already exists).

## Edge cases / notes

- Behavior change: files previously added from subdirs live at cwd-relative store paths; new adds
  of the same file will go to the toplevel-relative path and hit "already exists in database"
  only if names collide. Do NOT write a migrator; note the change in README. The memory-framework
  redesign owns any store re-layout.
- Keep `addDirectory` consistent: its relPath is relative to the added dir (correct); only the
  dir's own anchor path changes with fix 2.

## Out of scope (do NOT bundle)

- Same-basename repo collisions and detached HEAD - `PLAN-fix-namespace-collisions.md`.
- Exit-code policy - `PLAN-fix-exit-codes-and-false-success.md`.

## Verify

```
cd ~/Code/aidb && go build ./... && go test ./...
# manual: temp git repo, add same file from root and subdir, diff store layout.
```
