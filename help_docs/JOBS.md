[← Back to Help Home](/help)

# Jobs Overview

The Jobs system is the core execution engine where all engineering tasks, feature requests, and bug fixes are processed. Each job tracks a specific assignment for the AI worker.

## Understanding Job Statuses

A Job's lifecycle is reflected through its status in the dashboard:

*   **Todo** (Internal: `pending`): Queued and waiting for an available AI worker.
*   **Build** (Internal: `working`): Currently being processed. You can watch streaming, colorized logs in real-time.
*   **Done**: Successfully completed with an exit code of 0.
*   **Crash**: Encountered an error or non-zero exit code during execution.
*   **No-Op**: Successfully completed, but the AI agent determined no changes were necessary.
*   **Timeout**: Terminated after exceeding the maximum allowed execution time.
*   **Stopped**: Manually interrupted by a user.
*   **Undone**: Invalidated by a "Rewind" operation or similar branch reset.
*   **Cancelled**: Manually cancelled by a user while in a pending state.

## Job Features and Actions

Each job entry provides detailed insights and actions:

### Monitoring and Logs

*   **Streaming Logs**: Click on any active or completed job to view the live terminal output.
*   **ANSI Support**: Experience full ANSI-color logs for a traditional terminal feel.
*   **Diff View**: For completed engineering jobs, a clickable short commit hash provides a detailed view of the code changes introduced.

### Operations

*   **Merge**: Use the **Merge** button next to a completed job to quickly initiate a merge of that job's branch into the working branch.
*   **Duplicate**: Quickly create a new job pre-filled with the same prompt as a previous job.
*   **Revert**: Create a new job to undo the changes introduced by a specific completed job.

## How to Use Jobs

While most jobs run automatically after being submitted, you can interact with them directly:

1.  **Monitor Progress**: Use the project dashboard to track your requests in real-time.
2.  **Inspect Results**: Review logs and diffs for completed jobs to verify the AI's work.
3.  **Manage Conflicts**: If a job fails or requires adjustments, use Duplicate or Revert to refine the outcome.

Head back to the [Help Home](/help) if you need more assistance.
