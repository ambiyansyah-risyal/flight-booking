# Scripts Directory

This directory contains various utility scripts organized by function:

## Directories

### `hooks/`
Git hooks for local development:
- `install-hooks.sh` - Installs all git hooks into your local repository
- `prevent-co-authored.sh` - Git prepare-commit-msg hook to prevent co-authored-by lines and enforce commit message conventions

### `verify/`
Verification and testing scripts:
- `compose-verify.sh` - Docker Compose-based verification workflow
- `local-verify.sh` - Local verification without Docker Compose

### `ci/`
CI/CD and pre-commit scripts:
- `pre-commit-check.sh` - Pre-commit checks (linting, tests)
- `run-act-check.sh` - GitHub Actions validation using `act`