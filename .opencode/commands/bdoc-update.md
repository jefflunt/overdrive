---
description: Update project documentation to reflect the current state of the codebase
agent: build
---

# Update Project Documentation

You are an expert at maintaining high-quality, lazy-loadable documentation designed specifically for AI agents. Your goal is to ensure `build_docs/` accurately reflects the codebase's current capabilities, architecture, and tooling.

## Phase 1: Assessment
1. **Audit Documentation**: Use the `Task` tool with the `explore` subagent (thoroughness: `medium`) to scan the codebase and compare it against the current contents of `build_docs/` (starting with `README.md`).
2. **Identify Gaps**: Create a list of missing documentation for:
   - High-level features not yet in `README.md`.
   - Modules, subsystems, or tools that lack detailed documentation.
   - Outdated information that no longer matches the implementation.

## Phase 2: Implementation
1. **Update README**: Ensure `build_docs/README.md` captures all high-level features, provides a clear map of the system, and mentions the error reporting directory.
3. **Create/Update Sub-files**:
   - For each significant module, subsystem, or tool, create or update a dedicated markdown file (e.g., `build_docs/auth-subsystem.md`, `build_docs/ci-tooling.md`).
   - These files should explain the component's purpose, design, and integration in depth.
3. **Lazy-Loading Structure**:
   - Update `build_docs/README.md` to include references/links to these sub-files.
   - **Crucial**: Design the documentation so that an agent can understand the "big picture" from the README and only read specific sub-files when it needs to work on that particular part of the system.

## Phase 3: Organization and Optimization
1. **Refactor**: You are encouraged to re-organize, simplify, or restructure `build_docs/` to keep it clean and intuitive.
2. **Deprecate**: Remove or archive documentation for features or modules that no longer exist.
3. **AI Optimization**: Ensure the documentation is optimized for AI ingestion—concise, well-structured, and avoiding redundancy that would unnecessarily fill context windows.

Inform the user once the documentation has been updated and synchronized with the current state of the codebase.
