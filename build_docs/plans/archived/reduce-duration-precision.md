---
title: Reduce Duration Precision on Jobs Page
status: todo
type: feature
---

## Feature Overview
* **Goal**: Show "Duration" in the Jobs list with only whole seconds/minutes/hours precision, instead of microsecond precision.
* **User Story**: "As a user, I want to see a clean duration for finished jobs so that I'm not distracted by unnecessary microsecond precision."
* **Out of Scope**: Changing how durations are stored or calculating durations for active jobs.

## Architecture & Design
* **Modified Files**:
    * `api/job.go`: Add a `Duration()` method to the `Job` struct.
    * `api/templates/jobs.html`: Use the new `Duration()` method for rendering.
* **New Files**:
    * `api/job_test.go`: Unit test to verify the `Duration()` method.

## Step-by-Step Implementation Plan

### Phase 1: Foundation & Logic

#### Step 1.1: Create unit test for Job.Duration()
Create `api/job_test.go` to test the rounding logic of the new method.
* **File**: `api/job_test.go`
* **Logic**: Test with different time differences (e.g., 5s, 1m5s, 1h2m3s) and ensure they round correctly to the second and return a clean string.
* **Verification**: Run `go test ./api/...` (it should fail to compile because `Duration()` doesn't exist yet).

#### Step 1.2: Implement Job.Duration() method
Add the `Duration()` method to the `Job` struct.
* **File**: `api/job.go`
* **Logic**:
    ```go
    func (j Job) Duration() string {
        if j.StartedAt == nil || j.CompletedAt == nil {
            return ""
        }
        return j.CompletedAt.Sub(*j.StartedAt).Round(time.Second).String()
    }
    ```
* **Verification**: Run `go test ./api/...` (it should pass now).

### Phase 2: Integration

#### Step 2.1: Update Jobs template
Update the template to use the new method.
* **File**: `api/templates/jobs.html`
* **Logic**: Replace the manual calculation `{{.CompletedAt.Sub .StartedAt}}` with `{{.Duration}}`.
* **Verification**: Start the server and view the `/jobs` page with some completed jobs.

## Critical Thinking & Edge Cases
* **Nil Timestamps**: The `Duration()` method handles cases where `StartedAt` or `CompletedAt` are nil by returning an empty string.
* **Negative Duration**: While unlikely in this system, `Round` handles negative durations correctly.
* **Zero Duration**: If a job is extremely fast, it will show as `0s`.

## Final Comprehensive Verification Plan
1. **Automated Suite**: Run `go test ./api/...` to ensure the logic is correct.
2. **Manual Walkthrough**:
    1. Start the server.
    2. Submit a new job via `/jobs/new`.
    3. Wait for the job to complete.
    4. Refresh `/jobs` and verify the "Duration" column shows a value like `12s` or `1m5s` instead of `12.345678s`.
