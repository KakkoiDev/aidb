---
name: aidb
description: AI knowledge database for accessing accumulated insights from past sessions. Use when starting tasks in projects with existing knowledge, looking for patterns/decisions from previous work, or before implementing something that may have been solved before.
license: MIT
metadata:
  author: KakkoiDev
  version: "1.0"
---

# aidb - AI Knowledge Database CLI

Command-line tool for managing accumulated knowledge from development sessions.

## Installation

### Binary Installation

```bash
# Go users
go install github.com/KakkoiDev/aidb/cmd/aidb@latest

# From source
git clone https://github.com/KakkoiDev/aidb && cd aidb && make install
```

### Skill Installation

```bash
# Using skills CLI (recommended)
npx skills add KakkoiDev/aidb

# Manual - Claude Code
mkdir -p ~/.claude/skills/aidb && cp SKILL.md ~/.claude/skills/aidb/
```

## When to Use

- Starting a new task in a project with existing knowledge
- Looking for patterns, decisions, or context from previous work
- Before implementing something that may have been solved before

## Context Awareness

**IMPORTANT:** aidb is context-aware based on current working directory.

- Commands operate on the **current project** (detected from git repo)
- Before running aidb commands, verify you're in the correct project directory

```bash
pwd                         # Verify: /Users/you/Code/myproject
aidb list                   # Shows files for 'myproject' only
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

## JSON Output

```bash
aidb list --unseen --json
```

Returns:
```json
[{"path":"project/branch/notes.md","seen":false,"hash":"","modified":false}]
```

## Agent Behavior

For autonomous knowledge harvesting behavior (what/when to capture), see the `aidb` agent definition in AGENTS.md or `~/.claude/agents/aidb.md`.
