# Worker Infrastructure

The "worker" is a containerized environment designed to execute OpenCode agents and scripts in isolation. It is defined in `imgs/worker`.

## Container Image

### `imgs/worker`
The worker image is built dynamically by `scripts/work`, using `imgs/base/Dockerfile` as a foundation. It installs the necessary dependencies to run the OpenCode CLI and git operations.

**Key Components:**
*   **Base**: `alpine` (via `imgs/base/Dockerfile`)
*   **System Dependencies**: `bash`, `curl`, `git`, `openssh`, `openssh-client`, `libc6-compat`, `libstdc++`, `gcompat`, `build-base`, `linux-headers`, `jq`.
*   **Project Dependencies**: Custom Alpine packages can be specified in the project configuration and are installed via `apk add --no-cache` during the worker image build.
*   **OpenCode CLI**: Installed via `curl -fsSL https://opencode.ai/install | bash`.
*   **Configuration**:
    *   Copies `.opencode/` to `/root/.opencode/` (Agent configuration).
    *   Copies `auth.json` to `/root/.local/share/opencode/auth.json` (Authentication).
    *   Copies `ssh.key` to `/ssh.key` (Git authentication, handled by build scripts).
    *   Copies `ssh_config` to `/root/.ssh/config` (SSH host verification settings).
*   **Database Integration**: If configured in the project settings, database connection details are passed to the worker via environment variables: `DB_TYPE`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`.
*   **Runtime**:
    *   `imgs/worker-entrypoint.sh` is used as the entrypoint.

## Entrypoint
The entrypoint script (`imgs/worker-entrypoint.sh`) handles:
1.  **Logging Setup**: Redirects stdout/stderr to `/log/worker.log` for real-time monitoring. **All output is prefixed with a timestamp** to ensure a clear audit trail of events.
2.  **Repository Setup**: Clones the target repository (using `REPO_URL`).
3.  **Branch Management**: Checks out the parent branch and creates the new child branch.
4.  **Task Execution**: Runs the OpenCode CLI command provided in the `PROMPT` environment variable.
5.  **Commit & Push**: Commits the changes (if any) and pushes the new branch to the remote.

### Logging Standards
All shell scripts in the overdrive project (entrypoints and helper scripts in `scripts/`) must include timestamped logging for consistency. The following boilerplate should be included at the top of every script:

```bash
# Setup logging with timestamps
exec > >(while IFS= read -r line || [ -n "$line" ]; do
    if [[ ! "$line" =~ ^\[[0-9]{4}-[0-9]{2}-[0-9]{2} ]]; then
        printf "[%(%Y-%m-%d %H:%M:%S)T] %s\n" -1 "$line"
    else
        printf "%s\n" "$line"
    fi
done) 2>&1
```

This ensures that:
- Every line of output (stdout and stderr) is timestamped.
- Lines that already have a timestamp (e.g., from nested scripts) are not double-timestamped.
- Performance is optimized by using Bash builtins for timestamp generation.

For scripts where you also need to see the exact shell commands being executed (similar to `set -x`), you can combine the above with a custom `PS4`:

```bash
export PS4='+ [%(%Y-%m-%d %H:%M:%S)T] '
set -x
```

For entrypoints that also need to write to a log file, use `| tee -a /path/to/log` inside the `while` loop's process substitution.

Note: Git user configuration and SSH key placement are handled during the image build process (see `imgs/worker`).

## Usage
The worker is primarily invoked via the helper scripts in `scripts/`.
- `scripts/work`: The main script used by the API server to build and run the worker.
- `scripts/deploy`: Used for remote execution (see [Remote Deployment](DEPLOYMENT.md)).

## Chat Session Containers
For real-time project chat, the system uses ephemeral "chat session" containers. These containers run `opencode serve` to provide an interactive API for the chat interface.

**Key Differences from Job Workers:**
- **Entrypoint**: Uses `imgs/chat-server-entrypoint.sh`.
- **Persistence**: Mounts `projects/<project-name>/chats_data/<chat-id>` to `/root/.local/share/opencode` to persist conversation history and settings.
- **Networking**: Maps a random host port to container port `3000`. The API server proxies requests to this port.
- **Lifecycle**: Containers are automatically reaped after 5 minutes of inactivity.
- **Repository Management**: Clones the repository once and stays running, allowing for fast, multi-turn interactions within the same environment.

**Implementation Details:**
- **Script**: `scripts/chat-server` builds the chat-specific image and manages container startup.
- **Session Management**: `api/chat_session.go` handles session tracking, health checks, and reaping idle containers.
- **Proxying**: The API server uses `httputil.NewSingleHostReverseProxy` to tunnel chat messages and synchronization requests to the appropriate container.
