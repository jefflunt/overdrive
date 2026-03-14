# Plan: Toast Notification

Implementation of a global toast notification system.

## Status
status: done

## Tasks
1. [x] Add toast container to `api/templates/layout.html` inside the `<main>` element.
2. [x] Add `showToast` JavaScript function to `api/templates/layout.html`.
3. [x] Style the toast notifications using Tailwind CSS.
4. [x] Implement automatic disappearance of toasts.
5. [x] Support success, error, and info types.

## Verification Steps
1. Open the application.
2. Open the browser console.
3. Call `showToast('Test Success', 'success')`.
4. Call `showToast('Test Error', 'error')`.
5. Call `showToast('Test Info', 'info')`.
6. Verify that toasts appear at the top of the content area, have appropriate colors, and disappear after a few seconds.
