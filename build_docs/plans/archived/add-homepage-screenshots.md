---
title: Add App Screenshots to Homepage
status: done
type: feature
---

# Feature: Add App Screenshots to Homepage

## 1. Feature Overview
*   **Goal**: Automatically capture screenshots of the running application (specifically the Settings page as a proxy for the app interface) in both desktop and mobile formats, and display them on the homepage (landing page).
*   **User Story**: "As a potential user visiting the homepage, I want to see actual screenshots of the application interface so that I understand what the tool looks like before I sign up."
*   **Out of Scope**: Creating a full demo environment with fake data (we will use the existing `/settings` page for the screenshot).

## 2. Architecture & Design

### New Files
*   `scripts/take_screenshots.go`: A Go script using `chromedp` to capture screenshots of the running app.
*   `static/img/app-desktop.png`: The generated desktop screenshot.
*   `static/img/app-mobile.png`: The generated mobile screenshot.

### Modified Files
*   `go.mod` / `go.sum`: Add `github.com/chromedp/chromedp` dependency.
*   `api/templates/home.html`: Add the HTML markup to display the screenshots in the Hero section.

### Visual Design
*   **Location**: In the Hero section, below the CTA buttons.
*   **Layout**: A large desktop screenshot with a smaller mobile screenshot overlapping it (e.g., bottom-right corner).
*   **Styling**: Rounded corners, shadow-2xl, border to separate from background.

## 3. Step-by-Step Implementation Plan

### Phase 1: dependencies & Scripting

1.  **Step 1.1: Add `chromedp` Dependency**
    *   Run `go get github.com/chromedp/chromedp` to add the dependency.
    *   Run `go mod tidy` to ensure `go.sum` is updated.
    *   **Verification**: `cat go.mod | grep chromedp`

2.  **Step 1.2: Create Screenshot Script**
    *   Create `scripts/take_screenshots.go`.
    *   **Logic**:
        *   Connect to `http://localhost:3281/settings` (assuming the app is running on default port).
        *   **Desktop**: Set viewport to 1920x1080. Capture screenshot. Save to `static/img/app-desktop.png`.
        *   **Mobile**: Set viewport to 375x812 (iPhone X dimensions). Capture screenshot. Save to `static/img/app-mobile.png`.
        *   Handle errors (e.g., if app is not running).
    *   **Verification**: Run `go run scripts/take_screenshots.go`. (Note: This requires the app to be running in a separate terminal. Since we are in a single-stream agent, we will try to start the app in the background or ask the user to ensure it's running. *Self-correction*: The agent can run `go run api/main.go &` then run the script, then `kill` the app).

### Phase 2: Frontend Implementation

3.  **Step 2.1: Update Homepage Template**
    *   Edit `api/templates/home.html`.
    *   Find the Hero section (inside the `text-center space-y-8` block).
    *   Append the screenshot container below the buttons.
    *   **Markup**:
        ```html
        <div class="mt-16 relative max-w-5xl mx-auto">
            <!-- Desktop Frame -->
            <div class="relative rounded-xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800 z-10">
                <img src="/static/img/app-desktop.png" alt="Desktop Interface" class="w-full h-auto bg-slate-100 dark:bg-slate-900">
            </div>
            <!-- Mobile Frame -->
            <div class="absolute -right-4 -bottom-8 w-[25%] max-w-[240px] rounded-3xl overflow-hidden shadow-2xl border-4 border-slate-900 z-20">
                <img src="/static/img/app-mobile.png" alt="Mobile Interface" class="w-full h-auto bg-slate-100 dark:bg-slate-900">
            </div>
        </div>
        ```
    *   **Verification**: Manual inspection of code.

4.  **Step 2.2: Verify and Polish**
    *   Start the app: `go run api/main.go &`
    *   Run the screenshot script: `go run scripts/take_screenshots.go`
    *   Kill the app.
    *   Check if images exist: `ls -l static/img/`
    *   (Optional) If the images are empty/broken, replace with a placeholder logic or retry.

## 4. Critical Thinking & Edge Cases
*   **App not running**: The script should fail gracefully or try to start the app. We will assume the user (or the agent) manages the app process for this task.
*   **Headless Environment**: `chromedp` should work in headless mode by default.
*   **No "Settings" page**: We verified `HandleSettings` exists.
*   **Git Ignore**: Should we commit these screenshots? Yes, usually assets are committed.

## 5. Final Comprehensive Verification
1.  Run `go mod tidy`.
2.  Start app in background.
3.  Run `go run scripts/take_screenshots.go`.
4.  Verify `static/img/app-desktop.png` and `static/img/app-mobile.png` exist.
5.  Read `api/templates/home.html` to confirm markup is present.
