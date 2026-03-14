# Plan - Compact Todos on Dashboard

Status: done

Show todos in a more compact form on the dashboard "all projects" view.

## User Review Required

> [!IMPORTANT]
> I will make todos compact for BOTH "Starred items" and "All Projects" on the dashboard, as the goal is to save space on the dashboard generally.

- Does "abbreviated" mean I should truncate the description to a single line? (I'll assume yes, using `truncate` or `line-clamp-1`).

## Proposed Changes

### Dashboard View

#### [static/dashboard.js](static/dashboard.js)
- Update `createProjectGroup` to accept `isCompact` parameter.
- Update `createTodoNode` to accept `isCompact` parameter.
- Implement compact styling in `createTodoNode` when `isCompact` is true:
    - Reduce padding and margins.
    - Use a more horizontal layout for title, status, and description.
    - Truncate description to a single line.
- Update `render` to pass `isCompact: true` when calling `createProjectGroup`.

### Verification Plan

#### Automated Tests
- I'll check if there are any existing tests for the dashboard. (Doesn't seem like there are frontend tests I can run easily, but I can check backend handlers).
- Run `go test ./api/...` to ensure no regressions in dashboard data API.

#### Manual Verification
- Since I cannot see the UI, I will verify the generated HTML structure by examining the code and potentially adding a small test script to "render" and check the output.
- I'll verify that `static/todos.js` (the project-specific todos page) remains unchanged.
