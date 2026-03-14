---
title: Remove V2.0 Banner from Homepage
status: done
type: feature
---

# Feature Research & Planning

## Phase 1: Deep-Dive Research & Context Gathering

### 1. Environment & Documentation Scan
- **Project Structure**: Go backend with HTML templates (`api/templates/`).
- **Relevant File**: `api/templates/home.html` contains the "BUILDER V2.0 IS HERE" banner.
- **Rendering Logic**: `api/handlers.go` renders this template in `HandleListProjects`.

### 2. Codebase Exploration
- **Target**: The specific HTML block in `api/templates/home.html`:
    ```html
    <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary dark:text-primary-dark text-xs font-mono mb-4 mx-auto">
        <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
        </span>
        BUILDER V2.0 IS HERE
    </div>
    ```
- **Dependencies**: Uses Tailwind CSS classes (e.g., `animate-ping`, `bg-primary`). Removing the HTML is safe and won't break other parts of the UI as these are utility classes.

### 3. Constraint Analysis
- **Impact**: Purely frontend change. No backend logic needs modification.
- **Risk**: Minimal. Just ensuring valid HTML structure remains after removal.

## Phase 2: Detailed Technical Specification

### 1. Feature Overview
- **Goal**: Remove the "BUILDER V2.0 IS HERE" banner from the homepage hero section.
- **User Story**: As a user, I want a clean homepage without outdated announcements.
- **Out of Scope**: modifying any other part of the homepage.

### 2. Architecture & Design
- **New Files**: None.
- **Modified Files**: `api/templates/home.html`.
- **Data Models**: N/A.
- **API/Interface Changes**: N/A.

### 3. Step-by-Step Implementation Plan

#### Step 0: Safety & Setup
- Create a new branch `feature/remove-v2-banner`.

#### Step 1: Remove Banner HTML
- **Description**: In `api/templates/home.html`, remove the `<div>` block containing the "BUILDER V2.0 IS HERE" text. This block is located inside the `<section class="text-center space-y-8">` element, right before the `<h1>`.
- **File**: `api/templates/home.html`
- **Verification**: 
    - Run `grep "BUILDER V2.0 IS HERE" api/templates/home.html`. It should return no matches.
    - Run `grep "animate-ping" api/templates/home.html` to ensure no stray spans were left (unless used elsewhere, but this specific usage should be gone).

### 4. Critical Thinking & Edge Cases
- **Edge Case**: Ensure the spacing (`mb-4` on the removed div) doesn't cause the layout to look cramped. The parent container has `space-y-8`, so it should be fine, but we should verify the `h1` still looks good.
- **Performance**: Negligible improvement (less DOM).

### 5. Final Comprehensive Verification Plan
1.  **Automated Check**: Use `grep` to confirm the text is gone.
2.  **Manual Check**: (If possible) Open the homepage and visually confirm the banner is missing and the layout looks correct. Since this is an agent, we rely on the file content verification.
