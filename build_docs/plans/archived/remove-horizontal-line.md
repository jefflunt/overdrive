# Plan: Remove horizontal line in new job form

Remove the horizontal line (top border) in the new job form on the jobs page.

## Status
status: done

## Implementation
- Modify `api/templates/jobs.html` to remove the `border-t` and related classes from the div separating the prompt textarea and the submit button.

## Verification
- Visually verify that the horizontal line is gone.
- Ensure the form still looks good and is functional.
