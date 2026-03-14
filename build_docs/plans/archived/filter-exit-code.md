---
title: Filter Exit code column to only show numeric value
status: todo
type: feature
---

### 1. Feature Overview
* **Goal**: Show only the numeric exit code in the 'Exit code' column on the `/jobs` page.
* **User Story**: As a user, I want to see a clean numeric exit code without additional error messages cluttering the column.
* **Out of Scope**: Moving the error message to another column (unless requested).

### 2. Architecture & Design
* **Modified Files**:
    * `api/templates/jobs.html`: Remove the error message from the 'Exit code' column.

### 3. Step-by-Step Implementation Plan
#### Phase 1: Research & Preparation
1. **Step 1.1: Verify current behavior**
    * Check `api/templates/jobs.html` to confirm where the 'Exit code' column is rendered.
    * Already done: it's at line 83.

#### Phase 2: Implementation
2. **Step 2.1: Modify `api/templates/jobs.html` - Exit Code Column**
    * Remove `{{if .Error}}({{.Error}}){{end}}` from the `<td>` for the 'Exit code' column.
    * **File**: `api/templates/jobs.html`
    * **Verification**: `grep "ExitCode" api/templates/jobs.html`

3. **Step 2.2: Modify `api/templates/jobs.html` - Status Column Tooltip**
    * Add `title="{{.Error}}"` to the Status column `<td>` so the error message is still accessible on hover.
    * **File**: `api/templates/jobs.html`
    * **Verification**: `grep "status-{{.Status}}" api/templates/jobs.html`

### 4. Critical Thinking & Edge Cases
* **Missing Error Info**: By removing `.Error` from the main page, users might lose immediate visibility into *why* a job failed. However, they can still check the logs via the "View" or "Tail" links.
* **Nil Exit Code**: If `ExitCode` is nil, the column will be empty. This is consistent with current behavior (except for the error message part).
* **Process Exit Code 2**: Exit code 2 is "no-op". It will show "2" instead of "2 (no-op)" if that was the error message (though currently the error message for 2 is not set, see `api/worker.go` line 118).

### 5. Final Comprehensive Verification Plan
1. **Template Verification**: Use `grep` to ensure the `.Error` tag is no longer present in the "Exit Code" column of `api/templates/jobs.html`.
2. **Visual Verification (Optional)**: If you can run the server and trigger a failure, verify the page only shows the numeric exit code.
