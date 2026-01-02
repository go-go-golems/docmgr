# Source

Copied from /home/manuel/.cursor/commands/git-commit-instructions.md

# Git Commit Instructions

## Before Every Commit

```bash
# 1. Check what's changed
git status --porcelain

# 2. Review staged vs unstaged
git diff --stat          # unstaged changes
git diff --cached --stat # staged changes
```

## Adding Files

```bash
# Preferred: add specific files/dirs
git add path/to/file.go path/to/other.go

# Add all changes (use carefully)
git add -A
```

## Never Commit (Common Noise)

| Pattern | Why |
|---------|-----|
| `node_modules/` | npm deps |
| `vendor/` | go vendor (usually) |
| `*.exe`, `*.bin`, binaries | build artifacts |
| `.env`, `.env.local` | secrets |
| `*.log` | logs |
| `dist/`, `build/`, `out/` | build output |
| `.DS_Store`, `Thumbs.db` | OS junk |
| `*.pyc`, `__pycache__/` | python cache |
| `*.o`, `*.a` | compiled objects |
| `.idea/`, `.vscode/` | IDE config (unless shared) |
| `coverage/`, `*.cover` | test coverage |
| `tmp/`, `temp/` | temp files |

## Quick Checks

```bash
# Verify .gitignore is working
git status --ignored

# Check if a file would be ignored
git check-ignore -v path/to/file

# See what's about to be committed
git diff --cached --name-only
```

## Commit Pattern

```bash
# Stage specific files
git add file1.go file2.go dir/

# Commit with message
git commit -m "Short summary of change"

# Get commit hash (for diary)
git rev-parse HEAD
```

## If You Accidentally Staged Noise

```bash
# Unstage specific file
git reset HEAD path/to/file

# Unstage everything
git reset HEAD
```

## If Noise Got Committed

```bash
# Remove from git but keep on disk
git rm --cached path/to/noise

# Add to .gitignore
echo "path/to/noise" >> .gitignore

# Amend if it was the last commit
git add .gitignore
git commit --amend --no-edit
```
