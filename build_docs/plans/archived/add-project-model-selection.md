---
title: Add Model Selection to Project Configuration
status: done
type: feature
---

## 1. Feature Overview
*   **Goal**: Allow users to configure the default LLM model for each project via the "Add/Edit Project" modal.
*   **User Story**: "As a user, I want to select a specific Gemini model (e.g., `google/gemini-3-pro-preview`) for my project so that all automated jobs use that model by default."
*   **Out of Scope**: Changing models per-job in the UI (this is already supported in the backend `JobRequest` but not the focus here).

## 2. Architecture & Design
*   **Modified Files**:
    *   `api/project.go`: Add `Model` field to `Project` struct.
    *   `api/worker.go`: Update `EnqueueJob` to use `project.Model` as the default if the job request doesn't specify one.
    *   `api/templates/layout.html`: Add dropdown to the Create/Edit Project modal and update JS to handle the new field.
*   **Data Models**:
    *   Update `Project` struct:
        ```go
        type Project struct {
            // ... existing fields
            Model string `json:"model" yaml:"model"`
        }
        ```
*   **API/Interface Changes**:
    *   `POST /projects`: Accepts `model` field in JSON body.
    *   `GET /projects/{name}/config`: Returns `model` field in JSON response.

## 3. Step-by-Step Implementation Plan

### Phase 1: Backend Implementation

1.  **Step Name**: Phase 1.1: Update Project Struct
    *   **Description**: In `api/project.go`, add the `Model` field to the `Project` struct with JSON and YAML tags.
    *   **File(s)**: `api/project.go`
    *   **Verification**: Create a temporary Go test file `api/project_model_test.go` that creates a `Project` with a model, saves it using `SaveProject`, loads it back using `GetProject`, and asserts the `Model` field is preserved. Run `go test ./api/project_model_test.go`.

2.  **Step Name**: Phase 1.2: Update Job Enqueue Logic
    *   **Description**: In `api/worker.go`, modify the `EnqueueJob` function. Change the default model assignment logic to check `project.Model` first. If `project.Model` is empty, fall back to `"google/gemini-3-flash-preview"`.
    *   **File(s)**: `api/worker.go`
    *   **Verification**: Create a temporary Go test file `api/worker_model_test.go`. Test `EnqueueJob` with a project that has a custom model set, and verify the resulting job has that model. Test with a project that has NO model set, and verify it uses the default flash model. Run `go test ./api/worker_model_test.go`.

### Phase 2: Frontend Implementation

3.  **Step Name**: Phase 2.1: Update Project Modal UI
    *   **Description**: In `api/templates/layout.html`, inside the `createProjectForm` (id `createProjectModal`), add a new form group for the Model selection.
        *   Label: "DEFAULT MODEL"
        *   Element: `<select>` with id `p_model` and name `model`.
        *   Options:
            *   `google/gemini-3-flash-preview` (Default)
            *   `google/gemini-3-pro-preview`
    *   **File(s)**: `api/templates/layout.html`
    *   **Verification**: Open the browser, click "Add Project" (or "Create New Project"), and visually verify the dropdown exists with the correct options.

4.  **Step Name**: Phase 2.2: Update Frontend Logic
    *   **Description**: In `api/templates/layout.html`, update the JavaScript functions:
        *   `showCreateProjectModal()`: Reset the `p_model` select to the default value.
        *   `editProject(name)`: Populate `p_model` value from the fetched `project.model`.
        *   `document.getElementById('createProjectForm').onsubmit`: Include `model: document.getElementById('p_model').value` in the JSON payload sent to `/projects`.
    *   **File(s)**: `api/templates/layout.html`
    *   **Verification**:
        1.  Create a new project with "pro-preview" selected. Check `projects/<name>/project.yml` to see if `model: google/gemini-3-pro-preview` is saved.
        2.  Edit an existing project, change the model, save, and reopen the edit modal to verify the selection persists.

## 4. Critical Thinking & Edge Cases
*   **Backward Compatibility**: Existing projects won't have a `model` field in their YAML. The Go struct will default it to empty string `""`. The logic in `EnqueueJob` handles this by falling back to the hardcoded default.
*   **Invalid Models**: The dropdown restricts choices, but the API theoretically accepts any string. Since this is an internal tool, strict validation isn't critical, but we rely on the `EnqueueJob` logic to pass whatever string is there to the worker script.
*   **UI Defaults**: When creating a new project, we should select the "flash" model by default in the UI to encourage cost-efficiency/speed unless "pro" is needed.

## 5. Final Comprehensive Verification Plan
1.  **Full Flow Test**:
    *   Open the web UI.
    *   Create a new project "TestModelProject" with Repo URL "git@github.com:example/repo.git" and Model "google/gemini-3-pro-preview".
    *   Verify `projects/TestModelProject/project.yml` contains `model: google/gemini-3-pro-preview`.
    *   Trigger a job for this project (e.g., via the "Submit Job" feature if available, or manually trigger a build).
    *   Check the generated job YAML in `projects/TestModelProject/jobs/pending/` (or working/done). Verify `request.model` is `google/gemini-3-pro-preview`.
