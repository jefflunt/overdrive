# Plan: Combined Duration for Building and Merging

The "duration" time display should show the combined time between building and merging, whereas right now it only shows the building time.

## Proposed Changes

### 1. Update `Job.Duration()` in `api/job.go`
Currently, `Duration()` returns a fixed duration if `CompletedAt` is set. When a job is in `built` status, `CompletedAt` is set, so when it later moves to `merging` status, it continues to show the duration of the building phase.

- Modify `Duration()` to only return a fixed duration if the job is in a terminal state (done, merged, crash, etc.) or specifically in the `built` state.
- If the job is `merging`, it should show the time since `StartedAt`, which will include both the building time and the time spent merging so far.

### 2. Update `updateOriginalJobToMerged` in `api/worker.go`
When a separate merge job successfully merges a feature branch, it calls `updateOriginalJobToMerged` to move the original job to `merged` status. Currently, it doesn't update the `CompletedAt` timestamp of the original job.

- Update `updateOriginalJobToMerged` to set the original job's `CompletedAt` to the current time. This ensures that the final duration of the merged job reflects the time from its start until it was actually merged.

## Verification Plan

### Automated Tests
- Update `api/job_test.go` to include a test case for a job in `merging` status with `CompletedAt` set, ensuring it shows the ticking duration.
- Add a test case for `merged` status to ensure it shows the total duration.

### Manual Verification
1. Start a job that will be built.
2. Observe its duration while `working` (ticking).
3. Observe its duration when `built` (fixed).
4. Wait for it to move to `merging` (auto-merge or manual).
5. Observe its duration while `merging` (should start ticking again from where it left off, or rather, from `StartedAt` to now).
6. Observe its duration when `merged` (fixed at combined time).

status: done
