[← Back to Help Home](/help)

# Project Configuration

The Project Configuration (or Settings) screen allows you to manage the specific settings for each of your projects. You can access it by clicking the "config" button in the project actions menu in the sidebar, or by navigating to the global **Settings** page.

## Project Details

When creating or editing a project, you can configure the following:

### Code & Repository (CODE)

*   **Repo URL**: The Git SSH URL for your project repository.
*   **Project Name**: An optional name for your project. If left empty, it's derived from the Repo URL.
*   **Working Branch**: The main branch of your repository (e.g., `main` or `master`) where the AI will push its changes.
*   **SSH Key**: The private SSH key used to access your repository.

### AI Configuration (AI)

*   **Build Harness & LLM Providers**: Select the orchestration harness (OpenCode or Claude Code) and the underlying LLM provider (Google Gemini or Anthropic).
*   **Build Model**: The specific LLM model used for background build tasks (e.g., `google/gemini-3-flash-preview`).
*   **Chat Model**: The specific LLM model used for interactive chat sessions.

### Dependencies (DEPS)

*   **Alpine Linux Deps**: A list of Alpine Linux packages (one per line) that need to be installed in the project's container environment (e.g., `nodejs`, `python3`, `go`).

### Custom Commands (CMDS)

*   **Custom Build Commands**: A list of commands to run during the build process. These commands execute within the project's isolated container environment, allowing you to run tests, linters, or deployment scripts directly from the dashboard. Output is streamed live to the UI.

### Environment Variables (ENV)

*   **Environment Variables**: Key-value pairs that will be available as environment variables within the project's execution environment. This is useful for API keys, database URLs, etc.

### Features (FEAT)

The **Features** tab (also found in the global Settings page) allows you to configure how tasks and todos are managed:

*   **Native Todos**: A simple, built-in task management system within the overdrive application.
*   **Jira Integration**: Synchronize your tasks with a Jira project. When enabled, you'll need to provide:
    *   **Jira Instance URL**: e.g., `https://yourcompany.atlassian.net`.
    *   **Project Key**: The uppercase key for your Jira project (e.g., `PROJ`).
    *   **Account Email**: The email address associated with your Jira API token.
    *   **API Token**: Your Jira API token (or an environment variable name containing it).
    *   **Status Mapping**: Define which Jira statuses correspond to "Pick up" (In Progress) and "Done".

---

Looking for information about **System Health** or the **Danger Zone**? Those can be found in the [System Settings](/help?doc=SYSTEM.md) page.
