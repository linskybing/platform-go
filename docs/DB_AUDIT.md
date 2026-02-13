# Database and Dead-Code Audit

Last updated: 2026-02-13

This document records initial findings for DB schema and dead-code cleanup work.

## Findings (Preliminary)

- Course workloads: removed unused `internal/domain/course` models; drop `course_workloads` manually if it exists in DB.
- Workflow domain: `internal/domain/workflow` exists but is empty. Confirm whether workflow support will use this domain or remove the folder to avoid confusion.

## Next Steps

- Drop `course_workloads` table manually if present (AutoMigrate will not remove it).
- Scan for repositories without consumers and unused API routes after the handler refactor.
- Document a rollback-safe migration plan for any table removals.
