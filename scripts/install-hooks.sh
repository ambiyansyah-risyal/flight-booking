#!/bin/bash
# Script to install Git hooks for the project

set -e

HOOKS_DIR="$(git rev-parse --show-toplevel)/.git/hooks"
SCRIPTS_DIR="$(git rev-parse --show-toplevel)/scripts"

echo "Installing Git hooks..."

# Copy the hook scripts to the .git/hooks directory
cp "$SCRIPTS_DIR/pre-commit-check.sh" "$HOOKS_DIR/pre-commit"
cp "$SCRIPTS_DIR/prevent-co-authored.sh" "$HOOKS_DIR/prepare-commit-msg" 

# Make the hooks executable
chmod +x "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/prepare-commit-msg"

# Since the original hooks have different names, let me create proper hook content
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Git pre-commit hook to run checks before committing

SCRIPTS_DIR="$(dirname "$0")/../scripts"

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

cat > "$HOOKS_DIR/prepare-commit-msg" << 'EOF'
#!/bin/bash
# Git hook to prevent co-authored-by lines in commit messages

SCRIPTS_DIR="$(dirname "$0")/../scripts"

# Call the script to prevent co-authored-by lines
"$SCRIPTS_DIR/prevent-co-authored.sh" $1 $2 $3
EOF

chmod +x "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/prepare-commit-msg"

echo "✅ Git hooks installed successfully!"
echo "The following hooks were installed:"
echo "  - pre-commit: Runs linting and tests before each commit"
echo "  - prepare-commit-msg: Prevents co-authored-by lines from being added to commit messages"