# aidb - AI Knowledge Database

Centralized file management with git versioning. Track files in ~/.aidb with symlinks back to original locations.

## Commands

- `aidb init` - Initialize database
- `aidb add <file>` - Track file (move to ~/.aidb, create symlink)
- `aidb remove <file>` - Untrack file (restore original)
- `aidb list --unseen` - Files needing attention
- `aidb seen <file>` - Mark as processed
- `aidb commit "msg"` - Commit changes
- `aidb push/pull` - Sync with remote

## Two-Tier Knowledge System

- `{project}/_aidb/` - Project-specific knowledge
- `~/.aidb/_aidb/` - Global patterns

Use `aidb list --aidb` to list knowledge files.
