---
title: Add Merge Button next to Job ID
status: done
type: feature
---

# Feature Overview
* **Goal**: Add a "Merge" button next to the Job ID field in the dashboard that pre-fills the "Submit New Job" form for merging.
* **User Story**: "As a user, I want to quickly initiate a merge of a completed job's branch back into its parent branch."
* **Out of Scope**: Automating the actual merge process beyond pre-filling the form. The existing `merge` job type and agent handle the execution.

# Architecture & Design
* **Modified Files**:
    * `/overdrive/target/api/templates/jobs.html`: This file contains both the UI template and the client-side logic for pre-filling the job submission form.

# Step-by-Step Implementation Plan

## Step 0: Safety & Setup
1. **Create new branch**:
   * Command: `git checkout -b feature/merge-button`

## Phase 1: Client-side Logic
1. **Step 1.1: Add `mergeJob` JavaScript function**
   * **Description**: In `api/templates/jobs.html`, add a new function `mergeJob(btn)` inside the `<script>` tag. This function will be similar to `duplicateJob` but tailored for merging.
   * **Logic**:
     ```javascript
     function mergeJob(btn) {
         const data = btn.dataset;
         document.getElementById('repo_url').value = data.repoUrl;
         document.getElementById('branch_parent').value = data.branchParent;
         document.getElementById('job_type').value = 'merge';
         document.getElementById('prompt').value = `Merge branch ${data.id} into ${data.branchParent}`;
         document.getElementById('commit_msg').value = `Merge ${data.id} into ${data.branchParent}`;
         window.scrollTo({ top: 0, behavior: 'smooth' });
     }
     ```
   * **File**: `api/templates/jobs.html`
   * **Verification**: Verify the function is added by searching for it in the file.

## Phase 2: UI Changes
1. **Step 2.1: Add the Merge button to the `job-row` template**
   * **Description**: In `api/templates/jobs.html`, locate the `{{define "job-row"}}` block. Wrap `{{.ID}}` in a container that includes the new Merge button.
   * **Code to add**:
     ```html
     <td>
         <div style="display: flex; align-items: center; gap: 5px;">
             {{.ID}}
             <a href="javascript:void(0)" title="Merge" onclick="mergeJob(this)"
                data-id="{{.ID}}"
                data-repo-url="{{.Request.RepoURL}}"
                data-branch-parent="{{.Request.BranchParent}}"
                style="color: #009900; display: inline-flex;">
                 <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" fill="currentColor" viewBox="0 0 16 16">
                     <path fill-rule="evenodd" d="M10 15a2 2 0 1 0 0-4 2 2 0 0 0 0 4zm-8-2a2 2 0 1 0 0-4 2 2 0 0 0 0 4zm0-8a2 2 0 1 0 0-4 2 2 0 0 0 0 4zm9-1.939g-1.07.535 3 3 0 0 1 .438 1.878H12a1 1 0 0 1 1 1v6c0 .11-.018.217-.051.317l.74.37a2 2 0 0 1 .311 1.253V4.5a2 2 0 0 0-2-2h-.562zM4.5 3a1 1 0 0 1 1 1v6.793l1.146-1.147a.5.5 0 0 1 .708.708l-2 2a.5.5 0 0 1-.708 0l-2-2a.5.5 0 0 1 .708-.708L4 10.793V4a1 1 0 0 1 1-1z"/>
                 </svg>
             </a>
         </div>
     </td>
     ```
   * **File**: `api/templates/jobs.html`
   * **Verification**: Inspect the HTML of the job list in the dashboard to ensure the button is rendered.

# Critical Thinking & Edge Cases
* **Missing Data**: If a job somehow has no `RepoURL` or `BranchParent`, the button will still show but the pre-fill will be incomplete. Given the current validation, this is unlikely.
* **Job States**: The button is shown for all jobs. This is intentional as a user might want to prepare a merge job even before the previous job is fully finished (though it might fail if the branch isn't pushed yet).

# Final Comprehensive Verification Plan
1. **Open the Dashboard**: Navigate to the overdrive Jobs dashboard.
2. **Identify a Job**: Find an existing job in the list.
3. **Click Merge**: Click the green merge icon next to the Job ID.
4. **Verify Form**:
   * [ ] `Repo URL` matches the job's repo.
   * [ ] `Parent Branch` matches the job's original parent branch.
   * [ ] `Type` dropdown is set to `merge`.
   * [ ] `Prompt` is `Merge branch <ID> into <Parent>`.
   * [ ] `Commit Msg` is `Merge <ID> into <Parent>`.
5. **Submit**: Click "Submit" and verify a new job is created with these parameters.
