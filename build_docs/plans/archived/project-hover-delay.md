---
title: Project Icon Hover Delay
status: done
type: feature
---

# Feature Research & Planning

## 1. Feature Overview
*   **Goal**: Keep the project action buttons (Edit/Delete) visible for 1.5 seconds after the mouse leaves the project icon area to prevent accidental disappearances.
*   **User Story**: "As a user, I want the project action buttons to stay visible for a moment after I move my mouse away so that I can easily click them without perfect mouse precision."
*   **Out of Scope**: Changes to other hover menus (e.g., job lists).

## 2. Architecture & Design
*   **Modified Files**: `api/templates/layout.html`
*   **Technical Approach**:
    *   Use CSS transitions to implement the delay.
    *   Use `transition-all` with a custom delay `delay-[1500ms]` on the default state (fade out).
    *   Use `group-hover:delay-0` on the hover state (instant fade in).
    *   Replace `pointer-events-none` with `invisible` to ensure buttons are clickable while visible but not while hidden.

### CSS Class Changes
Target element: The action buttons container in the project list loop.

**Current:**
```html
class="absolute ... opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none group-hover:pointer-events-auto z-30"
```

**New:**
```html
class="absolute ... opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-300 delay-[1500ms] group-hover:delay-0 z-30"
```

## 3. Step-by-Step Implementation Plan

### Phase 1: Implementation
**Step 1.1: Update Project List Item Styles**
*   **Description**: Modify `api/templates/layout.html` to add the delay classes to the project action buttons container.
*   **File**: `api/templates/layout.html`
*   **Logic**:
    1.  Locate the loop `{{range listProjects}}`.
    2.  Locate the action buttons container `div` inside the loop (around line 123).
    3.  Replace `transition-opacity` with `transition-all duration-300`.
    4.  Replace `pointer-events-none` with `invisible`.
    5.  Replace `group-hover:pointer-events-auto` with `group-hover:visible`.
    6.  Add `delay-[1500ms]` to the base classes.
    7.  Add `group-hover:delay-0` to the hover classes.
*   **Verification**:
    *   Run `grep "delay-\[1500ms\]" api/templates/layout.html` to confirm the class was added.
    *   Run the server locally (if possible) and check the hover behavior (manual verification step).

## 4. Critical Thinking & Edge Cases
*   **Clickability**: Ensuring buttons remain clickable during the fade-out is handled by keeping `visibility: visible` (implied by the transition delay) and removing `pointer-events-none`.
*   **Responsiveness**: On mobile/touch devices, hover states are tricky. However, the sidebar is primarily a desktop feature. On mobile, `group-hover` often requires a tap. The delay won't hurt mobile users.
*   **Tailwind Compatibility**: `delay-[1500ms]` requires JIT or a configured theme. The project uses CDN Tailwind which supports JIT values.

## 5. Final Comprehensive Verification Plan
1.  **Code Check**: Verify the HTML file contains the new classes.
2.  **Visual Check**: (Manual) Hover over a project icon, move the mouse away. The buttons should remain visible for 1.5 seconds before fading out. During this time, they should be clickable.
