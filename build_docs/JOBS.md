[← Back to Help Home](/help)

# Jobs Overview

Welcome to the Jobs system! This is the engine room of your project where all the heavy lifting happens. Whenever you ask for a change, a new feature, or a bug fix, a "Job" is created to track that work.

## Understanding Jobs

Think of a Job as a specific assignment for the AI worker. It contains all the instructions needed to complete a task, from writing code to running tests.

### Key Concepts

*   **Status**: Every job has a status that tells you what's happening.
    *   **Todo** (Internal: `pending`): Waiting for a worker to pick it up.
    *   **Build** (Internal: `working`): Currently being processed. You can watch live logs in real-time!
    *   **Done**: Successfully completed.
    *   **Crash/Error**: Something went wrong. Check the logs for details.
    *   **Stopped**: Manually interrupted by a user.
    *   **Undone**: Invalidated by a "Rewind" operation.

*   **Logs**: By clicking on a job, you can view the detailed streaming ANSI-color logs of what the AI is doing. This is great for debugging or just satisfying your curiosity.

*   **Commits & Diffs**: For completed engineering jobs, you'll see a clickable short commit hash. Clicking this will show you the exact code changes (the "diff") introduced by that job.

*   **Merge Button**: Once a job is successfully completed on its own branch, you can use the **Merge** button to quickly initiate a merge into your primary branch.

### How to Use Jobs
Most of the time, jobs run automatically. However, you can:
1.  **Monitor Progress**: Watch the job list to see your request being fulfilled in real-time.
2.  **View Logs**: Click on any job ID or status to see the terminal output.
3.  **Duplicate**: If a job fails or you want to run it again with changes, use the Duplicate feature to pre-fill a new job request.
4.  **Revert & Rewind**: Use the Revert button to undo a specific job, or Rewind to hard-reset your branch to a previous stable state.

We hope this helps you understand the Jobs system better!
