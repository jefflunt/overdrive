---
status: done
---

# Help Page Update Plan

## Goal
Revamp the `/help` page to serve as a user-friendly support hub, linking to specific help articles for Jobs, Plans, and Todos.

## Proposed Changes

### 1. Backend (`api/handlers.go`)
- Modify `HandleHelp` function.
- Currently, it defaults to the first found document if no specific document is requested.
- Change this behavior: If no `doc` query parameter (and no implicit mapping from `from`), `selectedDoc` should be `nil`.
- This allows the template to distinguish between "viewing a doc" and "viewing the help home".

### 2. Frontend Template (`api/templates/help.html`)
- Update the main content area logic.
- **If `.Doc` is nil (Help Home):**
    - **Welcome Message:** "Welcome to the Help Center".
    - **Question Box:** A large text area or input for "Ask a question" (utilizing the existing AI chat script if possible, or just a placeholder for now as per request).
    - **Navigation Cards:** Three prominent cards/links:
        1.  **Jobs**: Links to `/help?doc=JOBS.md`.
        2.  **Plans**: Links to `/help?doc=PLANS.md`.
        3.  **Todos**: Links to `/help?doc=TODOS.md`.
- **If `.Doc` is present:**
    - Display the document content (existing behavior).
    - Ensure there is a "Back to Help Home" link or button, if the markdown file doesn't provide one (though the requirements say the markdown files should have it).

### 3. Documentation Content (`build_docs/`)
**(Completed)** Created three new markdown files with a helpful, user-centric tone.

#### `build_docs/JOBS.md`
- **Link:** `[← Back to Help Home](/help)` at the top.
- **Overview:** Explain what "Jobs" are in this system (tasks, builds, operations).
- **Specifics:**
    - How to view jobs.
    - Understanding job statuses (pending, working, done, etc.).
    - How to stop or retry jobs.
    - Viewing logs.

#### `build_docs/PLANS.md`
- **Link:** `[← Back to Help Home](/help)` at the top.
- **Overview:** Explain the "Plan" system (blueprints for changes).
- **Specifics:**
    - How plans are created.
    - How plans relate to jobs.
    - Viewing and executing plans.

#### `build_docs/TODOS.md`
- **Link:** `[← Back to Help Home](/help)` at the top.
- **Overview:** Explain the "Todo" system (task tracking).
- **Specifics:**
    - Creating todos.
    - Marking todos as complete.
    - Integration with external providers (if any known, otherwise general).

## Verification
- Visit `/help` to see the new landing page.
- Click the links to verify they open the correct markdown files.
- Verify the "Back" links work.
- Verify the question box is present.
