#!/bin/bash
# Script to install Git hooks for the project

set -e

HOOKS_DIR="$(git rev-parse --show-toplevel)/.git/hooks"
SCRIPTS_DIR="$(git rev-parse --show-toplevel)/scripts"

echo "Installing Git hooks..."

# Create the pre-commit hook with proper git root path
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Git pre-commit hook to run checks before committing

# Get the root directory of the git repository
REPO_ROOT="$(git rev-parse --show-toplevel)"
SCRIPTS_DIR="$REPO_ROOT/scripts"

echo "Running pre-commit checks..."

# Call the script to run pre-commit checks
"$SCRIPTS_DIR/pre-commit-check.sh"

if [ $? -ne 0 ]; then
    echo "❌ Pre-commit checks failed. Commit aborted."
    exit 1
fi

# Optionally run act to validate GitHub Actions (only syntax check to keep it fast)
"$SCRIPTS_DIR/run-act-check.sh"

if [ $? -ne 0 ]; then
    echo "❌ GitHub Actions validation failed. Commit aborted."
    exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0
EOF

# Create the prepare-commit-msg hook with proper git root path
cat > "$HOOKS_DIR/prepare-commit-msg" << 'EOF'
#!/bin/bash
# Git hook to prevent co-authored-by lines in commit messages

# Get the root directory of the git repository
REPO_ROOT="$(git rev-parse --show-toplevel)"
SCRIPTS_DIR="$REPO_ROOT/scripts"

# Call the script to prevent co-authored-by lines
"$SCRIPTS_DIR/prevent-co-authored.sh" $1 $2 $3
EOF

chmod +x "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/prepare-commit-msg"

echo "✅ Git hooks installed successfully!"
echo "The following hooks were installed:"
echo "  - pre-commit: Runs linting and tests before each commit"
echo "  - prepare-commit-msg: Prevents co-authored-by lines from being added to commit messages"