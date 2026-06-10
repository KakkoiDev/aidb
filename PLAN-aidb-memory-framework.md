# Plan: aidb -> human-in-the-loop memory framework

Roadmap for evolving aidb from a git-backed file tracker into a memory framework: tracked files are
ground-truth events; a conflict-free shared "seen" ledger drives an automated consolidation that
PROPOSES knowledge; a human promotes proposals into canonical `_aidb/`. TDD throughout. No phase
ships without a test that fails on revert. Build bottom-up: the ledger is the keystone; automation is
last.

> Companion doc: `PLAN-fix-staging-tracked-files.md` is Phase 0a below.

## Context / why

- aidb already harvests knowledge (`AGENTS.md` two-tier promotion) but consolidation is manual and
  the "seen" state does not survive multi-machine use (single git-tracked `.metadata.json`, pull is
  rebase-or-abort, no merge driver -> guaranteed conflicts).
- Goal: knowledge consolidates slowly + automatically over time, shared across machines, WITHOUT
  drift, with a human approving what becomes canonical.
- Design stance (from prior art): event-source the ledger (conflict-free by construction); the cron
  PROPOSES, never rewrites canonical memory (auto-rewrite = drift); the human promotes via their own
  editor (the gap competitors leave open). References: a3nm "git auto-conflict on logs/sets"
  (merge=union); Mem0 drift via accumulated extraction; Zep temporal supersession; MemMachine
  ground-truth preservation.

## Architecture (5 layers)

1. Ledger (Go core) - append-only seen-log + replay; conflict-free via `merge=union`.
2. Consolidation (ICM skill, not Go) - reads unseen files, extracts candidate patterns, writes
   PROPOSALS to an inbox, marks sources seen.
3. Review gate (Go) - `aidb review` (accept/edit/reject proposals) + `aidb encyclopedia` (browse
   canonical knowledge). Editing is markdown in `$EDITOR`.
4. Scheduler (Go + launchd) - `aidb consolidate enable/disable/status`, per-machine hourly job that
   pulls -> runs the skill headless -> commits -> pushes (with retry).
5. Drift safeguards (cross-cutting) - cron proposes only; raw files preserved; human promotes;
   provenance + timestamps; `[STALE: date]`; 2+-project rule for global promotion.

## Keystone design - conflict-free shared seen-state

- Replace `internal/metadata/.metadata.json` with an append-only log `~/.aidb/.seen-events.jsonl`,
  one event per line: `{"file":"rel/path","hash":"sha256:..","seenAt":"<UTC>","machine":"<host>","op":"seen|unseen"}`.
- `init` writes `.gitattributes` at the store root: `.seen-events.jsonl merge=union` (migration adds
  it to existing stores). Union merge = two machines appending different lines auto-merge; identical
  lines dedup; never a conflict.
- Append-only: `aidb seen` appends an `op:seen` event; `aidb unseen` appends an `op:unseen` tombstone.
  No in-place rewrite -> nothing for git to conflict on.
- State by replay: build file -> latest-event (max `seenAt`). `IsSeen(file)` = latest event exists,
  `op==seen`, AND `event.hash == hash(file now)` (so edits auto-unsee, preserving today's semantics).
- Migration: one-time `aidb migrate-metadata` reads `.metadata.json`, emits equivalent `seen` events,
  `git rm --cached .metadata.json`, writes `.gitattributes`. Idempotent; TDD'd.
- Compaction: optional `aidb seen --compact` (rare, manual, run solo) squashes to current state.
  Logs are tiny; default is never compact. Document that compaction reintroduces conflict risk.

## Multi-machine cron (every machine runs it) - dedup / idempotency

- Each `consolidate-run`: `aidb pull` (rebase+autostash) FIRST -> compute unseen AFTER pull ->
  consolidate only those -> append `seen` events + write proposals -> `aidb commit` -> push.
- Push safety (new): pull-before-push + one retry on rejection, in `push` and the run commands.
  Today `push` has no rejection handling - multi-machine needs it.
- Same-hour race (two machines consolidate the same file before either pushes): proposals are keyed
  by `(source-file, source-hash)` so identical proposals collide rather than duplicate; any residual
  duplicates are deduped by the human at review. Union-merge keeps both `seen` events harmlessly.
- Headless caveat: cron runs with zero interactive context; the consolidation skill must be
  LOCAL-ONLY (read files, write markdown) with NO dependency on interactively-authed MCP
  (Notion/Slack/etc.), which may be absent in launchd/cron. State this as a skill constraint.

## Consolidation skill (ICM, modeled on /todo-triage)

`skills/consolidate/SKILL.md` + ordered stage contracts. Reuse `AGENTS.md` categories + quality
filter + 2+-project promotion rule (do not reinvent).
- 01-ingest: the unseen queue (from `aidb list --unseen --json --aidb`) = files to process.
- 02-extract: per-file, pull candidate patterns into the standard categories; apply quality filter
  (capture surprising/reusable/decision-rationale; skip obvious/one-off).
- 03-propose: write candidates to an INBOX (`~/.aidb/_inbox/<date>/...` or `.aidb/proposals/`), each
  snippet carrying provenance (source file, source hash, date, category, suggested tier). NEVER write
  canonical `_aidb/`.
- 04-mark-seen: append `seen` events for processed sources.
Human-in-the-loop is between propose and canonical: the skill stops at proposals.

## Review gate (Go commands)

- `aidb review` - list pending inbox proposals with provenance; `accept <id>` appends the snippet to
  the right `_aidb/` file; `reject <id>` drops it; edit opens `$EDITOR`. Promotion to GLOBAL
  `~/.aidb/_aidb/` keeps the 2+-project rule (skill suggests, human confirms).
- `aidb encyclopedia` - categorized read view of canonical `_aidb/` (flags: `--category`, `--search`).
  Read-only browse; edits happen in `$EDITOR` on the markdown.

## Ordered roadmap (what to do, in what order)

Each task: write the failing test first (repo TASK.md mandates TDD; use `internal/testutil`), then
implement, then update docs. A task is done only when its regression test fails on revert.

Phase 0 - Foundation (smallest, unblocks all)
- 0a. Staging fix - implement `PLAN-fix-staging-tracked-files.md` (add re-stages tracked; commit
  `git add -u`). Tests: new `commit_test.go`; fix `add_test.go:87` (it locks the buggy behavior).
- 0b. Push safety - pull-before-push + one retry on rejection in `push.go`, reused by `backup-run`
  (`backup.go:199-227`) - the URGENT instance: it runs hourly via launchd, does `git add -A` +
  commit + push with NO pull (guaranteed multi-machine rejections), and fails every hour when no
  remote is configured (add a HasRemote no-op guard). Its `add -A` also sweeps untracked junk
  dropped into `~/.aidb` into auto-commits - decide whether that stays.

Phase 1 - Conflict-free shared seen-ledger (keystone)
- 1a. `init` writes `.gitattributes` (`.seen-events.jsonl merge=union`); add an idempotent ensure for
  existing stores. Also extend `list.go`'s skip-list (list.go:83-86 skips only `.metadata.json`) so
  `.gitattributes` and `.seen-events.jsonl` are not listed as knowledge files.
- 1b. New `internal/seenlog` (append + replay) replacing `.metadata.json` reads/writes in
  `seen.go`/`unseen.go`/`list.go`. `IsSeen` = replay + hash match.
- 1c. `aidb migrate-metadata` one-time converter; `git rm --cached .metadata.json`.
- 1d. TDD: concurrent-append test - two clones append different events, simulate `git pull`
  union-merge, assert no conflict and correct replay; edit-unsee test; tombstone (unseen) test.

Phase 2 - Consolidation skill (manual trigger first; prove before automating)
- 2a. Author `skills/consolidate` (4 stages above), reusing AGENTS.md categories/quality/promotion.
- 2b. Define proposal/inbox format + provenance keys (source file + hash + date + category + tier).
- 2c. Manual end-to-end run on real unseen files; assert proposals written, zero canonical writes,
  sources marked seen.

Phase 3 - Human review gate
- 3a. `aidb review` (list/accept/reject/edit; accept appends to `_aidb/`). TDD with a fixture inbox.
- 3b. `aidb encyclopedia` (categorized read of `_aidb/`). TDD with a fixture `_aidb/` tree.

Phase 4 - Automation (LAST)
- 4a. `aidb consolidate enable/disable/status` - per-machine launchd job; clone `backup.go`'s plist
  pattern (StartInterval 3600).
- 4b. Hidden `consolidate-run`: pull -> invoke the skill headless -> commit (proposals + seen-log) ->
  push-with-retry; log to `~/.aidb/consolidate.log`.
- 4c. Enforce the headless/local-only constraint; fail loudly if the skill needs absent MCP.

Phase 5 - Docs + drift safeguards
- 5a. Update `README.md` (new commands + framework framing), `AGENTS.md` (skill contract + inbox/
  promote flow), `MEMO.md` (design decisions: event-sourced ledger, propose-don't-rewrite), `TASK.md`
  (append these tasks to the roadmap), + a new skill doc.
- 5b. Encode invariants in docs + tests: cron proposes only; raw preserved; human promotes;
  provenance/timestamps; `[STALE]`; 2+-project global promotion.

## Critical files

- Ledger: `internal/metadata/metadata.go` -> new `internal/seenlog/`; `cmd/aidb/cmd/{seen,unseen,list,init}.go`.
- Staging: `cmd/aidb/cmd/{add,commit}.go` (+ new `commit_test.go`, update `add_test.go`).
- Push: `cmd/aidb/cmd/push.go`.
- Scheduler: `cmd/aidb/cmd/backup.go` (pattern to clone) -> new `cmd/aidb/cmd/consolidate.go`.
- Review: new `cmd/aidb/cmd/{review,encyclopedia}.go`.
- Skill: `~/Code/icm-runtime/skills/consolidate/` (SKILL.md + stages/).
- Docs: `README.md`, `AGENTS.md`, `MEMO.md`, `TASK.md`.

## Verification (per phase)

- `cd ~/Code/aidb && go build ./... && go test ./...` green after each phase.
- Phase 1 acceptance: two cloned stores, each `aidb seen` a different file, `git pull` on both -> no
  conflict, `aidb list --unseen` consistent on both after sync.
- Phase 2/3 acceptance: manual skill run -> proposals in inbox, `_aidb/` untouched; `aidb review
  accept` moves one snippet into `_aidb/`; `aidb encyclopedia` shows it categorized.
- Phase 4 acceptance: `aidb consolidate enable` installs the job; a forced `consolidate-run` pulls,
  proposes, commits, pushes; second machine sees the seen-events after pull (no re-analysis).

## Out of scope / non-goals

- No CRDT library (Automerge/Yjs) - overkill for single-user/multi-machine; the union-merge log is
  the right cut.
- The cron never writes canonical `_aidb/` - proposals only. This is the anti-drift invariant.
- No server/daemon; aidb stays a git-backed CLI.

## Open risks

- Headless auth: if consolidation ever needs MCP (Notion/Slack), cron will fail silently - keep it
  local-only or add an explicit precheck.
- Compaction vs union-merge: compaction rewrites the log (conflict-prone) - keep it manual/solo.
- Skill quality: extraction quality is the real risk; the human gate contains it but does not fix a
  noisy extractor - tune the quality filter in Phase 2 before automating in Phase 4.
