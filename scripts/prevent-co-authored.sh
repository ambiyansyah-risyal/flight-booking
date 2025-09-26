#!/bin/bash
# Script to remove co-authored-by lines from commit message

COMMIT_MSG_FILE=$1

# Remove any co-authored-by lines from the commit message
if grep -q "Co-authored-by:" "$COMMIT_MSG_FILE"; then
    echo "Removing co-authored-by lines from commit message..."
    sed -i '/^Co-authored-by:/d' "$COMMIT_MSG_FILE"
fi