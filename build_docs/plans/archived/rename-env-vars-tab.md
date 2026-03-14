# Plan: Rename "env vars" tab to "env"

Rename the "ENV VARS" tab in the project configuration modal to "ENV".

## Status
status: done

## Proposed Changes

### UI Changes
- In `api/templates/layout.html`, change the button text "ENV VARS" to "ENV".

## Verification Plan

### Automated Tests
- None applicable for this simple UI change.

### Manual Verification
1. Open the project configuration modal (Edit Project).
2. Verify the tab previously labeled "ENV VARS" is now labeled "ENV".
3. Verify the tab still functions correctly (shows environment variables).
