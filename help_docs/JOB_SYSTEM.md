# Job System

The **Job System** manages the lifecycle of worker tasks using a filesystem-based state machine. This allows for persistent job tracking without the overhead of a database. Jobs are isolated by project.

## Project Structure

Jobs and logs are stored within project-specific directories:
`projects/<project-name>/jobs/<state>/<job-id>.yml`
`projects/<project-name>/logs/<job-id>/worker.log`

## Job Lifecycle States

Jobs transition through the following directories within a project's `jobs/` folder:

*   **`pending/`**: Newly submitted jobs waiting to be picked up by the scheduler (displayed as "Todo" in the dashboard).
*   **`working/`**: Jobs currently being executed (displayed as "Build" in the dashboard).
*   **`done/`**: Successfully completed jobs (exit code 0).
*   **`crash/`**: Jobs that failed due to an error or non-zero exit code.
*   **`no-op/`**: Jobs that completed successfully but determined that no changes were necessary.
*   **`timeout/`**: Jobs that were terminated for exceeding the maximum execution time.
*   **`stopped/`**: Jobs that were manually terminated by the user while running.
*   **`undone/`**: Jobs that have been invalidated by a branch reset or rewind operation.
*   **`cancelled/`**: Jobs that were manually removed by the user while still in the pending queue.

## Todo System

The **Todo System** provides a hierarchical task management interface for each project, allowing users to plan complex features before submitting them as jobs.

### Key Features
- **Hierarchical Tasks**: Supports nesting subtasks within parent tasks to break down complex features.
- **Auto-Save**: Changes to titles and descriptions are automatically saved to the project's `todos.json` file using a debounced mechanism.
- **Provider Support**: Supports local storage (`native`) and external issue trackers like **Jira**.
- **Jira Sync**: When using the Jira provider, overdrive fetches issues via API and updates Jira status (e.g., to "Done") automatically when the corresponding overdrive job succeeds.
- **Job Integration**: Draft tasks can be "submitted" as jobs. Once submitted, the task is locked (`submitted` status) until the job completes.
- **Status Tracking**: Tasks reflect their execution state (e.g., `completed` or `crashed`) based on the outcome of the associated job.
- **Isolated Storage**: Todos are stored in `projects/<project-name>/todos.json`.

## Job Definition (YAML)

Each job is represented by a YAML file named `<job-id>.yml`. The Job ID is a Base62 encoded nanosecond timestamp (e.g., `Bq1s9Y2kL5`).

```yaml
id: "Bq1s9Y2kL5"
project: "my-project"
status: "pending"
created_at: "2026-02-09T12:34:56Z"
request:
  project: "my-project"
  repo_url: "..."
  branch_parent: "main"
  prompt: "implement feature xyz"
  model: "google/gemini-3-flash-preview"
exit_code: 0
```

## Execution Logic

1.  **Submission**: The API server writes a YAML file to `projects/<name>/jobs/pending/`.
2.  **Scheduling**: The scheduler picks up pending jobs from all projects and moves the file to `projects/<name>/jobs/working/`.
3.  **Execution**: The server executes `scripts/work` in a goroutine.
4.  **Logging**:
    *   `worker.log`: Standard output/error from the worker container.
5.  **Finalization**:
 The YAML file is moved to its final state directory based on the exit code of the work script.

## Special Prompt Prefixes

The system recognizes several special prefixes in the prompt for dashboard categorization and specialized behavior:

- `/bdoc-engineer`: Tags the job as a complex engineering task ("feature").
- `/bdoc-quick`: Tags the job as a simple fix or task ("quick change").
- `/bdoc-update`: Tags the job as a documentation update ("docs").
- `/bdoc-idea`: Tags the job as a plan task; exits as a "no-op" after generating ideas.
- `/bdoc-revert`: Reverts the changes from a specific commit.

## Manual Merge

The dashboard provides a **Merge** button next to the Job ID for completed jobs. Clicking this button pre-fills the "Submit New Job" form with the following details:
- **Type**: Automatically set to "FEATURE".
- **Prompt**: Pre-filled as `Merge branch <job-id> into <primary-branch>`.

This allows users to quickly initiate a merge job for the agent to handle.

## Docs Updates

To ensure project documentation stays in sync with code changes, the scheduler automatically enqueues a `/bdoc-update` job every 10 completed engineering jobs. This job is tagged as `docs` (displayed as "docs update" in the dashboard) and is intended to be handled by the agent to refresh relevant documentation files.

## Zombie Job Cleanup

On server startup, the system scans all projects. Any jobs found in `jobs/working/` are automatically moved to `jobs/crash/` with an "Interrupted" status.
