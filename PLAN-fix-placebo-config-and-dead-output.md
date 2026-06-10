# Plan: fix - placebo config keys, dead output package, dead global flags

Handoff plan for an AI agent. Self-contained. Go project (cobra CLI). Theme: surfaces that
claim to work but are wired to nothing. Verified by grep over the whole module.

## Symptoms

1. **`aidb config db.path <value>` is a placebo.** It saves to `~/.config/aidb/config.yaml` and
   prints success (`cmd/aidb/cmd/config.go:123-124,141`), but `config.New()` hardcodes
   `~/.aidb` (`internal/config/config.go:17-27`) and `UserConfig.DB.Path` is read NOWHERE
   (grep: written at config.go:124 only). Every command ignores the setting the tool confirmed.
2. **`aidb config backup.enabled` is a placebo.** Written and echoed back
   (`config.go:96,111,126`), consumed by nothing - backup enable/disable manage the launchd
   plist directly (`cmd/aidb/cmd/backup.go:48-145`). Two sources of truth, one fake.
3. **`internal/output` is dead code.** The package (TTY detection, writer injection, NO_COLOR
   building block - `internal/output/output.go`) is imported by zero files. Commands use
   duplicated helpers in `cmd/aidb/cmd/root.go:58-108` which do NOT detect TTY, so piped/captured
   output contains raw ANSI escapes (observed in practice: `[0;31m...` in captured tool output).
   `status.go:78-92` carries a THIRD color implementation.
4. **Global `--json` flag is dead.** Declared `root.go:51`, `flagJSON` read nowhere (grep). Only
   `list` implements JSON via its own local flag. Help text claims clig.dev compliance.

## Fix

1. `db.path`: wire it. `config.New()` loads `~/.config/aidb/config.yaml` (reuse
   `loadUserConfig`, move it into `internal/config` to avoid an import cycle) and uses `DB.Path`
   when non-empty, else `~/.aidb`. CAUTION: tests redirect the store via `$HOME` (see
   `internal/testutil`); keep `os.UserHomeDir()` as the default base so the test scaffolding
   keeps working, and add precedence: explicit config > default. Update `aidb config db.path`
   help to note existing symlinks are NOT migrated (document, do not migrate).
2. `backup.enabled`: delete the config key (reject with "managed by aidb backup enable/disable").
   `aidb config` (show-all) reports actual state instead: plist-exists check, same logic as
   `backupStatus` (`backup.go:158-171`). One source of truth.
3. Output: adopt `internal/output` in `cmd/` - replace root.go print helpers and status.go color
   funcs with one package-level `out *output.Output` initialized from the global flags in a
   `PersistentPreRun`. TTY detection then fixes piped-ANSI for free. Honor `NO_COLOR` env in
   `colorEnabled()` (one line, clig.dev). Delete the dead duplicates.
4. Global `--json`: remove the dead global flag; keep working per-command flags (`list --json`).
   Removing beats half-implementing JSON for status/config nobody asked for.

## Acceptance criteria

- [ ] `aidb config db.path /tmp/customdb` then `aidb add x.md` stores under /tmp/customdb.
- [ ] Default behavior with no config file is byte-identical to today (~/.aidb).
- [ ] `aidb config backup.enabled true` errors with guidance; show-all reports real plist state.
- [ ] `aidb status | cat` contains zero ANSI escape sequences; interactive TTY output still colored.
- [ ] `NO_COLOR=1 aidb status` uncolored even on a TTY.
- [ ] `aidb status --json` errors as unknown flag (global removed); `aidb list --json` unchanged.
- [ ] `internal/output` has at least one importer or the duplicates are gone (no third copy left).

## Tests (fail-on-revert required)

- `internal/config`: db.path precedence test (config file set -> DBDir honors it; unset ->
  $HOME/.aidb). Must FAIL today.
- Output: unit test on `internal/output` asserting no ANSI when Writer is a bytes.Buffer
  (non-TTY), plus NO_COLOR.
- Extend `cmd/aidb/cmd/config_test.go` (exists): backup.enabled rejection; db.path round-trip.

## Out of scope (do NOT bundle)

- Implementing JSON output for more commands.
- Any quiet-mode rework beyond what moving to internal/output gives for free.

## Verify

```
cd ~/Code/aidb && go build ./... && go test ./...
./aidb status | cat -v | grep -c "\^\[" # expect 0
```
