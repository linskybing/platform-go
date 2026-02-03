# Harbor Sync Tools

Utilities for synchronizing Harbor container registry with PostgreSQL database.

## Table of Contents

1. [Overview](#overview)
2. [Components](#components)
3. [Environment Variables](#environment-variables)
4. [Usage](#usage)
5. [Configuration](#configuration)
6. [Deployment](#deployment)
7. [Notes & Next Steps](#notes--next-steps)

---

## Overview

Tools for keeping Harbor container registry state synchronized with PostgreSQL `images` table.

Supports two approaches:
- Event-driven: Webhook receiver for real-time updates
- Periodic reconciliation: Batch verification of image state

## Components

### webhook_receiver.py

Small Flask app that accepts Harbor webhook push events.

- Listens for Harbor webhook push events
- Marks matching rows in Postgres `images(repository, tag)` with `is_pulled = true`
- Updates database on successful push notifications

### reconcile_harbor.py

One-shot script that pulls artifact lists from Harbor.

- Pulls artifact lists from Harbor API
- Updates `images.is_pulled` to reflect existence in Harbor (true/false)
- Fixes drift between Harbor and database

---

## Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `DATABASE_URL` | Postgres DSN | `postgres://user:pass@host:port/dbname` |
| `HARBOR_API` | Harbor API base URL | `https://harbor.example.com/api/v2.0` |
| `HARBOR_USER` | Harbor API username | `admin` |
| `HARBOR_PASS` | Harbor API password | `secret` |
| `HARBOR_PROJECT` | (Optional) Limit reconcile to one project | `my-project` |

---

## Usage

### Step 1: Install Dependencies

```bash
pip install -r tools/harbor_sync/requirements.txt
```

### Step 2: Run Webhook Receiver

```bash
export DATABASE_URL='postgres://user:pass@db:5432/mydb'
python tools/harbor_sync/webhook_receiver.py
```

Configure Harbor project webhook to POST to `http://<host>:8080/webhook`.

### Step 3: Run Manual Reconcile

```bash
export HARBOR_API=https://harbor.example.com/api/v2.0
export HARBOR_USER=admin
export HARBOR_PASS=secret
export DATABASE_URL='postgres://user:pass@db:5432/mydb'
python tools/harbor_sync/reconcile_harbor.py
```

---

## Configuration

### Database Schema

Adjust SQL and table/column names to match your schema. The examples assume:

```sql
CREATE TABLE images (
    repository TEXT NOT NULL,
    tag TEXT NOT NULL,
    is_pulled BOOLEAN DEFAULT false,
    CONSTRAINT images_pk UNIQUE (repository, tag)
);
```

### Flask Configuration (webhook_receiver.py)

For production use:

1. Run Flask app behind a WSGI server (gunicorn)
2. Secure the endpoint (token verification, TLS)
3. Configure proper logging
4. Add error handling and retries

---

## Deployment

### As Kubernetes CronJob

Deploy `reconcile_harbor.py` as a K8s CronJob to regularly fix drift:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: harbor-sync
spec:
  schedule: "0 * * * *"  # Every hour
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: reconcile
            image: python:3.9
            env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: harbor-sync
                  key: database-url
            - name: HARBOR_API
              value: https://harbor.example.com/api/v2.0
            - name: HARBOR_USER
              valueFrom:
                secretKeyRef:
                  name: harbor-sync
                  key: harbor-user
            - name: HARBOR_PASS
              valueFrom:
                secretKeyRef:
                  name: harbor-sync
                  key: harbor-pass
          restartPolicy: OnFailure
```

---

## Notes & Next Steps

### Key Considerations

- Table schema must match expectations in scripts
- Ensure Postgres and Harbor are accessible from webhook receiver
- Use environment variables for sensitive credentials
- Monitor webhook failures in application logs

### Recommended Improvements

1. **Token Verification** - Verify Harbor webhook signatures
2. **TLS/HTTPS** - Secure webhook endpoint with certificates
3. **Error Handling** - Implement retry logic for failed updates
4. **Logging** - Add comprehensive logging for debugging
5. **Metrics** - Track sync success/failure rates
6. **Batch Updates** - Optimize database updates for many images

### Testing

```bash
# Test webhook receiver locally
python -m pytest tools/harbor_sync/tests/

# Test Harbor API connectivity
curl -u admin:secret https://harbor.example.com/api/v2.0/projects

# Dry run reconcile (view changes without applying)
python tools/harbor_sync/reconcile_harbor.py --dry-run
```
