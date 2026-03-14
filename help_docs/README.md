# overdrive Project

**Version: 0.8.1**

**overdrive** is a workflow automation system
 designed to streamline software engineering tasks using AI agents running in isolated container environments. It features a Go-based API server for multi-project job management and a modern, sidebar-driven web dashboard.

## System Overview

The system provides an API to submit engineering tasks across multiple projects. A background scheduler manages ephemeral containers ("workers") that connect to git repositories, perform tasks using an AI model, and push results back to new branches.

## Key Features

- **Global Dashboard**: Unified view of all project tasks, active builds, and recent job history across the entire system.
- **PWA Ready**: Install overdrive as a Progressive Web App for a desktop-like experience with offline support.
- **Project Chat**: Dedicated AI chat interface for each project with image support and persistent history.
- **Color Themes**: Choose from several pre-configured dark themes (e.g., Dracula, Nord, Cyberpunk) or create your own custom theme via the user menu.
- **Multi-Project Management**: Isolated environments for different repositories with individual SSH and AI configurations.
- **Job Search**: Quickly find previous jobs using the real-time fuzzy search interface.
- **Revert and Replay**: Easily revert specific jobs or duplicate existing tasks to run them again with adjustments.
- **Live Logs**: Watch engineering agents work in real-time with streaming, ANSI-color logs.
- **Custom Environments**: Install project-specific Alpine Linux dependencies and configure persistent environment variables.
- **Isolated Execution**: Secure execution in ephemeral containers that are automatically cleaned up after completion.
- **Project Commands**: Define and execute custom scripts directly from the project interface with live output streaming.
- **Starring**: Highlight important Todo items to prioritize your workflow.

### Key Directories

- **`api/`**: The Go-based API server and web dashboard.
- **`projects/`**: Project-specific configurations, SSH keys, and job logs.
- **`scripts/`**: CLI tools and the primary worker execution script (`work`). **Note: All scripts follow the [Scripting Standards](WORKER_INFRASTRUCTURE.md#logging-standards).**
- **`imgs/`**: Dockerfile definitions for the worker environment.
- **`build_docs/`**: Project documentation and error reporting.

## Documentation Map

*   **[Changelog](CHANGELOG.md)**: Project version history and recent changes.
*   **[API Server](API_SERVER.md)**: Details on the Go service, endpoints, dashboard, and PWA features.
*   **[Job System](JOB_SYSTEM.md)**: Technical details on the file-based queue and job lifecycle.
*   **[Worker Infrastructure](WORKER_INFRASTRUCTURE.md)**: Technical details about the container environment and entrypoint.
*   **[Remote Deployment](DEPLOYMENT.md)**: Architecture for offloading tasks to remote hosts manually.
*   **[Error Logs](errors/README.md)**: Standardized error reporting and tracking (all errors are tracked in `build_docs/errors/`).

## Quick Start

### Prerequisites
1.  **Podman**: Must be installed and running.
2.  **SSH Key**: A valid SSH private key authorized for the target repository.
3.  **`auth.json`**: A valid OpenCode authentication file located in the project root.

### Usage

Run the demo script to verify your setup:

```bash
# Run a simple task locally
./scripts/demo-local
```

This will:
1.  Build the `worker` image.
2.  Run a container that clones the repo.
3.  Execute a simple prompt ("hi, how are you?").
4.  Logs will be saved to `projects/<project-name>/logs/<job-id>/`.

For remote execution, refer to **[Remote Deployment](DEPLOYMENT.md)**.
