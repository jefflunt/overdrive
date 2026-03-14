# build_docs (or bdoc) a self-documenting tool for continuous improvement of a codebase

Build Docs (bdoc) is a set of AI-native commands and agents designed to create a standardized, machine-readable interface between your codebase and AI agents. It ensures that documentation, plans, and technical context are easily discoverable and actionable.

## Approach: The `build_docs/` Directory
The `bdoc` standard dictates that all project-level documentation and implementation plans reside in a `build_docs/` directory at the project root. This keeps your root clean while providing a "Source of Truth" for AI agents.

### Key Components
- **`README.md`**: High-level overview of the project, architecture, and tech stack.
- **`plans/`**: A directory containing detailed implementation plans for features and bug fixes.

---

## Getting Started
If you are new to `bdoc`, the best way to start is by initializing your project.

1.  **Initialize**: Run `/bdoc-init`. This scans your codebase and creates the `build_docs/` directory with a project README.
2.  **Plan**: Use `/bdoc-feature` (for new features) or `/bdoc-bug` (for bug fixes). These agents will research the codebase and create a detailed implementation plan.
3.  **Implement**: Once a plan is created, run `/bdoc-engineer <path-to-plan>` to execute the steps and verify the changes.
4.  **Sync**: Run `/bdoc-update` after making changes to ensure the AI-facing documentation stays up-to-date.

---

## Available Commands

| Command | Category | Description |
| :--- | :--- | :--- |
| `/bdoc-init` | `docs` | Initializes the `build_docs/` structure and scans the codebase. |
| `/bdoc-read` | `docs` | Provides a high-level overview of the project from `build_docs/README.md`. |
| `/bdoc-update` | `docs` | Syncs documentation with the current state of the code. |
| `/bdoc-bug` | `plan` | Research and plan a bug fix (creates a plan file). |
| `/bdoc-feature` | `plan` | Research and plan a new feature (creates a plan file). |
| `/bdoc-trace` | `plan` | Traces data flow and provides debug info for a bug report. |
| `/bdoc-merge` | `build` | Automatically merge a topic branch into a destination branch, resolving conflicts if necessary. |
| `/bdoc-engineer` | `build` | Executes an implementation plan and verifies changes. |
| `/bdoc-quick` | `build` | Rapidly implements small code changes based on feedback. |
| `/bdoc-idea` | `brainstorm` | Brainstorm new projects, features, or concepts. |
| `/bdoc-next` | `prioritize` | Scans pending plans and recommends what to work on next. |

---

## Specialized Agents

- **`bdoc_feature`**: **The Architect.** Deep-dives into context and produces "spoon-fed" implementation plans.
- **`bdoc_bug`**: **The Investigator.** Reproduces issues and specifies exact technical fixes.
- **`bdoc_engineer`**: **The Builder.** Follows plans, writes code, and runs tests to ensure everything works.
- **`bdoc_quick`**: **The Sprinter.** Quickly implements small, focused changes from feedback.
- **`bdoc_trace`**: **The Detective.** Traces data flow and provides deep technical context for bugs.
- **`bdoc_merge`**: **The Mediator.** Handles complex merges and automatically resolves conflicts.
- **`bdoc_idea`**: **The Partner.** A professional sounding board for brainstorming and refining concepts.

---

## Installation

To use these tools in your project, add this repository as a submodule:

```bash
git submodule add <repo-url> .opencode
```

Alternatively, you can clone the repository directly. Most AI agents will automatically discover the commands and agents within the directory.
