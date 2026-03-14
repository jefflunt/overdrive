---
title: Fix Todo Expansion Toggle and Refactor to Idiomatic JS
status: done
type: bugfix
---

# Root Cause Analysis

## The Problem
On the todos page, clicking the expansion/collapse icon often fails to do anything.

## The "Why"
The investigation revealed an **event race condition** caused by the interaction between `contenteditable` elements and the global re-render strategy.

1.  **Event Interference**: The todo title and description are `contenteditable`. When they lose focus (e.g., when clicking an expansion button), they trigger an `onblur` event.
2.  **Disruptive Re-render**: The `onblur` handler calls `updateTodo`, which calls `load()` and then `render()`.
3.  **DOM Destruction**: The `render()` method clears `this.container.innerHTML` and rebuilds the entire DOM.
4.  **Race Condition**: If a user clicks the expansion button, the `mousedown` event happens, then the `blur` event fires. If the re-render happens immediately (which it does), the expansion button is destroyed before the `click` event (which happens on `mouseup`) can fire.
5.  **Global Variable Reliance**: The code uses `onclick="tree.toggleExpand(...)"`, which relies on a global `tree` instance, making it brittle and non-idiomatic.

# Evidence
- In `static/todos.js`, `toggleExpand` calls `this.render()`, which performs `this.container.innerHTML = ''`.
- The `onblur` event on titles/descriptions also triggers a sequence ending in `this.render()`.
- Other buttons like "Save" already use `onmousedown` with `event.preventDefault()` to mitigate focus loss, but the expansion button does not.

# Fix Plan

## Step 1: Codified Reproduction (Simulation)
Since we cannot run a browser environment, we will verify the logic via code analysis and then implement the fix. The fix will directly address the identified race condition and architectural debt.

## Step 2: Refactor `TodoTree` to use Event Delegation
Instead of embedding `onclick` attributes in HTML strings, we will use a single event listener on the container. This is more robust and avoids global variable reliance.

## Step 3: Implement Targeted Updates for Expansion
Modify `toggleExpand` to update only the necessary DOM elements (button rotation and children visibility) instead of triggering a full re-render.

## Step 4: Use `onmousedown` for Interactive Elements
Use `mousedown` listeners with `event.preventDefault()` for expansion and action buttons to prevent them from stealing focus and triggering the disruptive `onblur` re-render during a click.

## Step 5: Clean up `updateTodo` re-render
Modify `updateTodo` to avoid a full re-render if the change was already applied locally (e.g., during text editing).

# File Changes

### `static/todos.js`
- Update the constructor to attach event listeners to `this.container`.
- Remove `onclick` and other event attributes from the HTML generated in `createTodoNode`.
- Use `data-id` and `data-action` attributes to facilitate event delegation.
- Rewrite `toggleExpand` to perform targeted DOM updates.
- Refactor `updateTodo` to be more surgical.

### `api/templates/todos.html`
- Clean up the initialization to ensure it doesn't leak more than necessary (though `window.tree` is still useful for debugging, it won't be required for the UI to function).

# Verification
1.  **Toggle Test**: Verify that clicking the expansion icon toggles visibility of children.
2.  **Focus Test**: Verify that clicking expansion while editing a title works correctly and doesn't cause the UI to flicker or lose state.
3.  **CRUD Test**: Verify that adding, editing, and deleting todos still functions as expected.
4.  **Regression Test**: Ensure that "Submit to Job Queue" and other actions still work.

/bdoc-engineer static/todos.js static/todos.css
