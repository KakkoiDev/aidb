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
4. Scheduler (Go + launchd) - `aidb consolidate enable/disable/status`, per-machine scheduled job
   (default DAILY, not hourly) that pulls -> exits if nothing is unseen -> runs the skill headless ->
   commits -> pushes (with retry).
5. Drift safeguards (cross-cutting) - cron proposes only; raw files preserved; human promotes;
   provenance + timestamps; `[STALE: date]`; 2+-project rule for global promotion.

## Keystone design - conflict-free shared seen-state

- Replace `internal/metadata/.metadata.json` with an append-only log `~/.aidb/.seen-events.jsonl`,
  one event per line:
  `{"file":"rel/path","hash":"sha256:..","commit":"<store HEAD>","seenAt":"<UTC>","machine":"<host>","op":"seen|unseen"}`.
  The `commit` field (store HEAD at append time) lets the extractor diff a re-unseen file against its
  last-seen version (`git diff <commit> -- <file>`) instead of re-consolidating the whole file.
  Without it, every edit of a living doc (MEMO.md) re-proposes ALL its old insights under a new
  source hash, the (source-file, source-hash) dedup key never collides, and the review inbox drowns.
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
  EXIT 0 if the queue is empty, before any LLM invocation (most runs) -> consolidate only the queue
  -> append `seen` events + write proposals -> `aidb commit` -> push.
- Cadence: default DAILY. Hourly was inherited from backup.go's StartInterval, but a backup is cheap
  git ops while a consolidate-run is a paid LLM session against a queue that fills a few times a day
  at most. Interval is a flag on `consolidate enable`; offset minutes per machine to shrink the
  duplicate-run window.
- Push safety: DONE (a48de8f). `pullRebase` (pull.go) and `pushWithUpstream` with one
  retry-on-rejection (push.go) exist; `backup-run` already uses them. `consolidate-run` reuses them.
- backup-run and consolidate-run share the store and may interleave commits. Every artifact must be
  valid at file granularity (one file = one proposal, below). Do NOT add a lock file - a lock in a
  git-synced store is a conflict machine.
- Same-run race (two machines consolidate the same file before either pushes): proposal filenames
  are content-addressed (id = hash of source file + source hash + category + snippet), so identical
  candidates collide into one file rather than duplicate; residual near-duplicates are deduped by
  the human at review. Union-merge keeps both `seen` events harmlessly. This handles correctness,
  not cost - both machines still burned an LLM run; the schedule offset is the cost mitigation.
- Headless caveat: cron runs with zero interactive context; the consolidation skill must be
  LOCAL-ONLY (read files, write markdown) with NO dependency on interactively-authed MCP
  (Notion/Slack/etc.), which may be absent in launchd/cron. State this as a skill constraint.

## Consolidation skill (ICM, modeled on /todo-triage)

`skills/consolidate/SKILL.md` + ordered stage contracts. Reuse `AGENTS.md` categories + quality
filter + 2+-project promotion rule (do not reinvent).
- 01-ingest: TWO queues. (a) unseen RAW files (`aidb list --unseen --json` - MEMO/TASK/REVIEW etc.)
  = candidates for the PROJECT tier; (b) unseen project `_aidb/` files (`--unseen --json --aidb`)
  = candidates for GLOBAL promotion under the 2+-project rule. (An earlier draft listed only (b),
  which consolidates the encyclopedia into itself and never mines the raw session files.)
- 02-extract: per-file, pull candidate patterns into the standard categories; apply quality filter
  (capture surprising/reusable/decision-rationale; skip obvious/one-off). For a file with a prior
  `seen` event, extract from `git diff <last-seen-commit> -- <file>` plus minimal context, not the
  whole file - this is what keeps re-proposal noise down on living documents.
- 03-propose: write candidates to `~/.aidb/_inbox/<id>.md` - flat, ONE FILE PER PROPOSAL, id =
  content hash (see the race bullet above), provenance as YAML frontmatter (source file, source
  hash, source commit, date, category, suggested tier). Skip ids present in `~/.aidb/.rejected.jsonl`
  (append-only, `merge=union` like the seen log) so rejected candidates stay rejected. NEVER write
  canonical `_aidb/`. `list` excludes `_inbox/` like other store metadata.
- 04-mark-seen: append `seen` events for processed sources.
Human-in-the-loop is between propose and canonical: the skill stops at proposals.

## Review gate (Go commands)

- `aidb review` - list pending inbox proposals with provenance; `accept <id>` appends the snippet to
  the category's `_aidb/` file (category -> file mapping fixed, from AGENTS.md), deletes the
  proposal, and commits through the normal commit path; `reject <id>` deletes the proposal AND
  appends its id to `.rejected.jsonl` - without rejection memory, the next extraction resurrects
  every rejected candidate; `edit <id>` opens `$EDITOR` before accept. Promotion to GLOBAL
  `~/.aidb/_aidb/` keeps the 2+-project rule (skill suggests, human confirms).
- `aidb encyclopedia` - categorized read view of canonical `_aidb/`. DEFERRED: it gates nothing
  (the gate is `review`), and `_aidb/` is plain markdown in a git repo where grep and the editor
  already work. Build only if real usage shows the need.

## Ordered roadmap (what to do, in what order)

Each task: write the failing test first (repo TASK.md mandates TDD; use `internal/testutil`), then
implement, then update docs. A task is done only when its regression test fails on revert.

Phase 0 - Foundation (smallest, unblocks all) - DONE 2026-06-10
- 0a. DONE (f589f78). add re-stages tracked files; commit runs `git add -u`; regression tests in
  `commit_test.go` and the rewritten `add_test.go` fail on revert.
- 0b. DONE (a48de8f). `pullRebase` (pull.go) + `pushWithUpstream` with one retry (push.go);
  `backup-run` pulls before pushing and no-ops the push without a remote. Decision taken: `add -A`
  stays - whole-store backup is the command's job, and it is what carries `.origin` pins and will
  carry the seen-log.

Phase 1 - Conflict-free shared seen-ledger (keystone)
- 1a. `init` writes `.gitattributes` (`merge=union` for `.seen-events.jsonl` AND `.rejected.jsonl`);
  add an idempotent ensure for existing stores. Also extend `list.go`'s skip-list (it now skips
  `.metadata.json` and `.origin`) with `.gitattributes`, `.seen-events.jsonl`, `.rejected.jsonl`,
  and the `_inbox/` directory.
- 1b. New `internal/seenlog` (append + replay) replacing `.metadata.json` reads/writes in
  `seen.go`/`unseen.go`/`list.go`/`remove.go` (remove mutates metadata today - easy to miss).
  `IsSeen` = replay + hash match. Events record the store HEAD in the `commit` field at append time.
- 1c. `aidb migrate-metadata` one-time converter; `git rm --cached .metadata.json` AND delete the
  file from disk - `backup-run` does `git add -A` on schedule and silently re-commits anything left
  behind.
- 1d. TDD: concurrent-append test - two clones append different events, simulate `git pull`
  union-merge, assert no conflict and correct replay; edit-unsee test; tombstone (unseen) test.

Phase 2 - Consolidation skill (manual trigger first; prove before automating)
- 2a. Author `skills/consolidate` (4 stages above), reusing AGENTS.md categories/quality/promotion.
- 2b. Define proposal/inbox format + provenance keys (source file + hash + date + category + tier).
- 2c. Manual end-to-end run on real unseen files; assert proposals written, zero canonical writes,
  sources marked seen.

Phase 3 - Human review gate
- 3a. `aidb review` (list/accept/reject/edit; accept appends to `_aidb/` and commits; reject
  records the id in `.rejected.jsonl`). TDD with a fixture inbox.
- 3b. `aidb encyclopedia` - DEFERRED (see Review gate section): grep over `_aidb/` markdown covers
  browsing; build only on demonstrated need.

Phase 4 - Automation (LAST)
- 4a. `aidb consolidate enable/disable/status` - per-machine launchd job; clone `backup.go`'s plist
  pattern (`backupPlistPath` helper exists). Default interval DAILY with a flag; per-machine minute
  offset.
- 4b. Hidden `consolidate-run`: `pullRebase` -> exit 0 if the unseen queue is empty -> invoke the
  skill headless -> commit (proposals + seen-log) -> `pushWithUpstream` (both helpers exist since
  0b); log to `~/.aidb/consolidate.log`.
- 4c. Enforce the headless/local-only constraint; fail loudly if the skill needs absent MCP.
  Acceptance must include a run under `launchctl kickstart`, not just an interactive shell - the
  LLM CLI's own auth/env (keychain, PATH, HOME) is the likely headless failure, not the skill
  content.

Phase 5 - Docs + drift safeguards
- 5a. Update `README.md` (new commands + framework framing), `AGENTS.md` (skill contract + inbox/
  promote flow), `MEMO.md` (design decisions: event-sourced ledger, propose-don't-rewrite), `TASK.md`
  (append these tasks to the roadmap), + a new skill doc.
- 5b. Encode invariants in docs + tests: cron proposes only; raw preserved; human promotes;
  provenance/timestamps; `[STALE]`; 2+-project global promotion.

## Critical files

- Ledger: `internal/metadata/metadata.go` -> new `internal/seenlog/`; `cmd/aidb/cmd/{seen,unseen,list,remove,init}.go`.
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
  local-only or add an explicit precheck. The LLM CLI itself must also be proven under launchd
  (keychain access, PATH, HOME) before trusting the schedule.
- Inbox growth: nothing expires proposals. If review lapses, the inbox accumulates until the human
  returns - acceptable for one user, but `aidb review` should print the pending count prominently
  (and `aidb status` could surface it) so a stale inbox is visible, not silent.
- Compaction vs union-merge: compaction rewrites the log (conflict-prone) - keep it manual/solo.
- Skill quality: extraction quality is the real risk; the human gate contains it but does not fix a
  noisy extractor - tune the quality filter in Phase 2 before automating in Phase 4.
