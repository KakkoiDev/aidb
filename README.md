# aidb

Centralized knowledge for AI agents. Track insights across projects. Never lose context again.

## Quick Start

```bash
# Install
go install github.com/KakkoiDev/aidb/cmd/aidb@latest

# Initialize
aidb init
aidb init --remote git@github.com:user/kb.git  # With remote sync
```

## Workflows

**Starting a new task** → Check for existing knowledge
```bash
aidb list --unseen          # Find unread files
aidb read MEMO.md           # Read relevant context
```

**After learning something** → Mark as processed
```bash
aidb seen <file>            # Mark file as seen
aidb unseen <file>          # Re-queue for attention
```

**Tracking files** → Add to knowledge base
```bash
aidb add MEMO.md            # Track file (creates symlink)
aidb remove MEMO.md         # Untrack file (restores original)
```

**Syncing knowledge** → Git versioning
```bash
aidb status                 # Show changes
aidb commit "message"       # Commit changes
aidb push                   # Push to remote
aidb pull                   # Pull from remote
```

## Commands

| Command | Description |
|---------|-------------|
| `aidb init` | Initialize ~/.aidb |
| `aidb add <file>` | Track file (move to ~/.aidb, create symlink) |
| `aidb remove <file>` | Untrack file (restore to original location) |
| `aidb list` | List tracked files |
| `aidb list --unseen` | Show files needing attention |
| `aidb seen <file>` | Mark file as processed |
| `aidb unseen <file>` | Re-queue file for processing |
| `aidb status` | Show git status |
| `aidb commit "msg"` | Commit changes |
| `aidb push` | Push to remote |
| `aidb pull` | Pull from remote |

## File Types

| File | Purpose |
|------|---------|
| MEMO.md | Codebase analysis, architecture notes |
| TASK.md | Implementation plans, progress tracking |
| LEARN.md | Key insights, patterns, decisions |
| COACH.md | Development approach reflections |

## How It Works

- Files stored in `~/.aidb/{repo}/{branch}/{filename}`
- Symlinks created at original locations
- Git versioning for history and sync
- Seen/unseen tracking with automatic change detection (modified files become unseen)

## Configuration

```bash
# Initialize with remote
aidb init --remote git@github.com:user/kb.git

# Configure remote later
cd ~/.aidb && git remote add origin <url>
```

<details>
<summary>Custom installation path</summary>

```bash
PREFIX=~/.local make install   # Install to ~/.local/bin
```
</details>

## Requirements

- Go 1.21+
- Git

## Installation

**Go install (recommended)**
```bash
go install github.com/KakkoiDev/aidb/cmd/aidb@latest
```

**From source**
```bash
git clone https://github.com/KakkoiDev/aidb && cd aidb
sudo make install              # Install to /usr/local/bin
```

**Uninstall**
```bash
sudo make uninstall
```

## License

MIT
