# Plan: fix - store namespace collisions (same-basename repos, detached HEAD)

Handoff plan for an AI agent. Self-contained. Go project (cobra CLI). The store key for a git
project is the repo directory BASENAME plus branch; two different projects can silently share a
namespace. Detection-level fix only - no store re-layout.

## Symptoms

1. **Same-basename repos collide.** Project key = `filepath.Base(rev-parse --show-toplevel)`
   (`internal/config/config.go:131-138`, used at `config.go:70-77`). `~/Code/foo` and
   `~/Work/foo` both store under `~/.aidb/foo/<branch>/`. Consequences: cross-project knowledge
   contamination (agents read the wrong project's MEMO/TASK), spurious "already exists in
   database" on add, wrong-file restores on remove.
2. **Detached HEAD writes to a junk namespace.** `getGitBranch` returns the literal string
   `HEAD` when detached (`config.go:141-148`; `rev-parse --abbrev-ref HEAD` prints `HEAD`).
   Adds from CI checkouts, bisects, or detached worktrees silently land in
   `~/.aidb/<repo>/HEAD/`.

## Fix (detection, not re-keying)

1. **Origin pinning.** On first add for a project namespace, write
   `~/.aidb/<project>/.origin` containing the repo toplevel path and `origin` remote URL (use
   `GetRemoteURL` from `cmd/aidb/cmd/init.go:93-100`; empty allowed). On every later add,
   compare: same remote URL (when both non-empty) OR same toplevel path = OK; mismatch =
   hard error naming both paths and suggesting a rename. Storage layout unchanged; existing
   namespaces get `.origin` lazily on next add.
   - `.origin` is store metadata: exclude it from `aidb list` (extend the skip at
     `cmd/aidb/cmd/list.go:83-86`, which today skips only `.metadata.json`).
2. **Detached HEAD: refuse.** In `GetProjectFromCwd`/`GetStoragePath`, if branch == "HEAD",
   return an error ("detached HEAD - check out a branch before adding"). A transient SHA
   namespace has no retrieval value; refusing is honest. Do not invent a fallback namespace.

## Explicitly rejected alternative

Re-keying namespaces by remote-URL hash or full path is a breaking store migration touching
every symlink in every project. Not worth it now; the memory-framework redesign
(`PLAN-aidb-memory-framework.md`) owns any layout change. Record the constraint there if it
lands.

## Acceptance criteria

- [ ] Two temp repos with the same basename: first add succeeds and writes `.origin`; second
      repo's add fails with an error naming both toplevel paths.
- [ ] Same repo, second add (origin matches): no error, no behavior change.
- [ ] Pre-existing namespace without `.origin`: next add writes it, does not error.
- [ ] Detached HEAD checkout: `aidb add` refuses with a clear message; nothing moved.
- [ ] `.origin` never appears in `aidb list` output (any flags).

## Tests (fail-on-revert required)

- ADD cases to `cmd/aidb/cmd/add_test.go` using `internal/testutil`: collision (two repos, same
  basename - must FAIL today by silently sharing the namespace), origin-match pass,
  lazy-pin pass, detached-HEAD refusal (`git checkout --detach` in the temp repo).
- Extend `cmd/aidb/cmd/list_test.go`: `.origin` exclusion.

## Out of scope (do NOT bundle)

- cwd-dependent layout and `..` escapes - `PLAN-fix-path-handling.md`.
- Store re-layout / migration of any kind.

## Verify

```
cd ~/Code/aidb && go build ./... && go test ./...
# manual: mkdir -p /tmp/a/foo /tmp/b/foo; git init both; add a file from each; second must refuse.
```
