# Plan: Asynchronously Stop Containers on Scheduler Restart

## Implementation
1.  **Modify `api/handlers_settings.go`**:
    *   In `HandleRebuildRestartScheduler`, add a call to `podman stop --all` inside the goroutine before (or alongside) the scheduler restart script.
    *   Ensure the command is executed asynchronously and doesn't block the UI or the scheduler restart process unnecessarily.
    *   Use `Start()` instead of `Run()` for the podman command to ensure it's truly asynchronous if needed, although it's already in a goroutine. But `podman stop --all` can take time, so we should probably just let it run in the background.
    *   Wait, the instruction says "asynchronously send a command to podman to stop any running containers. Do not block the UI to wait for the containers to stop."
    *   The current `HandleRebuildRestartScheduler` already runs in a goroutine and returns immediately to the UI.
    *   If I add `podman stop --all` inside that goroutine, it won't block the UI.

2.  **Verification**:
    *   Check if the code compiles.
    *   Manually verify (if possible) that the command is triggered.
    *   Ensure no regressions in existing tests.

## Status
status: done
