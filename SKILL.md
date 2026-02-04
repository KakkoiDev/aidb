# aidb - AI Knowledge Database

AI agent skill for accessing accumulated knowledge from past sessions.

## When to Use

- Starting a new task in a project with existing knowledge
- Looking for patterns, decisions, or context from previous work
- Before implementing something that may have been solved before

## Quick Start

```bash
# Find unread knowledge files
aidb list --unseen

# After reading a file, mark it processed
aidb seen <file>

# Re-queue a file for processing
aidb unseen <file>
```

## File Types

| File | Purpose |
|------|---------|
| MEMO.md | Codebase analysis, architecture notes |
| TASK.md | Implementation plans, progress tracking |
| LEARN.md | Key insights, patterns, decisions |
| COACH.md | Development approach reflections |

## Workflow

1. `aidb list --unseen` - get files needing attention
2. Read relevant files based on current task context
3. `aidb seen <file>` - mark as processed
4. Apply knowledge to current task

## Path Structure

Files are stored as: `~/.aidb/<project>/<branch>/<file>.md`

Example: `~/.aidb/meetsone/master/LEARN.md`

## JSON Output

```bash
aidb list --unseen --json
```

Returns:
```json
[{"path":"project/branch/MEMO.md","seen":false,"hash":"","modified":false}]
```
