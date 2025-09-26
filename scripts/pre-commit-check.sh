#!/bin/bash
# Pre-commit check script to run act and check commit message format

echo "Running pre-commit checks..."

# Run golangci-lint to check for issues
echo "Running golangci-lint..."
if ! golangci-lint run; then
    echo "❌ golangci-lint found issues. Please fix them before committing."
    exit 1
fi

# Run tests to ensure they pass
echo "Running tests..."
if ! go test ./... -race; then
    echo "❌ Tests failed. Please fix them before committing."
    exit 1
fi

echo "✅ Pre-commit checks passed!"
exit 0