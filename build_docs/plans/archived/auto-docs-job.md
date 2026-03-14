# Plan: Automatic Docs Job Enqueueing

Update the scheduler to automatically enqueue a `docs` job every 11th job per project.

## Tasks
1. [x] Locate the scheduler logic in `api/worker.go`.
2. [x] Add a way to track the number of jobs processed per project.
3. [x] Modify the `schedule()` function to increment the counter and enqueue a `/bdoc-update` job every 11th job.
4. [x] Implement a helper function to enqueue jobs programmatically.
5. [x] Verify the implementation with a new test in `api/worker_test.go`.

## Status
status: done
