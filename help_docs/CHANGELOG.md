# Changelog

All notable changes to the overdrive project will be documented in this file.

## [0.8.1] - 2026-03-02

### Added
- **Compact Dashboard Todos**: Implemented a more compact view for tasks on the dashboard to improve space utilization.
- **Editable Global Concurrency**: Global concurrency limits (Containers, Build, Chat, Cmd) are now editable and persistent via the Settings page.

### Changed
- **UI Terminology**: Renamed dashboard headings "Starred items" to "Started todos" and "All Projects" to "All todos" for better clarity.
- **UI Terminology**: Updated global settings label "Max Global Concurrent Jobs" to "Max Global Containers" for consistency with system terminology.

## [0.8.0] - 2026-03-01

### Added
- **Global Dashboard**: Unified view of all project tasks, active builds, and recent job activity across the system.
- **Project Commands**: Support for defining and executing custom scripts directly from the project interface with live output streaming.
- **Todo Starring**: Added ability to star important tasks for better visibility.
- **Job Cancellation**: New endpoint to cancel pending or working jobs.
- **Enhanced Revert**: Consistent formatting for revert jobs in the UI.

### Changed
- **API Documentation**: Synchronized `API_SERVER.md` with missing endpoints and updated job status lists.
- **Version Bump**: Updated system version to 0.8.0.

## [0.7.9] - 2026-02-23

### Changed
- **UI Refinement**: Repositioned the "AI" tab in the project configuration modal to follow the "CODE" tab, improving the initial setup experience.
- **UI Refinement**: Updated the homepage button style in the sidebar to use a thinner border and a neutral grey color.

### Fixed
- **Stability**: Implemented asynchronous container cleanup during scheduler restart to ensure a clean state without blocking the UI.

## [0.7.8] - 2026-02-22

### Added
- **Chat Management**: Added endpoints for renaming, deleting, and restoring chat conversations.
- **Chat-to-Build**: New feature to quickly submit an engineering job directly from a chat conversation.
- **Project Resume**: Added ability to resume projects that were previously in a paused or interrupted state.
- **Documentation Update**: Synchronized technical guides with the current API implementation, adding missing endpoints for Chat, Project management, and Health monitoring.

### Fixed
- **API Documentation**: Completed the `API_SERVER.md` with missing endpoints: `/chat/rename`, `/chat/delete`, `/chat/restore`, `/chat/build`, `/chat/warm`, and `/projects/:projectId/resume`.

## [0.7.7] - 2026-02-20

### Added
- **Test Coverage**: Increased test coverage for API handlers and worker logic by 20%.

## [0.7.6] - 2026-02-20

### Added
- **Documentation Refactor**: Moved user-facing documentation to `help_docs/` for better separation of concerns.
- **Help Center**: Overhauled the `/help` page into a centralized support hub with navigation cards and improved layout.

### Changed
- **Navigation**: Updated the help icon to always navigate to the Help Center home page.
- **UI Performance**: Refined the Help page layout by removing the sidebar in favor of a full-width experience.

## [0.7.5] - 2026-02-17

### Added
- **Jira Integration**: Bi-directional sync for project tasks.
    - Support for Jira as a Todo provider with automated status updates.
    - Secure API token handling via environment variables.
    - Hierarchical task support (Epics/Tasks/Subtasks) mapped to the overdrive UI.
- **UI Refinements**:
    - Renamed "ENV VARS" tab to "ENV" in project settings for a cleaner look.
    - Updated homepage copy to reflect multi-device support (desktop, phone, and tablet).

### Fixed
- **Todo System**: Resolved a race condition in the Todo expansion toggle by implementing event delegation and `onmousedown` listeners.
- **Homepage**: Removed redundant marketing copy to improve focus on core features.

## [0.7.4] - 2026-02-16

### Added
- **Theme System**: Comprehensive theme selector with 11 built-in schemes (Dracula, Monokai, Nord, Synthwave '84, etc.) and a custom theme overdrive.
- **Project Environments**: Support for custom Alpine Linux dependencies and persistent database configurations (PostgreSQL, MySQL, SQLite) per project.
- **Improved Job Control**: Added "stop" functionality for active jobs, with safety locks for automated documentation tasks.
- **Documentation Refresh**: Updated technical guides to reflect recent UI changes and system features.

### Changed
- **Version Bump**: Updated system version to 0.7.4.

## [0.7.3] - 2026-02-15

### Added
- **Todo System Enhancements**:
    - **Auto-Save**: Implemented debounced auto-save for all todo fields, ensuring progress is never lost.
    - **Subtask Management**: Improved the hierarchy logic for adding and managing nested subtasks.
- **UI Refinements**:
    - **Global Toast Notifications**: Implemented a global toast notification system for real-time user feedback on actions like project creation, deletion, and job submission.
- **Commit Links**: Replaced separate "diff" links with clickable short commit hashes in the job dashboard.
- **Empty State UX**: Replaced "No messages found" with "build what you wish existed" in the chat interface for a more inspiring start.
- **Clutter Reduction**: Removed redundant job type badges from the status column to improve scanability.

### Changed
- **Todo UI**: Removed the manual "Save" button and "Draft" labels in favor of the new auto-save system.
- **Form UX**: Removed confusing default placeholder text from the todo description field.

## [0.7.2] - 2026-02-13

### Added
- **Real-time Project Chat**: Overhauled the chat system to use ephemeral containers for live interaction.
    - Support for multi-turn conversations with context preservation.
    - Image upload support (Base64) in chat messages.
    - Automated message synchronization between containers and YAML storage.
- **Job History Management**:
    - **Revert**: Added `/bdoc-revert` command to revert the project state to a specific commit.
- **Enhanced Live Monitoring**:
    - Streaming ANSI-color logs via HTMX for real-time progress tracking.
    - Live job status updates in the dashboard without full page refreshes.
- **System Settings**: Introduced a dedicated settings interface for system-level maintenance.
    - Integrated "Rebuild & Restart" for both the API server and the background scheduler.
- **OAuth Expansion**: Completed integration for Google and GitHub authentication.

### Changed
- **UI Performance**: Switched to HTMX-based partial updates for all major dashboard components.
- **Documentation**: Updated documentation map and technical guides to reflect the new architecture.

## [0.7.1] - 2026-02-12

### Added
- **Merge Button**: Added a dedicated merge button next to Job IDs in the dashboard to quickly initiate branch merges.
### Changed
- **Automatic Docs**: The scheduler now automatically enqueues a `/bdoc-update` job every 10 engineering jobs per project to keep documentation in sync.

## [0.7.0] - 2026-02-11

### Added
- **Project Chat**: Introduced a dedicated chat interface for each project.
    - Persistent YAML-based conversation storage.
    - Support for multi-turn dialogues and image uploads (Base64).
    - Real-time HTMX-powered messaging.
- **OAuth Integration**: Added infrastructure for Google and GitHub authentication.
- **Build System Enhancements**:
    - Automated build numbers (`b<count>`) based on git history.
    - Improved settings page with rebuild and restart functionality for both API server and background scheduler.
- **Template Helpers**: Added new functions to the Go template engine: `formatDate`, `getBuildNumber`, `dict`, and `add`.

### Changed
- **UI Navigation**: Updated the sidebar layout to include links for Chat and Settings.
- **Documentation**: Updated documentation to reflect the new API endpoints and features.

## [0.6.3] - 2026-02-11

### Added
- **Project Model Selection**: Added a dropdown to the Project modal to select the default Gemini model for each project.
- **Home Screenshots**: Added automated app screenshots to the homepage for better visual introduction.
- **Hover Delay**: Improved sidebar project actions with a 1.5s hover delay to prevent accidental hiding.
- **Job Duration**: Display "pending" for jobs in "Todo" status.

### Changed
- **Cleanup**: Removed outdated V2.0 banner from the homepage.

## [0.6.2] - 2026-02-10

### Added
- **PWA Support**: Added Progressive Web App (PWA) functionality, including a web manifest, service worker for offline support, and an "Install App" button in the sidebar.
- **Job Search**: Implemented a real-time search box in the dashboard to filter jobs by ID or prompt text.
- **Project Configuration**: Added better handling for SSH keys and configs, allowing them to be stored as files within the project directory.

### Changed
- **UI Terminology**: Updated job status labels in the dashboard for better clarity: "Pending" is now displayed as "Todo", and "Working" is now displayed as "Build".
- **Visual Styles**: Updated status neon effects and pulsing animations to match the new "Build" terminology.
- **Documentation**: Overhauled the documentation to reflect the removal of the workflow system and the simplified job execution model.

## [0.6.1] - 2026-02-09

### Added
- **Settings Page**: Introduced a dedicated settings page accessible via the sidebar, providing system-level controls.
- **Rebuild & Restart**: Added a button in the settings page to automatically pull the latest code, rebuild the server, and restart the process.

### Changed
- **Job Sorting**: Fixed job sorting to be reverse chronological (newest first) based on their Base62 IDs.
- **UI Refinement**: Removed the background grid from the settings page for a cleaner look and renamed "Active Jobs" to just "Jobs" in the dashboard.

## [0.6.0] - 2026-02-09

### Added
- **Visual Redesign**: Overhauled the user interface to use a sidebar-based navigation for projects, improving accessibility and space utilization.
- **Project Configuration Editing**: Added a gear icon next to projects in the sidebar, allowing users to modify existing project settings (SSH keys, repo URL, etc.) via a modal.
- **Workflow Selection**: The project creation and editing forms now include a dropdown menu for selecting from available YAML-defined workflows.

### Changed
- **RESTful API Routes**: Refactored job-related endpoints to follow a nested RESTful pattern under projects (e.g., `/projects/:projectId/jobs`).
- **Simplified Job Submission**: The "New Job" form now automatically inherits the repository URL, parent branch, and default workflow from the project configuration, reducing manual input.

### Fixed
- **Stability**: Improved server resilience by ensuring proper cleanup of "zombie" jobs during restart and fixing missing dependency imports.

## [0.5.0] - 2026-02-09

### Added
- **Project Management**: Introduced a multi-project architecture. Users can now manage multiple repositories independently, each with its own SSH configuration and job history.
- **Configurable Workflows**: Added support for custom workflows defined in YAML. Workflows allow for multi-step execution and fine-grained concurrency control using dynamic keys.
- **Project-specific Job Queues**: Jobs are now isolated within project directories (`projects/<name>/jobs/`), preventing state collision between different repositories.
- **Improved UI for Projects/Workflows**: Added dedicated dashboard views for managing projects and workflows using HTMX for a seamless experience.

### Changed
- **Directory Structure**: Moved job and log storage from the root into the `projects/` directory.
- **Job Submission**: The job submission form now requires selecting a project and optionally a specific workflow.

## [0.4.3] - 2026-02-08

### Added
- **Job Duplication**: Introduced a "Duplicate" feature for all terminal jobs (done, crash, timeout, etc.). This allows users to quickly repopulate the job submission form with the parameters of a previous job for easy modification and resubmission.
- **Auto-expanding Prompt**: The job submission form now features an auto-expanding textarea for the prompt, providing a better editing experience for long instructions.

### Changed
- **UI Enhancements**:
    - "More/Less" links in the dashboard are now dynamically displayed only when prompt text actually overflows the container.
    - Improved log view loading state with a "Loading logs..." indicator.
- **Job Management**: Replaced the "Retry" functionality with the more flexible "Duplicate" system.

### Removed
- **Retry Endpoint**: The `/jobs/retry/<job-id>` endpoint has been removed in favor of the duplication workflow.

## [0.4.2] - 2026-02-08

### Added
- **Markdown Support**: The jobs dashboard now renders prompts using Markdown, providing better readability for structured instructions and code snippets.

### Changed
- **UI Refinement**:
    - Improved prompt display in the dashboard with line clamping and more/less toggles.
    - Optimized HTMX partial updates for job rows, reducing flicker and improving performance.
    - "Quick Change" is now the default selection for new jobs.

## [0.4.1] - 2026-02-07

### Fixed
- **Job Duration**: Fixed a bug where crashed jobs continued to show an increasing duration in the UI. Terminal states ("crash", "done", "no-op") now correctly freeze the duration.
- **Job Finalization**: `CompletedAt` timestamp is now correctly set when a job is moved to "crash" status.

### Changed
- **UI Improvements**:
    - Rounded job duration to the nearest second for better readability in the dashboard.
    - Simplified the "Exit Code" column to show only the numeric value; detailed error messages are now available as tooltips on the job status.

## [0.4.0] - 2026-02-06

### Added
- **Seamless Refresh**: The jobs page now uses HTMX for partial updates, preserving user selection and scroll position during auto-refresh.
- **ANSI Color Logs**: The web dashboard and log view now support full ANSI color rendering for better log readability.

## [0.3.0] - 2026-02-05

### Added
- **Remote Execution**: Support for offloading worker tasks to remote hosts via `scripts/deploy` and SSH.
- **Worker Infrastructure**: Standardized Alpine-based worker image with pre-installed engineering tools.
- **Error Tracking**: Established standardized documentation for error reporting and debugging workflows.

## [0.2.0] - 2026-02-04

### Added
- **Job Queue**: Implemented the filesystem-based state machine for job tracking (`pending`, `working`, `done`, `crash`).
- **CLI Tools**: Introduced `scripts/work` and `scripts/demo-local` for manual job execution and testing.

## [0.1.0] - 2026-02-01

### Added
- **API Server**: Initial release of the Go-based API server with endpoints for job submission and monitoring.
- **Basic Dashboard**: Simple HTML dashboard to view the state of the job queue.
