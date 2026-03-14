---
title: Seamless Jobs Page Refresh
status: done
type: enhancement
---

# Feature Research & Planning

## 1. Feature Overview
*   **Goal**: Replace the full-page refresh on `GET /jobs` with HTMX-based partial updates to preserve user selection.
*   **User Story**: "As a user, I want the jobs list to update without a full page reload so that my text selections and scroll position are preserved."
*   **Method**: HTMX with `hx-get` and `hx-trigger="every 1s"`.

## 2. Implementation Plan
1.  **Modify `api/templates/jobs.html`**:
    *   Include HTMX script.
    *   Remove `<meta http-equiv="refresh" content="1">`.
    *   Wrap the job tables in a `div` with `id="job-list-container"`.
    *   Add HTMX attributes to `job-list-container` to poll `/jobs`.
2.  **Modify `api/handlers.go`**:
    *   Update `HandleListJobs` to check for `HX-Request` header.
    *   If `HX-Request` is present, render only the tables (or the container) instead of the full page.
    *   Actually, to keep it simple, we can just render the same template and use HTMX's `hx-select` to pick the container. However, for efficiency and better experience, a partial template is better.

## 3. Step-by-Step Implementation

### Step 1: Prepare the Template
*   Identify the parts of `api/templates/jobs.html` that need to be refreshed.
*   Add HTMX CDN script.
*   Add `hx-` attributes.

### Step 2: Update the Handler
*   Modify `HandleListJobs` to handle HTMX requests if we want to be efficient.
*   Alternatively, just let HTMX pick the div from the full response.

## 4. Verification Plan
*   **Manual Verification**:
    1.  Open `/jobs` in a browser.
    2.  Select some text in the page.
    3.  Wait for the refresh (every 1-2 seconds).
    4.  Verify the selection is still there.
    5.  Verify the job status still updates when a job progresses.
*   **Regression**:
    1.  Ensure "Submit New Job" link still works.
    2.  Ensure "View/Tail" links still work.
