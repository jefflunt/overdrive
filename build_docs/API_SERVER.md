# API Server

The **overdrive API Server** is a Go-based service that manages multi-project job submissions, handles the background scheduler, and provides a modern, sidebar-driven web dashboard for monitoring job progress, logs, and configuration.

## Overview

- **Language**: Go (Golang)
- **Port**: Default `3281` (configurable via `PORT` environment variable).
- **Storage**: Filesystem-based project structure.
- **Concurrency**: Manages simultaneous worker jobs with configurable limits.
- **Theme System**: Supports multiple color schemes (Dracula, Nord, Cyberpunk, etc.) and custom theme creation via a dedicated selector in the sidebar.
- **UI Feedback**: Global toast notification system for real-time status updates and user feedback.

## Endpoints

### `GET /up`
Health check endpoint. Returns `OK` if the server is running.

### `GET /`
Redirects to the dashboard of the first available project, or displays the project list if no projects are configured.

### `GET /projects`
Returns a list of all configured projects.

### `POST /projects`
Saves or updates a project configuration.
- **Payload**: JSON including `name`, `repo_url`, `ssh_key`, `ssh_config`, `primary_branch`, `build_model`, `chat_model`, `dependencies`, `env_vars`, `todo_provider`, and `jira` (optional configuration).
- **Jira Config**: Includes `enabled`, `instance` (URL), `project_key`, `email`, `api_token` (can be a `$ENV_VAR`), `status_pickup`, and `status_done`.
- **Todo Provider**: Can be set to `native` (default) or `jira`.
- **Dependencies**: A newline-separated list of Alpine Linux packages to be installed in the worker container.

### `GET /projects/:projectId/config`
Returns the configuration for a specific project as JSON. Used for the project editing modal.

### `DELETE /projects/:projectId`
Permanently deletes a project and all its associated jobs and logs.

### `POST /projects/:projectId/resume`
Resumes a project that was previously in a paused or interrupted state.

### `GET /projects/:projectId/jobs`
Returns an HTML dashboard listing jobs for a specific project:
- **Jobs**: List of all jobs for the project.
- **Active Jobs**: Currently executing jobs (displayed as "Build" in the UI, internal status `working`).
- **Pending Jobs**: Queued and waiting (displayed as "Todo" in the dashboard, internal status `pending`).
- **Finished Jobs**: Completed (`done`), failed (`crash`), timed out (`timeout`), skipped (`no-op`), manually stopped (`stopped`), reverted (`undone`), or cancelled (`cancelled`).
- **Short Commit Hashes**: For completed engineering jobs, the dashboard displays a clickable short commit hash (first 10 characters) that links directly to the job's diff.
- **Merge Button**: A dedicated button for completed jobs that pre-fills the prompt to merge the job's branch.
- **Query Parameters**:
    - `status`: Comma-separated list of statuses to filter by.
    - `q`: Fuzzy search query (matches against Job ID, Prompt, and formatted Prompt).
    - `offset`, `limit`: For pagination.

The dashboard uses **HTMX** for infinite scrolling (`hx-trigger="intersect"`) and real-time partial updates of job statuses and logs.

### `POST /projects/:projectId/jobs`
Submit a new job to the project.
- **Payload**: JSON
- **Fields**: `repo_url` (optional), `branch_parent` (optional), `commit_msg` (optional), `prompt`, `model`.
- **Special Prefixes**: Prompts starting with `/bdoc-engineer`, `/bdoc-quick`, `/bdoc-update`, `/bdoc-revert`, or `/bdoc-idea` are categorized as Feature, Quick Change, Docs, Revert, and Plan respectively.
- **Response**: `201 Created` with the `id` of the new job.

### `POST /projects/:projectId/jobs/revert/:jobId`
Creates a new job with the prompt `/bdoc-revert <commit>` to revert the changes introduced by a specific job.

### `POST /projects/:projectId/jobs/cancel/:jobId`
Cancels a pending or working job.

### `GET /projects/:projectId/jobs/diff/:jobId`
Displays the full git diff for a completed job in a dedicated page.

### `GET /projects/:projectId/jobs/logs/:jobId`
Displays the full log for a specific job in a dedicated page.

### `GET /projects/:projectId/jobs/logs-partial/:jobId`
Returns a partial HTML view of the logs for real-time dashboard embedding (HTMX).

### `GET /projects/:projectId/jobs/tail/:jobId`
Provides a dedicated "Tail" view of the logs that automatically scrolls to the bottom as new logs arrive.

### `GET /projects/:projectId/jobs/updates/:jobId`
Returns an HTML snippet with the current job status for HTMX-based updates.

### `POST /projects/:projectId/jobs/stop/:jobId`
Manually stops a running or pending job.
- **Restriction**: Automated documentation jobs (`/bdoc-update`) cannot be stopped to ensure system consistency.

### `DELETE /projects/:projectId/jobs/delete/:jobId`
Permanently deletes a specific job and its logs.

### `GET /settings`
Displays the system settings page.

### `POST /settings/global/save`
Saves global system settings.
- **Payload**: JSON with `max_global_containers`, `max_global_build_containers`, `max_global_chat_containers`, and `max_global_cmd_containers`.

### `POST /rebuild-restart`
Triggers the `scripts/rebuild-and-restart` script to update and restart the API server.

### `POST /rebuild-restart-scheduler`
Triggers the `scripts/rebuild-and-restart-scheduler` script to update and restart the background scheduler.

### `GET /health-info`
Returns system health information as JSON (or HTML snippet for HTMX), including running container counts and uptimes for the API and scheduler.

### `GET /auth/google/login`, `GET /auth/github/login`
Initiates OAuth login flows for Google and GitHub.

### `GET /auth/google/callback`, `GET /auth/github/callback`
Handles OAuth callbacks, exchanges codes for tokens, and sets authentication cookies.

### `GET /auth/logout`
Logs out the user and clears authentication cookies.

### `GET /projects/:projectId/chat`
Displays the chat interface for a project.
- **Query Parameters**:
    - `id`: The ID of the specific conversation to load.

### `GET /projects/:projectId/chat/list`
Returns an HTMX snippet of the chat sidebar list for real-time updates.

### `GET /projects/:projectId/chat/messages/:chatId`
Returns an HTMX snippet of the messages for a specific conversation.

### `POST /projects/:projectId/chat/create`
Creates a new conversation and returns the chat interface.

### `POST /projects/:projectId/chat/send`
Sends a message to an existing conversation.
- **Payload**: Multipart form with `chat_id`, `content`, and optional `images`.
- **Behavior**: Forwards the message to the ephemeral chat container via the internal `OpenCodeClient`.

### `POST /projects/:projectId/chat/delete/:chatId`
Moves a conversation to the trash or permanently deletes it if the `permanent=true` query parameter is provided.

### `POST /projects/:projectId/chat/restore/:chatId`
Restores a previously deleted conversation from the trash.

### `POST /projects/:projectId/chat/rename/:chatId`
Updates the title of a specific conversation.
- **Payload**: Form data with `title`.

### `POST /projects/:projectId/chat/build`
Converts a chat message into a new engineering job.
- **Payload**: Form data with `content`.
- **Behavior**: Redirects to the job dashboard with the pre-filled prompt.

### `GET /projects/:projectId/chat/warm/:chatId`
Ensures the ephemeral chat container for the specified session is running and ready for interaction.

### `GET /projects/:projectId/chat/proxy/:chatId/*`
Proxies requests directly to the ephemeral chat container. Used for streaming responses and other interactive features.

### `GET /projects/:projectId/chat/sync/:chatId`
Synchronizes the conversation history between the ephemeral container and the local YAML storage.

### `POST /projects/:projectId/cmds/:index/run`
Executes a custom project command. The response is a streamed text log of the execution.

### Todo Endpoints

The Todo system supports both **Native** (local JSON storage) and **Jira** providers. When Jira is enabled, overdrive synchronizes task statuses in real-time as engineering jobs complete.

#### `GET /projects/:projectId/todos`
Displays the hierarchical Todo management interface for a project. For Jira providers, this fetches and caches issues from the configured instance.

#### `POST /projects/:projectId/todos`
Adds a new root-level task to the project's Todo list.

#### `PUT /projects/:projectId/todos/:todoId`
Updates the title or description of a specific task. Supports auto-save from the UI.

#### `DELETE /projects/:projectId/todos/:todoId`
Deletes a task and all its nested subtasks.

#### `POST /projects/:projectId/todos/:todoId/star`
Toggles the star status of a specific task.
- **Payload**: JSON with `starred` (boolean).

#### `POST /projects/:projectId/todos/:todoId/submit`
Converts a draft task into an engineering job.

### Global Dashboard

#### `GET /dashboard`
Displays the global dashboard with an overview of all projects, running jobs, and recent activity.

#### `GET /api/dashboard/todos`
Returns JSON data for the global dashboard, including active todos per project, currently running jobs, and jobs completed within the last 24 hours.

### `GET /manifest.json`

Returns the Web App Manifest for PWA installation.

### `GET /sw.js`
Returns the Service Worker script for offline support and PWA functionality.

## Background Scheduler

The server runs a background loop that:
1.  **Polls**: Checks `projects/*/jobs/pending` for new work.
2.  **Execution**: Executes `scripts/work` in goroutines.
3.  **Completion**: Moves the job to its final state directory within the project based on its exit code.

## Server Management

- **Build**: `./scripts/server-build` (compiles the Go binary).
- **Run**: `go run api/*.go` or run the compiled binary.
- **Stop**: `./scripts/server-kill` (stops any running server instances).

## Template Helpers

The Go template engine is augmented with the following helper functions:
- `formatPrompt`: Renders markdown prompts into safe HTML.
- `markdown`, `renderMarkdown`: Alias for `formatPrompt`, renders markdown to HTML.
- `formatDate`: Formats `time.Time` objects as `2006-01-02 15:04:05`.
- `getBuildNumber`: Generates a build string (e.g., `b123`) based on git commit count.
- `dict`: Creates a map from a list of key-value pairs (useful for passing multiple data points to sub-templates).
- `add`: Performs simple integer addition.
- `hasPrefix`, `hasSuffix`, `contains`, `join`: Standard string manipulation functions.
- `json`: Marshals an interface to a JSON string.
- `listProjects`: Returns a list of all projects.
