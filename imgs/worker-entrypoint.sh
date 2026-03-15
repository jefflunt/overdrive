#!/bin/bash
source /etc/profile
set -e

# Setup logging with timestamps
mkdir -p /log
exec > >(while IFS= read -r line || [ -n "$line" ]; do
    if [[ ! "$line" =~ ^\[[0-9]{4}-[0-9]{2}-[0-9]{2} ]]; then
        printf "[%(%Y-%m-%d %H:%M:%S)T] %s\n" -1 "$line"
    else
        printf "%s\n" "$line"
    fi
done) 2>&1

function set_sub_status() {
    echo "$1" > /log/sub_status
}

function commit_and_push() {
    local phase="$1"
    echo "  Checking for changes to commit ($phase)..."
    git add --all
    if ! git diff --cached --quiet; then
        local current_msg="$COMMIT_MSG"
        if [ -n "$phase" ]; then
            current_msg="$COMMIT_MSG ($phase)"
        fi
        echo "  Committing and pushing changes for $phase..."
        git commit -m "$current_msg"
        git push -u origin "$WORKER_BRANCH"
    else
        echo "  No changes detected in $phase phase."
    fi
    
    # Always update these so the UI reflects the current state
    git rev-parse HEAD > /log/related_commit
    if [ -n "$INITIAL_SHA" ]; then
        git diff --color=always $INITIAL_SHA HEAD > /log/diff
    fi
}

function run_tests() {
    echo "Auto-detecting and running tests..."
    set_sub_status "testing"
    
    local test_cmd=""
    if [ -f "go.mod" ]; then
        test_cmd="go test ./..."
    elif [ -f "package.json" ]; then
        if grep -q "\"test\":" package.json; then
            test_cmd="npm test"
        fi
    elif [ -f "requirements.txt" ] || [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
        if command -v pytest >/dev/null 2>&1; then
            test_cmd="pytest"
        elif command -v python3 >/dev/null 2>&1; then
            test_cmd="python3 -m unittest discover"
        fi
    elif [ -f "Cargo.toml" ]; then
        test_cmd="cargo test"
    elif [ -f "Makefile" ]; then
        if grep -q "^test:" Makefile; then
            test_cmd="make test"
        fi
    fi

    if [ -z "$test_cmd" ]; then
        echo "No test suite detected."
        return 0
    fi

    echo "Running tests: $test_cmd"
    local test_output_file="/log/test_output"
    set +e
    $test_cmd > "$test_output_file" 2>&1
    local exit_code=$?
    set -e

    if [ $exit_code -eq 0 ]; then
        echo "passed" > /log/test_status
        echo "Tests passed!"
    else
        echo "failed" > /log/test_status
        echo "Tests failed with exit code $exit_code."
        cat "$test_output_file"
    fi
    return $exit_code
}

if [ -n "$CUSTOM_CMD" ]; then
    echo "Running custom command: $CUSTOM_CMD"
    set_sub_status "executing"
    cd target
    eval "$CUSTOM_CMD"
    exit $?
fi

# Ensure required environment variables are present
if [ -z "$REPO_URL" ]; then echo "Error: REPO_URL is not set"; exit 1; fi
if [ -z "$WORKER_BRANCH" ]; then echo "Error: WORKER_BRANCH is not set"; exit 1; fi
if [ -z "$COMMIT_MSG" ]; then echo "Error: COMMIT_MSG is not set"; exit 1; fi
if [ -z "$PROMPT" ]; then echo "Error: PROMPT is not set"; exit 1; fi

set_sub_status "initializing"

# Setup authentication
mkdir -p ~/.local/share/opencode
if [ -n "$AUTH_JSON" ]; then
    echo "$AUTH_JSON" > ~/.local/share/opencode/auth.json
fi

# Configure git user
git config --global user.email "overdrive@txnal.com"
git config --global user.name "overdrive by transactional"

echo "Prepping $REPO_URL ..."

# Setup SSH command for bypass host key verification
export GIT_SSH_COMMAND="ssh -i /ssh.key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# Clone the repo mentioned in REPO_URL into the folder WORKDIR/target/
# Current directory is WORKDIR (/overdrive)
echo "  Cloning $REPO_URL into target..."

set_sub_status "getting latest code"
git clone --depth 1 "$REPO_URL" target
cd target

# Switch to or create the WORKER_BRANCH
if git ls-remote --exit-code --heads origin "$WORKER_BRANCH" >/dev/null 2>&1; then
    echo "  Branch $WORKER_BRANCH exists on remote. Fetching..."
    git fetch --depth 1 origin "$WORKER_BRANCH"
    git checkout "$WORKER_BRANCH" 2>/dev/null || git checkout -b "$WORKER_BRANCH" FETCH_HEAD
    git pull origin "$WORKER_BRANCH"
else
    echo "  Branch $WORKER_BRANCH does not exist on remote. Creating locally..."
    git checkout "$WORKER_BRANCH" 2>/dev/null || git checkout -b "$WORKER_BRANCH"
fi

INITIAL_SHA=$(git rev-parse HEAD)

# Resolve Job ID to SHA for revert if necessary
if [[ "$PROMPT" == "/bdoc-revert "* ]]; then
    CMD=$(echo "$PROMPT" | cut -d' ' -f1)
    TARGET=$(echo "$PROMPT" | cut -d' ' -f2)
    
    # Try to resolve to a SHA. If it's a Job ID, it might be in the commit message.
    COMMIT_SHA=$(git rev-parse "$TARGET" 2>/dev/null || git log --grep="$TARGET" -n 1 --format=%H || echo "")
    
    if [ -n "$COMMIT_SHA" ] && [ "$COMMIT_SHA" != "$TARGET" ]; then
        echo "  Resolved $TARGET to $COMMIT_SHA for $CMD"
        PROMPT="$CMD $COMMIT_SHA"
    fi
fi

# Working directly on WORKER_BRANCH
echo "  Working on branch $WORKER_BRANCH..."

# Use the first 70 characters of the prompt as the commit message for /bdoc-engineer jobs
if [[ "$PROMPT" == "/bdoc-engineer "* ]]; then
    CLEAN_PROMPT="${PROMPT#/bdoc-engineer }"
    # Skip # title header if present
    if [[ "$CLEAN_PROMPT" == "# title"* ]]; then
        CLEAN_PROMPT="${CLEAN_PROMPT#\# title}"
    fi
    # Trim leading whitespace/newlines
    CLEAN_PROMPT="$(echo "$CLEAN_PROMPT" | xargs | sed 's/^[[:space:]]*//')"
    
    NEW_COMMIT_MSG="${CLEAN_PROMPT:0:70}"
    if [ -n "$NEW_COMMIT_MSG" ]; then
        # Escape quotes for reliability
        COMMIT_MSG=$(printf "%s" "$NEW_COMMIT_MSG" | sed 's/"/\\"/g; s/'"'"'/\\'"'"'/g')
    fi
fi

echo "Running opencode v$(opencode --version) for work"
# Assuming opencode is in the PATH
echo "  MODEL  $MODEL"
echo "  PROMPT $PROMPT\n"

echo "-= AGENT start =-"
if [[ "$PROMPT" == "/bdoc-idea "* ]]; then
    set_sub_status "planning"
else
    set_sub_status "building"
fi
timeout "${TIMEOUT:-20m}" opencode run "$PROMPT" --model=$MODEL
echo "-= AGENT finish =-"

commit_and_push "build"

if [[ "$PROMPT" == "/bdoc-idea "* ]]; then
  echo "PLAN_COMPLETE: Exiting as no-op."
  rm -f /log/sub_status
  exit 2
fi

# Automatic Test Execution Loop
MAX_TEST_RETRIES=4
RETRY_COUNT=0

while [ $RETRY_COUNT -le $MAX_TEST_RETRIES ]; do
    run_tests
    TEST_EXIT_CODE=$?
    
    commit_and_push "test run $((RETRY_COUNT + 1))"

    if [ $TEST_EXIT_CODE -eq 0 ]; then
        break
    fi
    
    if [ $RETRY_COUNT -lt $MAX_TEST_RETRIES ]; then
        echo "Tests failed. Asking agent to fix..."
        RETRY_COUNT=$((RETRY_COUNT + 1))
        set_sub_status "fixing tests ($RETRY_COUNT/$MAX_TEST_RETRIES)"
        
        TEST_OUTPUT=$(cat /log/test_output | tail -c 10000)
        FIX_PROMPT="The tests failed after your changes. Please fix the issues and ensure tests pass.
Test Output:
$TEST_OUTPUT"
        
        echo "-= AGENT restart (fix tests) =-"
        timeout "${TIMEOUT:-20m}" opencode run "$FIX_PROMPT" --model=$MODEL
        echo "-= AGENT finish (fix tests) =-"
    else
        echo "Tests failed after $MAX_TEST_RETRIES retries. Proceeding anyway."
        break
    fi
done

echo "Finishing $REPO_URL/$WORKER_BRANCH ..."

set_sub_status "verifying"
commit_and_push "final"

# Check if anything was committed at all
if [ "$(git rev-parse HEAD)" == "$INITIAL_SHA" ]; then
  echo "  NO_CHANGES_DETECTED: The agent produced no modifications."
  rm -f /log/sub_status
  exit 2
fi

rm -f /log/sub_status
