# API Server

The **overdrive API Server** is a Go-based service that manages multi-project job submissions, handles the background scheduler, and provides a modern, sidebar-driven web dashboard for monitoring progress, logs, and configuration.

## Overview

*   **Language**: Go (Golang)
*   **Port**: Default `3281` (configurable via `PORT` environment variable).
*   **Storage**: Filesystem-based project structure.
*   **Concurrency**: Manages simultaneous worker jobs with configurable limits.
*   **Theme System**: Supports multiple color schemes (Dracula, Nord, Cyberpunk, etc.) and custom theme creation via a dedicated selector in the user menu.
*   **UI Feedback**: Global toast notification system for real-time status updates and user feedback.

## Key Endpoints

### System Health and Monitoring

*   **`GET /up`**: Simple health check endpoint. Returns `OK` if the server is running.
*   **`GET /health-info`**: Returns system health information, including running container counts and uptimes for the API and scheduler. Supports HTMX snippets.
*   **`GET /dashboard`**: Displays the global dashboard with an overview of all projects, active jobs, and recent activity.
*   **`GET /api/dashboard/todos`**: Returns JSON data for the global dashboard, including active todos and recent job history.

### Project Management

*   **`GET /projects`**: Returns a list of all configured projects.
*   **`POST /projects`**: Saves or updates a project configuration. Payload includes repository details, SSH keys, AI harness settings, dependencies, and environment variables.
*   **`DELETE /projects/:projectId`**: Permanently deletes a project and all its associated data.
*   **`POST /projects/:projectId/resume`**: Resumes a paused project.

### Job Operations

*   **`GET /projects/:projectId/jobs`**: Displays the project-specific job dashboard. Supports filtering by status (`status`) and fuzzy search (`q`).
*   **`POST /projects/:projectId/jobs`**: Submits a new engineering job. Recognizes special prefixes like `/bdoc-engineer`, `/bdoc-quick`, and `/bdoc-update`.
*   **`POST /projects/:projectId/jobs/stop/:jobId`**: Manually stops a running job. Documentation updates (`/bdoc-update`) are protected from being stopped.
*   **`POST /projects/:projectId/jobs/revert/:jobId`**: Creates a new job to revert changes from a specific completed assignment.
*   **`GET /projects/:projectId/jobs/diff/:jobId`**: Displays the git diff for a completed engineering job.
*   **`GET /projects/:projectId/jobs/logs/:jobId`**: Provides a full, colorized log view for a specific job.

### Chat and Interaction

*   **`GET /projects/:projectId/chat`**: Displays the project's interactive chat interface.
*   **`POST /projects/:projectId/chat/send`**: Sends a message to the AI agent, including optional image attachments.
*   **`POST /projects/:projectId/chat/build`**: Converts a chat conversation directly into an engineering job.
*   **`GET /projects/:projectId/chat/proxy/:chatId/*`**: Proxies requests to the ephemeral chat container for streaming and interactivity.

### Todo Management

*   **`GET /projects/:projectId/todos`**: Displays the hierarchical Todo management interface.
*   **`POST /projects/:projectId/todos`**: Adds a new task or subtask.
*   **`PUT /projects/:projectId/todos/:todoId`**: Updates task details with auto-save support.
*   **`POST /projects/:projectId/todos/:todoId/submit`**: Converts a todo directly into an engineering job.

## Background Scheduler

The server runs a background loop that:
1.  **Polls**: Checks `projects/*/jobs/pending` for new work.
2.  **Execution**: Executes `scripts/work` in managed goroutines.
3.  **Completion**: Transitions jobs to their final state directory (e.g., `done`, `crash`) based on exit codes.

## Server Management

*   **Build**: `./scripts/server-build` compiles the Go binary.
*   **Run**: Use `go run api/*.go` or execute the compiled binary.
*   **Rebuild**: Use the **Danger Zone** actions in System Settings to recompile and restart the server or scheduler in-place.
