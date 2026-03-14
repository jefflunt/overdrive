# Plan: Display Docs Job Prompt as "automatic documentation update"

Update the display logic for jobs to show "automatic documentation update" whenever a job is a "docs" job (i.e., its prompt starts with `/bdoc-update`).

## Tasks
1. [x] Locate the prompt formatting logic in `api/ansi.go`.
2. [x] Update `FormatPrompt` to return "automatic documentation update" for `/bdoc-update` jobs.
3. [x] Update tests in `api/ansi_test.go`.
4. [x] Verify the change in the UI (simulated by checking if all relevant places use `FormatPrompt`).

## Status
status: done
