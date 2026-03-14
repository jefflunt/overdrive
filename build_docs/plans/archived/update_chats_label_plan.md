# Plan: Update Chat Label

Change the "conversations" label to "chats" on the chats page and double its font size.

## Tasks
1. Update `api/templates/chat.html` to change "Conversations" to "chats" and update font size class from `text-xs` to `text-base` or `text-2xl`.
    - Note: `text-xs` is 0.75rem. `text-2xl` is 1.5rem. Doubling it leads to `text-2xl`.
2. Update other occurrences of "conversations" to "chats" for consistency if they are part of labels/placeholders.

## Verification
- Inspect `api/templates/chat.html` and verify the text and classes.
- If possible, check the rendered output (though I can't easily do that here, I can verify the code).

## Status
status: done
