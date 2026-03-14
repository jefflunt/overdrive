status: done

# Refactor Help Files Location

## Goal
Move user-facing help files from `build_docs/` to a new `help_docs/` directory to separate them from developer documentation and build artifacts. Update the application to serve help content from this new location.

## Changes

### 1. Create `help_docs` directory
- Create a new directory at the root of the repository.

### 2. Move/Copy User Documentation
- Copy the following files from `build_docs/` to `help_docs/`:
    - `JOBS.md`
    - `JOB_SYSTEM.md` (Referenced by help logic)
    - `API_SERVER.md` (Referenced by help logic)
    - `DEPLOYMENT.md` (Referenced by help logic)
    - `PLANS.md` (Referenced by help logic)
    - `TODOS.md` (Referenced by help logic)
    - `README.md` (General project info)
    - `CHANGELOG.md` (Version history)
    - `WORKER_INFRASTRUCTURE.md` (Technical details, linked in README)

### 3. Update Backend Logic
- Update `api/handlers.go`:
    - `HandleHelp`: Read files from `help_docs/*.md`.
    - `HandleHelpSearch`: Read files from `help_docs/*.md`.

## Verification
- Verify `help_docs` exists and contains the files.
- Verify `api/handlers.go` points to the new location.
- Verify the application builds successfully.
