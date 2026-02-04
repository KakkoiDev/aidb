# aidb

AI Knowledge Database - centralized knowledge file management for AI agent workflows.

## Installation

```bash
git clone https://github.com/KakkoiDev/aidb && cd aidb
sudo make install              # Install to /usr/local/bin
# or
PREFIX=~/.local make install   # Install to ~/.local/bin
```

## Setup

```bash
aidb init                                      # Initialize ~/.aidb
aidb init --remote git@github.com:user/kb.git  # With remote for sync
```

## Usage

### Core Commands

```bash
aidb add TASK.md         # Move file to ~/.aidb, create symlink, stage in git
aidb commit "Add notes"  # Commit staged changes
aidb status              # Show staged/unstaged changes
aidb push                # Push commits to remote
aidb pull                # Pull changes from remote
aidb remove TASK.md      # Untrack file, restore to original location
```

### AI Agent Commands

```bash
aidb list                # List tracked files with metadata
aidb list --unseen       # Show files needing AI processing
aidb seen TASK.md        # Mark file as processed by AI
aidb unseen TASK.md      # Mark file for re-processing
```

### Configuration

```bash
aidb config              # Show configuration
aidb backup enable       # Enable automatic backup (hourly commit + push)
```

## How It Works

- Files are stored in `~/.aidb/{repo-name}/{branch}/{filename}`
- Symlinks are created at original locations
- Built-in git versioning provides disaster recovery
- Metadata tracks AI processing state

## Uninstall

```bash
sudo make uninstall
```
