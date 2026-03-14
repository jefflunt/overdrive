---
title: ANSI Escape Sequence Support for Logs
status: todo
type: feature
---

# Feature Overview
* **Goal**: Update the 'View' and 'Tail' pages to render ANSI escape sequences (colors) as HTML.
* **User Story**: "As a developer, I want to see colored logs in the web interface so that I can easily distinguish between different log levels."
* **Out of Scope**: Support for non-color ANSI sequences (e.g., cursor movement) except for ignoring them. Support for 256-color or TrueColor.

# Architecture & Design
* **New Files**:
    * `api/ansi.go`: Core logic for ANSI-to-HTML conversion.
    * `api/ansi_test.go`: Unit tests for the conversion logic.
* **Modified Files**:
    * `api/handlers.go`: Update `HandleViewLogs`, `HandleTailLogs`, and `LogPageData`.
    * `api/templates/tail_logs.html`: Update template to render HTML logs and include CSS.

## Data Models
* Update `LogPageData` in `api/handlers.go`:
  ```go
  type LogPageData struct {
      JobID string
      Logs  template.HTML // Changed from string
  }
  ```

# Step-by-Step Implementation Plan

## Phase 1: Foundation (ANSI Utility)

### Step 1: Create the ANSI to HTML conversion logic
* **Description**: Create `api/ansi.go` with a function `AnsiToHtml(text string) template.HTML`.
* **Logic**:
    1. Use `html.EscapeString(text)` to escape the input.
    2. Use `regexp.MustCompile("\x1b\\[([0-9;]*)([mK])")` to find ANSI sequences.
    3. Use a state machine to track `openSpans` (int).
    4. Process each match:
        - If the terminator is `K`, replace with empty string (ignore).
        - If the terminator is `m`, split the numeric part by `;`. If empty, treat as `0`.
        - For each numeric code:
            - `0`: Close all `openSpans` by appending `</span>` for each and set `openSpans = 0`.
            - `1`: Append `<span class="ansi-bold">` and `openSpans++`.
            - `30-37`: Append `<span class="ansi-fg-X">` (where X is 30-37) and `openSpans++`.
            - `90-97`: Append `<span class="ansi-fg-bright-X">` (where X is 90-97) and `openSpans++`.
            - Others: Ignore.
    5. After processing all matches, close any remaining `openSpans`.
    6. Return `template.HTML(result)`.
* **File**: `api/ansi.go`
* **Verification**: Run `go test ./api/ansi_test.go` (created in Step 2).

### Step 2: Create unit tests for ANSI conversion
* **Description**: Create `api/ansi_test.go` with these test cases:
    - `"Normal text"` -> `"Normal text"`
    - `"\x1b[31mRed text\x1b[0m"` -> `<span class="ansi-fg-31">Red text</span>`
    - `"\x1b[1;32mBold Green\x1b[0m"` -> `<span class="ansi-bold"><span class="ansi-fg-32">Bold Green</span></span>`
    - `"\x1b[31mRed \x1b[32mGreen\x1b[0m"` -> `<span class="ansi-fg-31">Red <span class="ansi-fg-32">Green</span></span>`
    - `"Mixed \x1b[Kignore \x1b[31mColor"` -> `Mixed ignore <span class="ansi-fg-31">Color</span>` (with final span closed)
* **File**: `api/ansi_test.go`
* **Verification**: `go test ./api/ansi_test.go` should pass.

## Phase 2: Integration

### Step 3: Update LogPageData and HandleTailLogs
* **Description**:
    - Update `LogPageData` struct in `api/handlers.go` to use `template.HTML` for `Logs`.
    - Update `HandleTailLogs` to call `AnsiToHtml` on the log data.
* **File**: `api/handlers.go`
* **Verification**: `go build ./api/...` should succeed.

### Step 4: Update Tail Logs Template
* **Description**: 
    - Add CSS styles for `ansi-` classes in the `<style>` block:
      ```css
      .ansi-bold { font-weight: bold; }
      .ansi-fg-30 { color: #000000; }
      .ansi-fg-31 { color: #cd3131; }
      .ansi-fg-32 { color: #0dbc79; }
      .ansi-fg-33 { color: #e5e510; }
      .ansi-fg-34 { color: #2472c8; }
      .ansi-fg-35 { color: #bc3fbc; }
      .ansi-fg-36 { color: #11a8cd; }
      .ansi-fg-37 { color: #e5e5e5; }
      .ansi-fg-bright-90 { color: #666666; }
      .ansi-fg-bright-91 { color: #f14c4c; }
      .ansi-fg-bright-92 { color: #23d18b; }
      .ansi-fg-bright-93 { color: #f5f543; }
      .ansi-fg-bright-94 { color: #3b8eea; }
      .ansi-fg-bright-95 { color: #d670d6; }
      .ansi-fg-bright-96 { color: #29b8db; }
      .ansi-fg-bright-97 { color: #e5e5e5; }
      ```
* **File**: `api/templates/tail_logs.html`
* **Verification**: Visual verification in later step.

### Step 5: Update HandleViewLogs
* **Description**:
    - Change `Content-Type` to `text/html`.
    - Wrap the converted logs in a full HTML page. Use the same CSS as in Step 4.
    - Example structure:
      ```go
      fmt.Fprintf(w, "<!DOCTYPE html><html><head><style>%s</style></head><body><pre>%s</pre></body></html>", css, convertedLogs)
      ```
* **File**: `api/handlers.go`
* **Verification**: `go build ./api/...` should succeed.

## Phase 3: Verification

### Step 6: Create Dummy Log File
* **Description**: Create a script to generate a log file with ANSI codes.
* **File**: `scripts/gen-ansi-logs.sh` (or just use `mkdir -p logs/test-job && printf "..." > logs/test-job/worker.log`)
* **Verification**: File exists and contains ANSI codes.

### Step 7: Manual Verification
* **Description**: 
    - Start the server: `go run api/*.go`.
    - Visit `http://localhost:3281/jobs/tail/test-job`.
    - Visit `http://localhost:3281/jobs/logs/test-job`.
    - Verify colors are rendered correctly.

# Critical Thinking & Edge Cases
- **Security**: Always HTML-escape the log content *before* inserting our own `<span>` tags to prevent XSS.
- **Large Logs**: The current implementation reads the entire log file into memory. This is fine for now as it matches existing behavior, but ANSI conversion adds some overhead.
- **Nested Spans**: The state machine must correctly close tags in the right order or just close all and reopen to keep it simple. Closing all and reopening is safer for ANSI reset.

# Final Comprehensive Verification Plan
1. **Automated**: Run `go test ./api/...` to verify the conversion logic.
2. **Manual**:
    - Run the dummy log generator.
    - Open the browser to the tail and view pages.
    - Check for:
        - Red, Green, Yellow colors.
        - Bold text.
        - Correct reset (text after `[0m` should be normal).
        - Background color of the page is dark.
        - No raw ANSI codes visible (e.g., `[31m`).
