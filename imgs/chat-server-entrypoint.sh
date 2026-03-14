#!/bin/bash
source /etc/profile
set -e

# Setup logging
mkdir -p /log
# Just redirect to log file simply
exec > >(tee -a /log/chat.log) 2>&1

# Ensure required environment variables are present
if [ -z "$REPO_URL" ]; then echo "Error: REPO_URL is not set"; exit 1; fi
if [ -z "$PRIMARY_BRANCH" ]; then echo "Error: PRIMARY_BRANCH is not set"; exit 1; fi

echo "Prepping $REPO_URL ..."

# Setup authentication
mkdir -p ~/.local/share/opencode
if [ -n "$AUTH_JSON" ]; then
    echo "$AUTH_JSON" > ~/.local/share/opencode/auth.json
fi

# Switch to the target folder and then switch to the git branch PRIMARY_BRANCH
export GIT_SSH_COMMAND="ssh -i /ssh.key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
if [ ! -d "target" ]; then
    echo "  Cloning $REPO_URL (branch: $PRIMARY_BRANCH) into target..."
    git clone --depth 1 --branch "$PRIMARY_BRANCH" "$REPO_URL" target
fi
cd target

# Ensure we are on the right branch (should already be true from clone)
echo "  Ensuring we are on $PRIMARY_BRANCH..."
git checkout "$PRIMARY_BRANCH" || git checkout -b "$PRIMARY_BRANCH"

echo "Running opencode serve for chat"
echo "  MODEL  $MODEL"
echo "  SESSION $OPENCODE_SESSION_ID"

SESSION_FLAG=""
if [ -n "$OPENCODE_SESSION_ID" ]; then
    SESSION_FLAG="--session $OPENCODE_SESSION_ID"
fi

# Start opencode serve on port 3000
exec opencode serve --port 3000 --hostname 0.0.0.0 $SESSION_FLAG
