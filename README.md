# overdrive

**overdrive** is a workflow automation system designed to streamline software engineering tasks using AI agents running in isolated container environments.

## 1. Installation

### Prerequisites

- **Go**: Required to build the API server and scheduler binaries.
- **[Podman](https://podman-desktop.io/)**: Required to execute AI worker tasks and chat sessions in isolated containers.
- **`auth.json`**: An authentication file for the OpenCode service, located in the project root.

### Getting Started

To build and start the full system, simply run:

```bash
./scripts/rebuild-and-restart
```

Then, open your browser to [http://localhost:3281](http://localhost:3281).

## 2. High-level Overview of Code Organization

Overdrive is built around the concept of **Projects** and **Jobs**. A project is a workspace that hosts AI agents, and a job is a specific task assigned to those agents. When a job is triggered, the **Scheduler** spins up an isolated **Worker** container using **Podman** to execute the task safely. The **API Server** provides the web interface and API for managing this entire lifecycle.

```text
.
├── api/             # Core Go logic, HTTP handlers, and web templates
├── bin/             # Compiled project binaries (server and scheduler)
├── build_docs/      # Source documentation and development plans
├── cmd/             # Entry points for server and scheduler binaries
├── help_docs/       # Final documentation served by the help system
├── imgs/            # Dockerfiles and scripts for Podman containers
├── projects/        # Persistent data: configs, job logs, and chat data
├── scripts/         # Utility scripts for building and managing the app
└── static/          # Web dashboard assets (CSS, JS)
```

## 3. The Role of `build_docs/`

The `build_docs/` directory serves as the primary source of truth for all project-related documentation and tracking. Its role includes:

- **Documentation Source**: Markdown files (`.md`) describing the system architecture, API endpoints, job system, and deployment strategies.
- **Changelog**: Tracking version history and recent feature updates.
- **Feature Plans**: The `plans/` subdirectory houses active and archived plans for feature development and bug fixes.
- **Error Tracking**: Standardized error reporting and tracking are managed within the `errors/` subdirectory.
- **Documentation Updates**: Running the `/bdoc-update` command (either manually as a job prompt or automatically by the system) triggers the engineering agent to synchronize and refresh the documentation in `build_docs/` to match the current state of the codebase.
