#!/bin/bash

# Define the plans directory
PLAN_DIR="build_docs/plans"

# Check if the directory exists
if [ ! -d "$PLAN_DIR" ]; then
    echo "No plans directory found at $PLAN_DIR."
    exit 0
fi

# Counter for deleted files
count=0

# Find files matching *.md in PLAN_DIR that were modified more than 30 days ago
# Use a while loop to process each file safely
find "$PLAN_DIR" -name "*.md" -mtime +30 | while read -r file; do
    # Check if the file contains "status: done"
    # We use grep to search for the pattern.
    if grep -q "^status: *done" "$file"; then
        echo "Deleting old completed plan: $file"
        rm "$file"
        ((count++))
    fi
done

echo "Cleanup complete."
