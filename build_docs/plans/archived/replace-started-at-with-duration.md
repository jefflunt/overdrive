# Plan: Replace "Started at" and "Created at" with "Duration" in Active Jobs table

## Problem
The "Active Jobs" table currently shows "Created At" and "Started At" columns. We want to replace these with a single "Duration" column that shows the time elapsed since the job started.

## Proposed Changes
1.  **Update `Job` struct in `api/job.go`**: Add a method `Duration() string` (or similar) that returns the elapsed time for active jobs and the total duration for completed jobs.
2.  **Update `api/templates/jobs.html`**:
    *   In the "Active Jobs" table, remove "Created At" and "Started At" columns.
    *   Add a "Duration" column.
    *   Update the table body to display the duration.
    *   (Optional but recommended) Update the "Finished Jobs" table to use the new method for consistency.

## Verification Plan
1.  **Build and Run**: Start the API server.
2.  **Submit a Job**: Submit a new job and observe the "Active Jobs" table.
3.  **Check Duration**: Ensure the "Duration" column shows the time since the job started.
4.  **Verify Finished Jobs**: Ensure the "Duration" column in "Finished Jobs" still shows the correct duration.
5.  **Manual UI Check**: Verify the UI looks as expected and columns are correctly aligned.

status: done
