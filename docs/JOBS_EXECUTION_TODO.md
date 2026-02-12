# Jobs and Workflow Execution TODO

This document captures the backend work needed to deliver Argo-style job/workflow submission and tracking.

## API Surface

- POST /api/jobs/submit
  - Accepts: project_id, config_file_id (template), submit_type (job/workflow), optional overrides.
  - Validates: user access to project/group, template belongs to project.
  - Creates: job record and dispatches to executor or CRD client.
- GET /api/jobs/templates
  - Returns config files that include Job or Workflow resources (optional filter by project_id).
- GET /api/jobs
  - Already exists. Ensure it returns fields used by frontend (status, namespace, submitted_at).
- GET /api/jobs/:id
  - Already exists. Ensure it resolves to job/workflow status from DB + K8s.
- POST /api/jobs/:id/cancel
  - Already exists. Implement for local executor (delete Job/Workflow CRD).

## Execution Logic

- Add a Job/Workflow submit handler that routes to executor:
  - Local executor: create K8s Job or Workflow CRD via k8s client.
  - Scheduler executor: call external scheduler API (flash-sched) and update DB status.
- Extend executor SubmitRequest to include submit_type and template overrides.
- Add type-safe validation for resource overrides (cpu/mem/gpu, command/args, env, mounts).

## CRD + K8s Integration

- Create a Workflow CRD client (if using Argo or another workflow CRD).
- Implement status reconciliation:
  - Watch Job/Workflow resources and update job status in DB.
  - Store last known status and timestamps.

## Storage/Namespace Rules

- Namespace for submissions: proj-{PID}-{username}.
- Only bind group/user storage referenced in config resources (already handled in deploy helpers).

## Observability

- Emit audit logs for submit/cancel actions.
- Add metrics for submission latency and failure rate.

## Frontend Contract Notes

- The jobs page submits templates via config file instance creation for now.
- When /api/jobs/submit is ready, switch the UI to call it and return a job ID.
