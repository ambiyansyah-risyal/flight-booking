#!/bin/bash
# Script to run act for CI/CD validation

echo "Checking if act is available..."
if ! command -v act &> /dev/null; then
    echo "⚠️  act is not installed. Skipping CI/CD check."
    exit 0
fi

echo "Running act to validate GitHub Actions workflow..."
if ! act -n; then  # Use dry-run first to check syntax
    echo "⚠️  Dry run of act failed. Please check your workflow configuration."
    exit 1
fi

echo "✅ GitHub Actions syntax check passed."
# Optionally run the actual workflow - commented out to avoid long execution during commit
# if ! act; then
#     echo "❌ GitHub Actions workflow failed. Please fix issues before committing."
#     exit 1
# fi

exit 0