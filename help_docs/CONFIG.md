[← Back to Help Home](/help)

# Project Configuration

The Project Configuration (or Settings) screen manages specific settings for each of your projects. You can access it by clicking the **Config** button in the project actions menu within the sidebar, or by navigating to the global **Settings** page.

## Project Details

Configuring a project involves several key areas:

### Repository and Code (CODE)

*   **Repo URL**: The Git SSH URL for your project repository.
*   **Project Name**: An optional name for your project; if empty, it's derived from the Repo URL.
*   **Working Branch**: The branch where the AI pushes its changes (e.g., `main` or `master`).
*   **SSH Key**: The private SSH key authorized for repository access.

### AI Configuration (AI)

Project execution is managed through a Build Harness and an interactive Chat Harness. You can configure each independently:

*   **Harness Providers**: Select from several orchestration tools, including OpenCode (default), Claude Code, Codex CLI, Cursor CLI, Gemini CLI, and GitHub CLI.
*   **LLM Providers**: Choose your preferred AI backend from Google Gemini, Anthropic, OpenAI, or Anysphere.
*   **Model**: Specify the exact model used (e.g., `google/gemini-3-flash-preview`).
*   **API Key**: Provide the API key for your chosen LLM provider. This can be entered directly or as an environment variable name (e.g., `$MY_API_KEY`).

### Dependencies (DEPS)

*   **Alpine Linux Deps**: A list of Alpine Linux packages (one per line) that will be installed in the project's container environment (e.g., `nodejs`, `python3`, `go`).

### Custom Commands (CMDS)

*   **Custom Build Commands**: A set of user-defined scripts that can be executed directly from the project dashboard. Each command has a label and the actual script to run.

### Environment Variables (ENV)

*   **Environment Variables**: Key-value pairs that are available as environment variables within the project's execution environment. This is useful for storing API keys or configuration secrets.

### Features and Integrations (FEAT)

The **Features** tab controls task management and external tool integrations:

*   **Native Todos**: A simple, built-in task management system within overdrive.
*   **Jira Integration**: Synchronizes tasks with a Jira project. Configuration requires:
    *   **Jira Instance URL**: (e.g., `https://yourcompany.atlassian.net`).
    *   **Project Key**: The uppercase key for your Jira project (e.g., `PROJ`).
    *   **Account Email**: The email address associated with your Jira API token.
    *   **API Token**: Your Jira API token or an environment variable name containing it.
    *   **Status Mapping**: Define which Jira statuses correspond to "Pick up" (In Progress) and "Done".

For information about **System Health** or the **Danger Zone**, please refer to the [System Settings](/help?doc=SYSTEM.md) page.
