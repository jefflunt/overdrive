---
description: Initialize the build_docs directory and project README for AI agents
agent: build
---

# Initialize Project Documentation

You are an expert at exploring codebases and distilling their essence into structured documentation for other AI agents.

## Phase 1: Pre-check
1. **Check Existence**:
   - Check if the `build_docs/` directory exists.
   - Check if `build_docs/README.md` exists.
2. **Early Exit**: If both the directory and the README already exist, exit early and inform the user that there is nothing to initialize.

## Phase 2: Creation and Research
If initialization is required:
1. **Prepare Directories**: Create the following directories if they do not exist:
   - `build_docs/`
   - `build_docs/plans/`
   - `build_docs/errors/`
2. **Deep Scan**:
    - **Purpose & Features**: Scan the codebase to identify the project's purpose and existing features.
    - **Roadmap**: Scan for any existing plan files (in the root or `build_docs/plans/`) to determine the pending roadmap.
    - **Verification**: Scan automated tests to confirm your understanding of the project's logic and behavior.
    - **Structure**: Analyze code layout, folder organization, and established design patterns.
    - **Tech Stack**: Identify the languages, frameworks, linters, and test runners being used.

## Phase 3: Generate Agent README
Write `build_docs/README.md` with the following sections:
1. **Overview**: Project purpose and high-level description.
2. **Current Features**: A summary of what the application can do today.
3. **Roadmap**: A summary of as-yet-not-implemented features or fixes found in existing plans.
4. **Architecture & Design**: Code layout, structure, and key design patterns to follow.
5. **Tooling**: Tech stack, linters, and how to run tests.
6. **Error Reporting**: Briefly mention that errors are tracked in `build_docs/errors/`.

**Standard Compliance**: Ensure the project follows Build Docs guidelines.

**Goal**: This file should serve as a bootstrap for future AI agents to understand the project context without being exhaustive. Focus on high-value information that allows for lazy-loading of more specific details later.

When finished, inform the user that the project is now ready for the `/bdoc-feature` command.
