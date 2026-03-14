---
description: Scan for incomplete plans and recommend the next steps
agent: plan
---

# Next Plan Recommendation

You are an expert project manager and software engineer. Your task is to analyze existing implementation plans and recommend the next most critical items to build.

## Phase 1: Scan Plans
1. **Locate Plans**:
   - Scan for files in `build_docs/plans/*.md`.
   - Scan for files in the root matching `feature-plan-*.md`.
2. **Read and Analyze**:
   - Read each plan to determine its purpose (bug fix vs. feature), dependencies, and current status.
   - **Filter**: Only include plans where the status is `todo` or `in-development`. Ignore plans marked as `done`.

## Phase 2: Prioritize
1. **Evaluate Importance**:
   - **Bugs**: Prioritize based on criticality (security > data loss > broken workflows > UI glitches).
   - **Features**: Prioritize based on user value and strategic importance.
2. **Check Dependencies**:
   - Identify if any plan requires another plan to be completed first. Prioritize the prerequisite items.
3. **Select Top 5**: Select the 5 most important and actionable plans.

## Phase 3: Present to User
Present the top 5 plans in a prioritized list. For each item:
1. Provide the file path.
2. Provide a concise, 3-line summary to refresh the user's memory on the feature/bdoc-bug.
3. If there's a dependency, explicitly call out why this item must be done first.

**Call to Action**: 
- Ask the user which of these plans they would like to move to the `engineer` command (e.g., `/bdoc-engineer <path>`).
