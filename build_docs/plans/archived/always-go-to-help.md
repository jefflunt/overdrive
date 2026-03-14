---
status: done
---

# Always Go to /help Plan

## Goal
Ensure that clicking the help ("?") icon always navigates to `/help`, regardless of the current page.

## Proposed Changes

### 1. Frontend Template (`api/templates/layout.html`)
- Update the help icon link.
- Change `href="/help?from={{.CurrentPath}}"` to `href="/help"`.

## Verification
- Navigate to various pages (e.g., Home, Jobs, Todos).
- Click the help ("?") icon in the sidebar/header.
- Verify that it always navigates to `/help` (and not `/help?from=...`).
