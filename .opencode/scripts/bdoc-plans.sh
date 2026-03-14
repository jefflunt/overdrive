#!/bin/bash

PLAN_DIR="build_docs/plans"

if [ ! -d "$PLAN_DIR" ] || [ -z "$(ls -A "$PLAN_DIR")" ]; then
    echo "No plans found in $PLAN_DIR"
    exit 0
fi

# Detect output mode (Terminal vs Pipe/Context)
if [ -t 1 ]; then
    MODE="ansi"
    # Colors
    BLUE='\033[1;34m'
    YELLOW='\033[1;33m'
    GREEN='\033[1;32m'
    RED='\033[1;31m'
    RESET='\033[0m'
    
    # Header
    printf "%-15s %-10s %-30s %-s\n" "STATUS" "TYPE" "FILENAME" "TITLE"
    printf "%-15s %-10s %-30s %-s\n" "------" "----" "--------" "-----"
else
    MODE="markdown"
    # Header
    echo "| STATUS | TYPE | FILENAME | TITLE |"
    echo "|:---|:---|:---|:---|"
fi

# Temporary file for sorting
TMP_FILE=$(mktemp)

for f in "$PLAN_DIR"/*.md; do
    [ -e "$f" ] || continue
    
    # Extract metadata
    title=$(sed -n 's/^title: *//p' "$f" | head -1)
    status=$(sed -n 's/^status: *//p' "$f" | head -1)
    type=$(sed -n 's/^type: *//p' "$f" | head -1)
    filename=$(basename "$f")
    
    # Assign sort weight
    case "$status" in
        todo) weight=1 ;;
        in-progress|in-development) weight=2 ;;
        done) weight=3 ;;
        *) weight=4 ;;
    esac
    
    echo "$weight|$status|$type|$filename|$title" >> "$TMP_FILE"
done

# Sort and print
sort -t'|' -k1,1n "$TMP_FILE" | while IFS='|' read -r weight status type filename title; do
    if [ "$MODE" = "ansi" ]; then
        # Color status
        case "$status" in
            todo) s_color=$BLUE ;;
            in-progress|in-development) s_color=$YELLOW ;;
            done) s_color=$GREEN ;;
            *) s_color=$RESET ;;
        esac
        
        # Color type
        case "$type" in
            feature) t_color=$BLUE ;;
            bug) t_color=$RED ;;
            *) t_color=$RESET ;;
        esac
        
        printf "${s_color}%-15s${RESET} ${t_color}%-10s${RESET} %-30s %-s\n" "$status" "$type" "$filename" "$title"
    else
        # Markdown Icons
        case "$status" in
            todo) s_icon="🔴" ;;
            in-progress|in-development) s_icon="🟡" ;;
            done) s_icon="🟢" ;;
            *) s_icon="⚪" ;;
        esac
        
        # Output Markdown table row
        printf "| %s %s | %s | %s | %s |\n" "$s_icon" "$status" "$type" "$filename" "$title"
    fi
done

rm "$TMP_FILE"
