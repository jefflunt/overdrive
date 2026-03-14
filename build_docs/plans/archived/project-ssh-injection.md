---
title: Project SSH Key & Config Injection
status: todo
type: feature
---

## 1. Feature Overview
*   **Goal**: Enhance project configuration to store SSH keys and configs in the project directory (`projects/<name>/`) instead of the root or `project.yml`, and update the worker build process to inject these specific credentials into the worker image on demand.
*   **User Story**: "As a platform engineer, I want projects to have their own isolated SSH keys and configurations stored within their project directory, so that the worker image can be built with the correct credentials for each specific project without manual file management in the root."
*   **Out of Scope**: Changing the runtime behavior of the worker *inside* the container (beyond receiving the files). Managing secrets via a vault (we are using file-based injection for now).

## 2. Architecture & Design

### New Files
*   `projects/test-project/project.yml` (for verification)
*   `projects/test-project/ssh.key` (dummy key for verification)
*   `projects/test-project/ssh_config` (dummy config for verification)

### Modified Files
*   `api/project.go`: Update `Project` struct and `GetProject` to resolve `ssh.key` and `ssh_config` from the project directory.
*   `api/worker.go`: Update `processJob` to pass `SSH_CONFIG` path environment variable to the worker script (it already passes `SSH_KEY`).
*   `scripts/work`: Major refactor to accept `SSH_KEY` and `SSH_CONFIG` paths, create a temporary build context, copy dependencies, and build a project-specific Docker image.
*   `projects/overdrive/project.yml`: Remove the `ssh_key` field (or make it optional/ignored).

### Data Models
*   **Project Struct (`api/project.go`)**:
    *   `SSHKey` (string): Changed semantics to be the full absolute path to the key file.
    *   Add `SSHConfig` (string): Full absolute path to the ssh_config file (empty if not present).

### API/Interface Changes
*   **Environment Variables** passed to `scripts/work` (via `api/worker.go` -> `step.Env`):
    *   `SSH_KEY`: Path to the project's SSH key.
    *   `SSH_CONFIG`: Path to the project's SSH config.

## 3. Step-by-Step Implementation Plan

### Phase 1: API & Project Logic Updates

1.  **Phase 1.1: Update Project Struct and Loading Logic**
    *   **File**: `api/project.go`
    *   **Description**:
        *   Add `SSHConfig` string field to `Project` struct.
        *   In `GetProject`, after unmarshaling:
            *   Check if `projects/<name>/ssh.key` exists. If so, set `project.SSHKey` to the absolute path of that file.
            *   Check if `projects/<name>/ssh_config` exists. If so, set `project.SSHConfig` to the absolute path of that file.
            *   Ensure `filepath.Abs` is used.
    *   **Verification**: Create a test go file `api/project_test_manual.go` that calls `GetProject("overdrive")` and prints the `SSHKey` and `SSHConfig` paths. Run with `go run api/project_test_manual.go api/project.go ...`.

2.  **Phase 1.2: Pass SSH Config to Worker Environment**
    *   **File**: `api/worker.go`
    *   **Description**:
        *   In `processJob`, inside the `envMap` definition:
            *   Update `SSH_KEY` to use `project.SSHKey`.
            *   Add `SSH_CONFIG` mapped to `project.SSHConfig`.
    *   **Verification**: Inspect `api/worker.go` to ensure map is updated. (Functional verification will happen in Phase 2).

### Phase 2: Worker Script & Build Process

3.  **Phase 2.1: Refactor `scripts/work` for Dynamic Build Context**
    *   **File**: `scripts/work`
    *   **Description**:
        *   Read `SSH_KEY` and `SSH_CONFIG` from environment variables.
        *   Validate they exist (if provided).
        *   Create a temporary directory `BUILD_CTX=$(mktemp -d)`. Setup `trap` to remove it on exit.
        *   Copy the following to `BUILD_CTX`:
            *   `.opencode/` (recursive)
            *   `auth.json` (if exists)
            *   `imgs/` (recursive, to get Dockerfile and entrypoint)
        *   **Crucial Step**: Copy the *specific* key and config to the names expected by Dockerfile:
            *   `cp "$SSH_KEY" "$BUILD_CTX/ssh.key"`
            *   `cp "$SSH_CONFIG" "$BUILD_CTX/ssh_config"` (if `$SSH_CONFIG` is set, else touch empty file).
        *   Change build command:
            *   Tag: `worker-$PROJECT_NAME` (sanitize project name to be valid docker tag).
            *   File: `-f "$BUILD_CTX/imgs/worker"`
            *   Context: `"$BUILD_CTX"`
        *   Change run command to use the new tag `worker-$PROJECT_NAME`.
    *   **Verification**: Run `export SSH_KEY=projects/overdrive/ssh.key PROJECT_NAME=overdrive; ./scripts/work` (mocking other env vars like JOB_ID). Verify it builds and runs (it might fail inside container due to missing job args, but build should succeed).

### Phase 3: Migration & Verification

4.  **Phase 3.1: Clean up `overdrive` Project**
    *   **File**: `projects/overdrive/project.yml`
    *   **Description**: Remove the `ssh_key: ssh.key` line to verify we are relying on the convention-based file discovery.
    *   **Verification**: Run `go run ./api list` (if CLI exists) or check logs to ensure project still loads.

5.  **Phase 3.2: End-to-End Test**
    *   **Description**:
        *   Create `projects/test-ssh/` directory.
        *   Create `projects/test-ssh/project.yml` (minimal).
        *   Create dummy `projects/test-ssh/ssh.key` and `projects/test-ssh/ssh_config`.
        *   Create a dummy job `projects/test-ssh/jobs/pending/test-job.yml` that runs a command `cat /ssh.key && cat /root/.ssh/config`.
        *   Start the server (or run the scheduler logic manually if possible).
        *   Wait for job to complete.
        *   Check logs in `projects/test-ssh/logs/test-job/worker.log`.
    *   **Verification**: The log should show the contents of the dummy key and config, proving injection worked.

## 4. Critical Thinking & Edge Cases
*   **Missing Keys**: If `projects/<name>/ssh.key` is missing, `scripts/work` will fail to copy it.
    *   *Mitigation*: In `scripts/work`, check if source file exists. If not, maybe create a dummy empty file to satisfy Dockerfile `COPY`? Or fail fast. The requirement implies these are necessary. I'll make it fail fast with a clear error message.
*   **Concurrency**: Unique image tags (`worker-$PROJECT_NAME`) prevent race conditions between projects. `mktemp -d` prevents race conditions during build context creation.
*   **Performance**: Copying `.opencode` (6MB) per job is acceptable but not instant. `podman build` caching should still work for the layers *before* the `COPY ssh.key` instruction, assuming we copy `ssh.key` late in the Dockerfile.
    *   *Note*: The current Dockerfile copies `ssh.key` *before* `imgs/worker-entrypoint.sh`. We might want to optimize Dockerfile order later, but for now we stick to the plan.
*   **Security**: The temp build context contains the private key. `trap ... rm -rf` is essential.

## 5. Final Comprehensive Verification Plan
1.  **Setup**:
    *   `mkdir -p projects/test-verify`
    *   `echo "dummy-key-content" > projects/test-verify/ssh.key`
    *   `echo "dummy-config-content" > projects/test-verify/ssh_config`
    *   Write `projects/test-verify/project.yml`:
        ```yaml
        name: test-verify
        repo_url: https://github.com/test/verify
        workflow: default.yml
        ```
    *   Write `projects/test-verify/jobs/pending/job-1.yml`:
        ```yaml
        id: job-1
        request:
          workflow_name: default.yml
          repo_url: https://github.com/test/verify
        ```
    *   Write `workflows/default.yml` (if not exists, or ensure it uses `scripts/work`).
2.  **Execution**:
    *   Run the server: `go run api/*.go`
    *   Wait for job to be picked up (moved to working, then done).
3.  **Verification**:
    *   `cat projects/test-verify/logs/job-1/worker.log`
    *   Search for "dummy-key-content" (requires the job to print it, see Phase 3.2).
    *   Actually, for the automated test, I will modify the job to print specific markers.
