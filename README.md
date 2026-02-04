# aidb

Centralized file management with git versioning.

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
aidb init --remote git@github.com:user/kb.git  # With remote
```

## Usage

```bash
aidb add <file>          # Track file (move to ~/.aidb, create symlink)
aidb remove <file>       # Untrack file (restore to original location)
aidb status              # Show changes
aidb commit "message"    # Commit changes
aidb push                # Push to remote
aidb pull                # Pull from remote
```

## How It Works

- Files stored in `~/.aidb/{repo}/{branch}/{filename}`
- Symlinks created at original locations
- Git versioning for history and sync

## Uninstall

```bash
sudo make uninstall
```
