---
title: Remove placeholder text from todo description field
status: done
type: bugfix
---

## Root Cause Analysis
The description field in the todo list uses a `contenteditable` div. Instead of using a real CSS-based placeholder, it hardcodes "Add description..." as the default content when the description is empty. Additionally, when new todos are created, they are initialized with default description text like "Describe the feature...". Because these are actual text content, they do not disappear when the user starts typing.

## Evidence
In `static/todos.js`:
- Line 156-157:
  ```javascript
  placeholder="Add description...">
  ${todo.description || 'Add description...'}
  ```
- Line 191: `description: 'Describe the feature...',` in `addRootTodo()`
- Line 200: `description: 'Describe the subtask...',` in `addChild()`

## Fix Plan

### Step 1: Remove placeholder attribute and fallback text
In `static/todos.js`, locate the `todo-description` div within the `createTodoNode` method. Remove the `placeholder` attribute and change the fallback value for `${todo.description}` to an empty string.

**File:** `static/todos.js`

```javascript
// Old
<div class="todo-description ... " 
    contenteditable="${isEditable}" 
    onblur="tree.updateTodo('${todo.id}', 'description', this.innerText)"
    placeholder="Add description...">
    ${todo.description || 'Add description...'}
</div>

// New
<div class="todo-description ... " 
    contenteditable="${isEditable}" 
    onblur="tree.updateTodo('${todo.id}', 'description', this.innerText)">
    ${todo.description || ''}
</div>
```

### Step 2: Update default values for new todos
In `static/todos.js`, update `addRootTodo` and `addChild` to use an empty string for the initial description.

**File:** `static/todos.js`

- In `addRootTodo()`, change `description: 'Describe the feature...'` to `description: ''`.
- In `addChild()`, change `description: 'Describe the subtask...'` to `description: ''`.

## Verification
1. Verify that "Add description..." no longer appears in the rendered HTML for todos without descriptions.
2. Verify that new todos are created with an empty description field.
3. Confirm that the `todo-description` div still has height when empty due to existing CSS rules in `static/todos.css`.
