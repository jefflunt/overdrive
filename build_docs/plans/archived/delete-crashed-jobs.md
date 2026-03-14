# Plan: Delete Link for Crashed Jobs

Add a "delete" link to all crashed jobs in the UI to allow users to remove them from the list.

## Proposed Changes

### 1. Backend: Implement Delete Job Handler
- Add `HandleDeleteJob` to `api/handlers.go`.
- This handler will:
    - Identify the job by ID and project.
    - Find the job YAML file in any of the status directories.
    - Delete the YAML file.
    - Delete the corresponding logs directory.
- Register the route in `api/main.go`: `DELETE /projects/{project}/jobs/{id}`.

### 2. Frontend: Add Delete Link to UI
- Update `api/templates/jobs.html`.
- In the `job-status` template (or `job-actions`), add a "delete" link if the job status is `crash`, `timeout`, or `stopped`.
- Use HTMX to call the delete endpoint.
- Upon successful deletion, remove the job row from the UI.

## Verification Plan

### Automated Tests
- Since there are no existing automated tests for deleting jobs, I'll focus on manual verification and ensuring no regressions.

### Manual Verification
1.  Submit a job that is guaranteed to crash (e.g., a prompt that results in a non-zero exit code if I can force it, or just wait for one).
2.  Once the job is in `crash` status, verify that a "delete" link appears.
3.  Click the "delete" link.
4.  Verify that the job row disappears from the UI.
5.  Verify that the YAML file and logs directory are gone from the filesystem.

status: done
