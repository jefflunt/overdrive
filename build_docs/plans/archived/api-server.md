---
title: Background Job API & Worker System
status: todo
type: feature
---

# Feature Research & Planning

## 1. Feature Overview
*   **Goal**: Create a Go-based API server that accepts job requests, manages a filesystem-based job queue, and executes containerized workers with a concurrency limit.
*   **User Story**: "As a developer, I want to submit jobs via an API so that they are queued and executed by background workers without overloading the system."
*   **Out of Scope**:
    *   Authentication/Authorization (unless explicitly requested later).
    *   Database integration (using filesystem for persistence).
    *   Advanced job retry logic (basic error handling only).
    *   Websocket updates (polling only).

## 2. Architecture & Design

### New Files
*   `api/main.go`: Main entry point for the API server.
*   `api/handlers.go`: HTTP handlers for `/up`, `/jobs` (POST/GET).
*   `api/worker.go`: Background worker logic (scheduler, execution).
*   `api/job.go`: Job struct definitions and YAML handling.
*   `api/templates/jobs.html`: HTML template for the job listing page.
*   `scripts/work`: Shell script to execute the worker container (wraps `podman run`).

### Modified Files
*   `go.mod`: New file (init module).
*   `go.sum`: New file (dependencies).

### Data Models
**Job Structure (YAML in `jobs/<status>/<id>.yml`)**:
```yaml
id: "job-1234567890"
status: "pending" # pending, working, complete, error
created_at: "2023-10-27T10:00:00Z"
started_at: "2023-10-27T10:00:05Z" # Optional
completed_at: "2023-10-27T10:05:00Z" # Optional
request:
  repo_url: "https://github.com/..."
  branch_parent: "main"
  branch_child: "feature-branch"
  commit_msg: "feat: add feature"
  prompt: "implement feature X"
  model: "google/gemini-pro"
exit_code: 0 # Optional
```

### API Interface
*   `GET /up`: Returns 200 OK.
*   `POST /jobs`: Accepts JSON payload (maps to `request` field in YAML). Returns 201 Created with Job ID.
*   `GET /jobs`: Returns HTML page listing jobs.

## 3. Step-by-Step Implementation Plan

### Phase 1: Project Setup & Healthcheck
1.  **Step 1.1: Initialize Go Module**
    *   Run `go mod init overdrive` in the root directory.
    *   Run `go get gopkg.in/yaml.v3` (for YAML parsing).
    *   **Verification**: `cat go.mod` shows the module and dependency.

2.  **Step 1.2: Create API Skeleton**
    *   Create `api/` directory.
    *   Create `api/main.go`:
        *   Setup `http.ServeMux`.
        *   Add `GET /up` handler returning "OK".
        *   Start server on port 8080.
    *   **Verification**: Run `go run api/main.go` and `curl http://localhost:8080/up`.

### Phase 2: Job Submission & Queue Structure
3.  **Step 2.1: Create Job Struct & Directories**
    *   Create `api/job.go`: Define `Job` and `JobRequest` structs with YAML tags.
    *   Create `jobs/pending`, `jobs/working`, `jobs/complete`, `jobs/error` directories using `os.MkdirAll` in `main.go` init.
    *   **Verification**: Run `go run api/main.go` and check if directories are created.

4.  **Step 2.2: Implement `POST /jobs`**
    *   In `api/handlers.go`: Create `HandleSubmitJob`.
    *   Parse JSON body to `JobRequest`.
    *   Generate ID (e.g., `job-<timestamp>`).
    *   Create `Job` struct, set status="pending", created_at=now.
    *   Marshal to YAML and write to `jobs/pending/<id>.yml`.
    *   Return 201 Created with `{"id": "..."}`.
    *   **Verification**: `curl -X POST -d '{"repo_url":"...", "prompt":"test"}' http://localhost:8080/jobs`. Check if file exists in `jobs/pending`.

### Phase 3: The Worker Script (`scripts/work`)
5.  **Step 3.1: Create `scripts/work`**
    *   Create a bash script `scripts/work` that accepts environment variables: `REPO_URL`, `PRIMARY_BRANCH`, `NEW_BRANCH`, `COMMIT_MSG`, `PROMPT`, `MODEL`, `JOB_ID`.
    *   It should:
        *   Define log dir: `logs/$JOB_ID`.
        *   Run `podman run` (similar to `scripts/demo-local`) but use the passed env vars.
        *   Ensure the `imgs/worker` image is built (or build it quietly).
        *   Exit with the container's exit code.
    *   Make it executable: `chmod +x scripts/work`.
    *   **Verification**: Run `JOB_ID=test REPO_URL=... scripts/work` manually to see if it triggers the container. (Use a dummy prompt to save time/cost).

### Phase 4: Job Scheduler & Execution
6.  **Step 4.1: Implement Scheduler Loop**
    *   Create `api/worker.go`: `StartScheduler()`.
    *   Start a goroutine in `main.go` that runs every 1 second.
    *   Logic:
        *   Count files in `jobs/working` (using `filepath.Glob`).
        *   If count >= 64, continue.
        *   Read files from `jobs/pending` (sort by creation time, oldest first).
        *   If pending found:
            *   Move file from `jobs/pending` to `jobs/working`.
            *   Update status in YAML to "working" and add `started_at`.
            *   Launch `go processJob(filename)` (new goroutine).

7.  **Step 4.2: Implement Job Processor**
    *   In `api/worker.go`: `processJob(filename)`.
    *   Parse YAML to get request details.
    *   Construct `exec.Command("scripts/work")`.
    *   Set `Cmd.Env` with values from YAML request.
    *   Run command and wait.
    *   Capture exit code.
    *   Update YAML: `completed_at`, `exit_code`.
    *   Move file to `jobs/complete` (if exit 0) or `jobs/error` (if exit != 0).
    *   **Verification**: Submit a job via `POST /jobs`. Watch it move from pending -> working -> complete/error.

### Phase 5: Job Listing (`GET /jobs`)
8.  **Step 5.1: Create HTML Template**
    *   Create `api/templates/jobs.html`.
    *   Simple HTML with 3 sections: Pending, Working, Completed/Error.
    *   Use Go `html/template` syntax to iterate over lists.

9.  **Step 5.2: Implement `GET /jobs`**
    *   In `api/handlers.go`: `HandleListJobs`.
    *   Read all files from `jobs/pending` and `jobs/working`.
    *   Read all files from `jobs/complete` and `jobs/error`.
    *   Combine complete/error lists, sort by `completed_at` (descending), take top 50.
    *   Parse YAML for each file to get details (ID, status, prompt, time).
    *   Render template.
    *   **Verification**: Visit `http://localhost:8080/jobs` and verify the list matches the filesystem state.

### Phase 6: Robustness & Cleanup
10. **Step 6.1: Startup Cleanup**
    *   In `main.go`, before starting scheduler:
    *   Check `jobs/working`.
    *   If any files exist, move them to `jobs/error` (assume server crash = job failed).
    *   Update YAML with "Job interrupted by server restart".
    *   **Verification**: Manually put a file in `jobs/working`, start server, check if it moves to `jobs/error`.

## 4. Critical Thinking & Edge Cases
*   **Concurrency**: The scheduler checks file count. Since `processJob` is a goroutine, the "check" and "move" logic in the main loop must be synchronous to avoid race conditions where we spawn > 64 goroutines.
*   **File Locking**: Only one process (the scheduler) moves files from pending -> working. `processJob` moves working -> complete. This avoids conflicts.
*   **Large Output**: `scripts/work` redirects logs to `logs/$JOB_ID`. The API doesn't need to stream logs, just status.
*   **YAML Parsing Errors**: If a YAML file is corrupted, move it to `jobs/error` immediately to unblock the queue.

## 5. Final Verification Plan
1.  **End-to-End Test**:
    *   Start server: `go run api/main.go`.
    *   Submit 5 jobs via `curl`.
    *   Watch `GET /jobs` page auto-update (refresh manually).
    *   Verify `jobs/complete` has 5 files.
    *   Verify `logs/` has 5 log folders.
2.  **Concurrency Test**:
    *   Submit 70 dummy jobs (mock `scripts/work` to just sleep 1s).
    *   Verify `jobs/working` never exceeds 64 files.
    *   Verify all eventually finish.
