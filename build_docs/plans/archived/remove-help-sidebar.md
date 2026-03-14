---
status: done
---

# Remove Help Sidebar Plan

## Goal
Remove the help-specific navigation sidebar from `/api/templates/help.html` and adjust the layout so that the main content area occupies the full width.

## Proposed Changes

### 1. Frontend Template (`api/templates/help.html`)
- Remove the `<aside>` element with `id="doc-list"`.
- Remove the `flex` and `flex-row` classes from the outer container to allow the main content to occupy the full width.
- Ensure the `<main>` element and its container are styled correctly for a single-column layout.

## Verification
- Visit `/help` to see the landing page without the sidebar.
- Click on the cards (Jobs, Plans, Todos) to verify that the documentation pages also render without the sidebar and occupy the full width.
- Ensure all content is still accessible and correctly styled.
