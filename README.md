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
| `aidb list` | List tracked files (excludes _aidb/) |
| `aidb list --unseen` | Show files needing attention |
| `aidb list --aidb` | Show only _aidb/ knowledge files |
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

## Knowledge Harvesting

Two-tier knowledge system for pattern extraction:

| Tier | Location | Purpose |
|------|----------|---------|
| Project | `{project}/_aidb/` | Insights specific to that project |
| Global | `~/.aidb/_aidb/` | Patterns across all projects |

```bash
# Regular files (MEMO, TASK, LEARN)
aidb list --unseen

# Knowledge files (_aidb/)
aidb list --unseen --aidb

# Mark as processed
aidb seen project/_aidb/patterns.md
```

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

## Usage Modes

All components are opt-in. Mix and match based on your needs:

### CLI Only

Just the `aidb` tool for manual knowledge management.

```bash
go install github.com/KakkoiDev/aidb/cmd/aidb@latest
aidb init
```

Use cases:
- Manual tracking of MEMO.md, TASK.md, LEARN.md
- Git-backed knowledge versioning
- Seen/unseen workflow for processing files

### Skill + CLI

Add the skill for knowledge-aware AI prompting.

```bash
# Install skill
mkdir -p ~/.claude/skills/aidb && cp SKILL.md ~/.claude/skills/aidb/
```

Use cases:
- AI agents can query knowledge with `/aidb`
- Context discovery at session start
- Pattern lookup before implementation

### Agent + Skill + CLI

Full automation with the harvesting agent.

```bash
# Install agent
mkdir -p ~/.claude/agents && cp agents/aidb.md ~/.claude/agents/

# Or fetch directly
curl -o ~/.claude/agents/aidb.md https://raw.githubusercontent.com/KakkoiDev/aidb/master/agents/aidb.md
```

Use cases:
- Automatic knowledge extraction after tasks
- Two-tier pattern synthesis (project → global)
- Hands-off knowledge base maintenance

## License

MIT
