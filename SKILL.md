---
name: aidb
description: AI knowledge database for accessing accumulated insights from past sessions. Use when starting tasks in projects with existing knowledge, looking for patterns/decisions from previous work, or before implementing something that may have been solved before.
license: MIT
metadata:
  author: KakkoiDev
  version: "1.0"
---

# aidb - AI Knowledge Database

AI agent skill for accessing accumulated knowledge from past sessions.

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

# Manual - GitHub Copilot
mkdir -p .github/skills/aidb && cp SKILL.md .github/skills/aidb/

# Manual - Cursor/other
mkdir -p .cursor/skills/aidb && cp SKILL.md .cursor/skills/aidb/
```

## When to Use

- Starting a new task in a project with existing knowledge
- Looking for patterns, decisions, or context from previous work
- Before implementing something that may have been solved before

## Context Awareness

**IMPORTANT:** aidb is context-aware based on current working directory.

- Commands operate on the **current project** (detected from git repo)
- Before running aidb commands, verify you're in the correct project directory
- Use `pwd` to confirm current location

```bash
# Check current project context
pwd                         # /Users/you/Code/myproject
aidb list                   # Shows files for 'myproject' only
```

## Quick Start

```bash
# Find unread knowledge files (for current project)
aidb list --unseen

# After reading a file, mark it processed
aidb seen <file>

# Re-queue a file for processing
aidb unseen <file>
```

## Workflow

1. `aidb list --unseen` - get files needing attention
2. Read relevant files based on current task context
3. `aidb seen <file>` - mark as processed
4. Apply knowledge to current task

## Knowledge Files

Two-tier knowledge system for pattern extraction:

| Tier | Location | Purpose |
|------|----------|---------|
| Project | `~/.aidb/{project}/{branch}/_aidb/` | Insights specific to that project |
| Global | `~/.aidb/_aidb/` | Patterns across all projects |

### Workflow

```bash
# Tracked files
aidb list --unseen

# Knowledge files (_aidb/)
aidb list --unseen --aidb

# Mark as processed
aidb seen project/_aidb/patterns.md
```

### File Format

```markdown
# Topic Name

What this file covers: [brief description]

---

## [Date] Entry
...
```

### Guidelines

- Max 500 lines per file
- Categorize into best-fit existing file
- Create new file only if no category fits
- lowercase-kebab-case filenames (e.g., `api-patterns.md`)

## aidb Agent

A standalone Claude Code subagent for full lifecycle knowledge management.

### Agent Installation

```bash
# Copy to Claude Code agents directory
mkdir -p ~/.claude/agents
curl -o ~/.claude/agents/aidb.md https://raw.githubusercontent.com/KakkoiDev/aidb/master/AGENTS.md

# Or from local clone
cp AGENTS.md ~/.claude/agents/aidb.md
```

### Agent Workflow

1. **Verify project context**: `pwd` to confirm current directory
2. Check `aidb list --unseen` for current project files
3. Read and categorize insights into project `_aidb/`
4. Check `aidb list --unseen --aidb` for synthesis candidates
5. Extract patterns to global `~/.aidb/_aidb/`
6. Mark processed: `aidb seen <file>`
7. Commit: `aidb commit "message"`

## Path Structure

Files are stored as: `~/.aidb/<project>/<branch>/<file>.md`

Example: `~/.aidb/myproject/main/notes.md`

## JSON Output

```bash
aidb list --unseen --json
```

Returns:
```json
[{"path":"project/branch/notes.md","seen":false,"hash":"","modified":false}]
```
