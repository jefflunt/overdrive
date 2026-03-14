---
title: Autofocus Prompt Textarea on Jobs Page
status: done
type: feature
---

# Feature Overview
*   **Goal**: Automatically give keyboard focus to the prompt textarea when the jobs page is loaded or when a job is duplicated.
*   **User Story**: "As a user, I want the prompt textarea to be focused automatically so that I can start typing my prompt immediately without clicking."
*   **Out of Scope**: Focusing other input elements like search or project creation modals (unless they are part of the jobs page prompt flow).

# Architecture & Design
*   **Modified Files**:
    *   `api/templates/jobs.html`: This file contains the jobs page template, including the prompt textarea and the associated JavaScript.

# Step-by-Step Implementation Plan

## Phase 1: Implementation

### Step 1.1: Add autofocus attribute to textarea
Add the `autofocus` attribute to the `textarea#prompt` element in `api/templates/jobs.html`. This provides the initial focus on page load in most browsers.

*   **File**: `api/templates/jobs.html`
*   **Verification**: Open the jobs page in a browser and check if the prompt textarea is focused.

### Step 1.2: Add explicit focus in DOMContentLoaded listener
In the `DOMContentLoaded` event listener in `api/templates/jobs.html`, add an explicit call to `.focus()` on the prompt element. This ensures focus is applied after the initial requirement check.

*   **File**: `api/templates/jobs.html`
*   **Verification**: Refresh the jobs page and verify the cursor is in the prompt textarea.

### Step 1.3: Add focus in duplicateJob function
In the `duplicateJob` function in `api/templates/jobs.html`, add a call to `.focus()` on the prompt element after setting its value. This improves UX when a user clicks "Redo" on an existing job.

*   **File**: `api/templates/jobs.html`
*   **Verification**: Click the "Redo" button on a job and verify the prompt textarea is focused and contains the duplicated prompt.

## Phase 2: Verification

### Step 2.1: Final Verification
Verify both scenarios (initial load and redo duplication) work as expected.

*   **Verification**:
    1.  Navigate to a project's jobs page (e.g., `/projects/demo/jobs`). Verify prompt is focused.
    2.  Click "Redo" on any job in the list. Verify prompt is focused and page scrolled to top.

# Critical Thinking & Edge Cases
*   **Mobile Devices**: On some mobile browsers, autofocus might not trigger the virtual keyboard. This is standard browser behavior to prevent UX issues, and we should not try to force it against browser policies.
*   **Disabled State**: The `updatePromptRequirement` function may disable the prompt if the job type is "docs". If it's disabled, `.focus()` will have no effect, which is correct.
*   **HTMX Swaps**: If the page content is swapped via HTMX (e.g., searching or filtering), the prompt form remains at the top and is not typically part of the swapped `#jobs-container`. If it were swapped, we would need to handle `htmx:afterSwap`.

# Final Comprehensive Verification Plan
1.  **Manual Walkthrough**:
    *   Open the dashboard and click on a project.
    *   Confirm the cursor is blinking in the "Type a prompt..." textarea.
    *   Scroll down to a previous job and click "Redo".
    *   Confirm the page scrolls up and the cursor is in the textarea with the duplicated text.
2.  **Edge Case Check**:
    *   Switch "Job Type" to "DOCS". The textarea should become disabled and lose focus (or at least not be focusable).
    *   Switch back to "FEATURE". The textarea should be enabled. (Optional: we could also focus it when switching back to "FEATURE", but it's not strictly requested).
