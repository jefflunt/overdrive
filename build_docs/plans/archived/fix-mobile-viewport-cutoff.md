---
title: Fix Mobile Viewport Cutoff
status: completed
type: feature
---

# Feature Overview
*   **Goal**: Ensure the entire web dashboard is visible on mobile devices, preventing the bottom of the navigation bar (including the build number) from being cut off by browser address bars.
*   **User Story**: "As a mobile user, I want to see the entire dashboard without the bottom being obscured by the browser's address bar, so I can access all navigation items and see the build version."
*   **Out of Scope**: Major layout redesigns or changing the navigation structure.

# Architecture & Design
*   **Modified Files**:
    *   `api/templates/layout.html`: Update the `body` tag's CSS classes to use dynamic viewport height (`dvh`).
*   **CSS Changes**:
    *   Replace `h-screen` (which is `100vh`) with `h-screen h-[100dvh]` on the `body` element.
    *   Update the viewport meta tag to include `viewport-fit=cover`.

# Step-by-Step Implementation Plan

## Phase 1: Setup
1.  **Step 1.1**: Create a new branch.
    *   **Description**: Create a feature branch for the layout fix.
    *   **Action**: Run `git checkout -b fix/mobile-viewport-cutoff`.
    *   **Verification**: `git branch` shows the new branch.

## Phase 2: Layout Update
1.  **Step 2.1**: Update viewport meta tag in `api/templates/layout.html`.
    *   **Description**: Add `viewport-fit=cover` to ensure the app handles mobile notches and address bars correctly.
    *   **File**: `api/templates/layout.html`
    *   **Change**: Update `<meta content="width=device-width, initial-scale=1.0" name="viewport"/>` to `<meta content="width=device-width, initial-scale=1.0, viewport-fit=cover" name="viewport"/>`.
    *   **Verification**: Inspect the file content.

2.  **Step 2.2**: Update `body` height unit in `api/templates/layout.html`.
    *   **Description**: Replace the fixed `h-screen` (100vh) with dynamic viewport height `h-[100dvh]` to account for mobile browser UI elements. We keep `h-screen` as a fallback.
    *   **File**: `api/templates/layout.html`
    *   **Change**: Change `<body class="... h-screen flex overflow-hidden">` to `<body class="... h-screen h-[100dvh] flex overflow-hidden">`.
    *   **Verification**: Inspect the file content.

# Critical Thinking & Edge Cases
*   **Browser Support**: `dvh` is supported in all modern mobile browsers (iOS Safari 15.4+, Chrome 108+). Older browsers will fall back to `h-screen` (100vh), which is the current (slightly broken) behavior, so there is no regression.
*   **PWA Mode**: When installed as a PWA, the address bar is usually hidden. In this case, `100vh` and `100dvh` are equivalent, so the behavior remains consistent.

# Final Comprehensive Verification Plan
1.  **Code Inspection**:
    *   Verify `api/templates/layout.html` has the updated viewport meta tag.
    *   Verify `api/templates/layout.html` has `h-[100dvh]` on the `body` tag.
2.  **Manual Verification (Mobile)**:
    *   Open the application in a mobile browser (e.g., Safari on iOS or Chrome on Android).
    *   Verify that the sidebar (aside) extends exactly to the bottom of the visible area.
    *   Confirm the build number (e.g., `b123`) at the bottom of the sidebar is fully visible and not cut off by the address bar.
    *   Toggle the address bar (by scrolling or clicking the URL bar) and verify the layout adjusts correctly or remains stable (depending on the browser's `dvh` implementation).
