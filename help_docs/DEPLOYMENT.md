# Remote Deployment & Execution

The overdrive project supports executing AI worker tasks on a remote host (e.g., a powerful build server or GPU machine) using `podman` and `ssh`. This allows for local development while offloading the heavy lifting of container execution.

## Architecture

The remote execution flow is handled primarily by `scripts/deploy`, which performs the following steps:

1.  **Build**: Builds the container image locally using `podman`.
2.  **Compress & Transfer**: Pipes the image to the remote host using SSH:
    ```bash
    podman save "$DOCKERFILE" | gzip | ssh "$REMOTE_HOST" 'gunzip | podman load'
    ```
    This ensures the remote host has the exact same image as your local machine.
3.  **Execute**: Runs the container on the remote host, passing all necessary environment variables (Git credentials, model info, task prompt).
4.  **Log Streaming**: Streams standard output and error from the remote container back to your local terminal and saves it to a local log file.

## Requirements

### Local Machine
*   **Podman**: Installed and available in PATH.
*   **SSH Client**: Configured to connect to the remote host without password prompts (key-based auth).

### Remote Host
*   **Podman**: Installed and running.
*   **SSH Server**: Accessible from the local machine.

## Configuration

Remote execution requires the `REMOTE_HOST` environment variable to be set.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `REMOTE_HOST` | The SSH alias or user@ip of the remote machine. | `overdrive-worker-01` or `user@192.168.1.50` |

## Scripts

*   **`scripts/demo-remote`**: A wrapper script that sets default environment variables and calls `scripts/deploy` to run a sample task on the remote host.
*   **`scripts/deploy`**: The core deployment script. It verifies that all required variables (`REPO_URL`, `SSH_KEY_PATH`, etc.) are set before proceeding.
*   **`scripts/debug-remote`**: Launches an interactive Bash shell inside a container on the remote host, useful for debugging environment issues.
