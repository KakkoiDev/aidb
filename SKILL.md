---
name: aidb
description: AI knowledge database for accessing accumulated insights from past sessions. Use when starting tasks in projects with existing knowledge, looking for patterns/decisions from previous work, or before implementing something that may have been solved before.
license: MIT
metadata:
  author: KakkoiDev
  version: "1.0"
---

# aidb - AI Knowledge Database CLI

Context-aware tool operating on the current git project.

## Installation

### Binary

```bash
# Go users
go install github.com/KakkoiDev/aidb/cmd/aidb@latest

# From source
git clone https://github.com/KakkoiDev/aidb && cd aidb && make install
```

### Skill

```bash
# Using skills CLI (recommended)
npx skills add KakkoiDev/aidb

# Manual - Claude Code
mkdir -p ~/.claude/skills/aidb && cp SKILL.md ~/.claude/skills/aidb/
```

## Commands

| Command | Description |
|---------|-------------|
| `aidb init` | Initialize ~/.aidb |
| `aidb add <file>` | Track file (move to ~/.aidb, create symlink) |
| `aidb remove <file>` | Untrack file (restore original) |
| `aidb list` | List tracked files |
| `aidb list --unseen` | Files needing attention |
| `aidb list --aidb` | Knowledge files only (_aidb/) |
| `aidb seen <file>` | Mark as processed |
| `aidb unseen <file>` | Re-queue for processing |
| `aidb commit "msg"` | Commit changes |
| `aidb push` | Push to remote |
| `aidb pull` | Pull from remote |

## Path Structure

Files stored as: `~/.aidb/<project>/<branch>/<file>.md`

Two-tier knowledge system:
- Project: `~/.aidb/{project}/{branch}/_aidb/`
- Global: `~/.aidb/_aidb/` (promoted patterns only)

## Agent Behavior

For harvesting logic (what/when to capture), see `~/.claude/agents/aidb.md`.
