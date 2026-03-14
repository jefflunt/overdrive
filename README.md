# overdrive

**overdrive** is a workflow automation system designed to streamline software engineering tasks using AI agents running in isolated container environments.

## 1. Installation

### Prerequisites

- **Go**: Required to build the API server and scheduler binaries.
- **[Podman](https://podman-desktop.io/)**: Required to execute AI worker tasks and chat sessions in isolated containers.
- **`auth.json`**: An authentication file for the OpenCode service, located in the project root.

### Building

The project includes two primary components that need to be built:

1. **API Server**: The web dashboard and core service.
   ```bash
   ./scripts/server-build
   ```
   This command compiles the server and outputs the binary to `bin/overdrive`.

2. **Scheduler**: Manages background jobs and worker lifecycles.
   ```bash
   ./scripts/scheduler-build
   ```
   This command compiles the scheduler and outputs the binary to `bin/scheduler`.

### Running

To start the full system:
1. Start the API server: `./bin/overdrive`
2. Start the scheduler: `./bin/scheduler`

For more detailed setup information, refer to [help_docs/DEPLOYMENT.md](help_docs/DEPLOYMENT.md).

## 2. High-level Overview of Code Organization

The codebase is organized as follows:

- **`api/`**: Contains the core Go logic for the API server, including HTTP handlers (`handlers_*.go`), data models, and the `templates/` for the web interface.
- **`cmd/`**: Entry points for the application's binaries.
  - `api-server/`: Main function for the web server.
  - `scheduler/`: Main function for the background job scheduler.
- **`bin/`**: Destination for compiled project binaries.
- **`projects/`**: Data directory for all projects. This includes project configurations (`project.yml`), job logs, and individual project data like `chats_data/`.
- **`imgs/`**: Contains Dockerfiles and entrypoint scripts used by Podman to build worker and chat environments.
- **`scripts/`**: A collection of utility scripts for building, deploying, and running common tasks (e.g., `server-build`, `scheduler-build`, `deploy`).
- **`static/`**: Static assets for the web dashboard, such as CSS and JavaScript files.
- **`build_docs/`**: Source directory for project documentation and development plans.
- **`help_docs/`**: Final documentation files served by the application's help system.

## 3. The Role of `build_docs/`

The `build_docs/` directory serves as the primary source of truth for all project-related documentation and tracking. Its role includes:

- **Documentation Source**: Markdown files (`.md`) describing the system architecture, API endpoints, job system, and deployment strategies.
- **Changelog**: Tracking version history and recent feature updates.
- **Feature Plans**: The `plans/` subdirectory houses active and archived plans for feature development and bug fixes.
- **Error Tracking**: Standardized error reporting and tracking are managed within the `errors/` subdirectory.
- **Documentation Updates**: Running the `/bdoc-update` command (either manually as a job prompt or automatically by the system) triggers the engineering agent to synchronize and refresh the documentation in `build_docs/` to match the current state of the codebase.
