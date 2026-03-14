# Plan: Stop updating duration for crashed jobs

## Problem
Jobs that have crashed continue to have their duration updated in the UI because the `Duration()` method uses `time.Since()` if `CompletedAt` is not set. Some code paths that move a job to the "crash" status do not set `CompletedAt`.

## Proposed Changes
1.  **Update `moveToCrash` in `api/worker.go`**: Set `CompletedAt` to the current time when moving a job to the "crash" status.
2.  **Update `Duration()` in `api/job.go`**: Add a safety check to ensure that if a job is in a terminal state ("crash", "done", "no-op"), it doesn't use `time.Since()`, even if `CompletedAt` is missing.

## Verification Plan
1.  **Manual Verification**:
    *   Start the server.
    *   Trigger a "zombie" job scenario (though this is hard to do manually without stopping/starting).
    *   Alternatively, modify a job's YAML file manually to set status to "crash" and remove `CompletedAt`, then observe the UI.
2.  **Automated Tests**:
    *   Check if there are existing tests for `Duration()`.
    *   Add a test case for `Duration()` when status is "crash" and `CompletedAt` is nil.

status: done
