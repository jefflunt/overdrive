# Plan: Update Homepage Button Style

Update the homepage button in the top-left of the UI to have a thinner border and a grey color.

## Status
status: done

## Proposed Changes

### UI Components
- Modify `api/templates/layout.html`:
    - Update the homepage link (`<a>` tag with `href="/"`) in the sidebar.
    - Remove `bg-primary` and `dark:bg-primary-dark`.
    - Add `border border-slate-200 dark:border-border-dark`.
    - Ensure it retains `rounded` for rounded corners.
    - Keep `w-12 h-12 flex items-center justify-center`.

## Verification Plan

### Automated Tests
- No automated tests for this UI change.

### Manual Verification
- Open the UI and verify the homepage button:
    - Border is 1px thick.
    - Border color is grey (`#E2E8F0` / `#1A1A1A`).
    - Corners are rounded.
    - Icon is still centered and visible.
