---
title: Operations Flows
description: Detailed ops flows for K8s provisioning, caching, monitoring and rollout.
---

# Operations Flows

## O1. Kubernetes Provisioning & Rollout

1. Prepare image via CI and push to registry.
2. Update `k8s/` manifest or Helm chart with new image tag.
3. Apply via `kubectl apply -f k8s/` or via CI deploy step.
4. Monitor readiness via `/ready` and `/health`; if readiness fails, CI/ops triggers rollback.
5. Use rolling update strategy (see `k8s/` deployment spec `rollingUpdate` settings).

Reference files:
- `k8s/` manifests
- CI scripts in `.github/scripts`

## O2. Redis Cache Refresh & Invalidation

1. Services read using `pkg/cache/GetOrFetchJSON` with TTL.
2. On write operations (create/update/delete), service calls `cache.Delete` for affected keys.
3. For periodic refresh, cron jobs under `cron/` can warm caches.

Reference files:
- `pkg/cache/get_or_fetch.go`, `pkg/cache/operations.go`, `cron/`

## O3. Monitoring & Alerts

1. Instrument handlers with Prometheus metrics for latency and error counts.
2. Expose metrics endpoint and integrate with Prometheus scrape config.
3. Alert rules: p99 latency > 1s, error rate > 1%, CPU > 80%.

Reference files:
- monitoring examples in `pkg` or `internal` logging/metrics helpers
