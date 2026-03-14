---
title: Display N/A Duration for Todo Jobs
status: done
type: feature
---

# Feature: Display N/A Duration for Todo Jobs

## 1. Feature Overview
*   **Goal**: Display "n/a" for the duration of jobs that are in the "todo" (pending) status.
*   **User Story**: "As a user, I want to see 'n/a' for the duration of pending jobs so that I know the duration is not applicable yet, rather than seeing a blank space."
*   **Out of Scope**: Changing duration display for other statuses.

## 2. Architecture & Design
*   **Modified Files**:
    *   `api/job.go`: Update `Duration()` method.
    *   `api/job_test.go`: Add unit test case.
*   **API/Interface Changes**: `Job.Duration()` will return "n/a" instead of "" for pending jobs.

## 3. Step-by-Step Implementation Plan

### Phase 1: Implementation & Testing

1.  **Step 1.1: Update Job Duration Logic**
    *   **Description**: In `api/job.go`, modify the `Duration()` method. Check if `j.Status == "pending"`. If so, return "n/a". This should be done before checking if `StartedAt` is nil (though `StartedAt` should be nil for pending jobs anyway).
    *   **File**: `api/job.go`
    *   **Verification**: Create a temporary test or wait for the next step to run the full suite.

2.  **Step 1.2: Update Unit Tests**
    *   **Description**: In `api/job_test.go`, add a new test case to `TestJob_Duration` for a job with `status: "pending"`, `startedAt: nil`, and `completedAt: nil`. The expected result is `"n/a"`.
    *   **File**: `api/job_test.go`
    *   **Verification**: Run `go test -v api/job.go api/job_test.go api/ansi.go api/handlers.go api/main.go api/project.go api/worker.go` (or simply `go test ./api/...` if inside the module).

## 4. Critical Thinking & Edge Cases
*   **Edge Case**: What if a pending job somehow has a `StartedAt` time set?
    *   *Decision*: The requirement says "When a job is in the 'todo' status". So we should prioritize the status check. If it's "pending", return "n/a" regardless of timestamps.
*   **Edge Case**: "todo" vs "pending".
    *   *Context*: The UI displays "todo" for the "pending" status. The backend uses "pending". We must check for "pending".

## 5. Final Comprehensive Verification Plan
1.  **Automated Suite**: Run `go test -v ./api/...` and ensure all tests pass, including the new one.
