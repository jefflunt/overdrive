# Logging & Errors

The overdrive system captures execution logs for every job to facilitate debugging and auditing. Logs are isolated by project.

## Log Directory Structure

Each job execution creates a unique directory within the project's logs folder:

```
projects/
└── <project-name>/
    └── logs/
        └── <job-id>/
            └── worker.log  # Worker stdout/stderr
```

## Error Tracking

Errors are captured in the `worker.log` file. The Job System categorizes jobs into `jobs/crash/` if any workflow step fails (non-zero exit code). The job's YAML file contains the high-level error message.

### Debugging Workflow

1.  **Identify Failure**: Check the project dashboard for jobs in the "Crashed" state.
2.  **Inspect Logs**: Read the logs via the dashboard or directly from the filesystem.
3.  **Reproduce**: Use `scripts/work` with the same environment variables or duplicate the job via the UI to reproduce the failure.
