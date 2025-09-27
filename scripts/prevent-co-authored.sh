#!/bin/bash
# Script to remove co-authored-by lines from commit message and enforce single-line commit messages

COMMIT_MSG_FILE=$1

# Remove any co-authored-by lines from the commit message
if grep -q "Co-authored-by:" "$COMMIT_MSG_FILE"; then
    echo "Removing co-authored-by lines from commit message..."
    sed -i '/^Co-authored-by:/d' "$COMMIT_MSG_FILE"
fi

# Count the number of lines in the commit message file and enforce single-line commit message
LINE_COUNT=$(wc -l < "$COMMIT_MSG_FILE" | tr -d ' ')
if [ "$LINE_COUNT" -gt 1 ]; then
    # Extract the first line and remove empty lines
    FIRST_LINE=$(head -n1 "$COMMIT_MSG_FILE" | sed '/^[[:space:]]*$/d')
    
    # If first line is empty (only whitespace), show error and exit
    if [ -z "$FIRST_LINE" ]; then
        echo "Error: Commit message cannot be empty or only whitespace."
        echo "Commit messages should be a single line following the project conventions."
        exit 1
    fi
    
    # Write only the first line back to the commit message file
    echo "$FIRST_LINE" > "$COMMIT_MSG_FILE"
    
    echo "Warning: Commit message was multi-line. Only the first line was preserved."
    echo "Following project convention of single-line commit messages."
fi