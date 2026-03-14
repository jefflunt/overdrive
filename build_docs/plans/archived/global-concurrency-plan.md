# Plan: Make global concurrency values editable

status: done

Currently, the global concurrency limits on the "settings" page are read-only and cannot be changed through the UI. This plan outlines the steps to make them editable and persistent.

## User Review Required

> [!IMPORTANT]
> The global settings will be stored in `projects/global_settings.yml`. This directory is already used for project-specific configurations.

- Should we use a different location for global settings? (Currently assuming `projects/global_settings.yml` is fine as `projects/` directory already exists and contains project configs).

## Proposed Changes

### API Server

#### [api/worker.go]
- Change global concurrency variables to be updated by a new function.
- Add `LoadGlobalSettings()` function to be called at startup.

#### [api/settings.go] (New File)
- Define `GlobalSettings` struct.
- Implement `LoadGlobalSettings()` and `SaveGlobalSettings()` functions.
- Global settings will be saved to `projects/global_settings.yml`.

#### [api/handlers_settings.go]
- Add `HandleSaveGlobalSettings` to handle POST requests for global settings.
- Update `HandleSettings` to ensure it uses the latest global settings.

#### [main.go] (or wherever the server starts)
- Call `LoadGlobalSettings()` on startup.

### UI

#### [api/templates/settings.html]
- Remove `readonly` and `opacity-60` from global concurrency inputs.
- Update the JavaScript to handle saving global settings when `.Project.Name` is empty.
- Send the POST request to `/settings/global/save`.

## Verification Plan

### Automated Tests
- Create a test in `api/handlers_settings_test.go` (or a new test file) to verify that global settings can be saved and retrieved.
- Verify that changing global settings affects the concurrency checks in `api/worker.go`.

### Manual Verification
1. Navigate to the global settings page.
2. Change the global concurrency values.
3. Click "Save Global Settings".
4. Verify the "Settings saved successfully" toast appears.
5. Refresh the page and ensure the values persist.
6. Verify that the new limits are enforced when running jobs.
