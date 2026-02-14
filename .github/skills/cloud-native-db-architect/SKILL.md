---
name: cloud-native-db-architect
description: Guides agents in designing multi-tenant, high-concurrency PostgreSQL and Kubernetes architectures. Use when users ask for database schemas involving hierarchy optimization, versioned configuration storage, polymorphic relations, scheduling time constraints, or preemptive task queues.
metadata:
  version: "1.0"
  author: Sky
---

# Cloud-Native Database Architect Skill

This skill provides architectural patterns for complex Kubernetes and PostgreSQL integrations. When responding to architecture requests, apply the following design principles to ensure high concurrency, referential integrity, and optimal scheduling.

## Topological Hierarchy & RBAC

Structure: Design Group, Project, and SubProject entities.

Implementation: Avoid traditional adjacency lists for deeply nested structures. Employ parent_id for referential integrity and the PostgreSQL ltree extension for materialized paths to ensure fast ancestor and descendant queries.

Optimization: Create GiST indexes on the ltree path column to optimize topological lookups.

## Git-like Configuration Storage

Structure: Avoid storing entire YAML/JSON files repeatedly upon every change to prevent table bloat.

Implementation: Replicate Git's internal logic using a deduplicated immutable blob storage approach. Store configurations as JSONB in a blobs table (using a hash or UUID as the primary key). Use a commits table to track the history and link back to the blobs.

## Polymorphic Storage & K8s Mapping

Anti-pattern: Never use string-based polymorphic associations (e.g., owner_type and owner_id), as they cannot enforce data consistency on the database level using foreign keys.

Implementation: Use "Class Table Inheritance" (or Exclusive Foreign Keys). Create a base super-type table (e.g., parties) that user/group tables inherit from. The storage table then uses a single, strictly enforced foreign key (owner_party_id).

K8s Mapping: Store Kubernetes Node Affinity rules as JSONB in the database to map cross-namespace hostPath storage securely to specific nodes. This helps restrict workloads and prevent cross-tenant privilege escalation.

## Time-Window Constraints

Anti-pattern: Do not validate time overlaps in the application layer via loops.

Implementation: Use PostgreSQL range types (e.g., int4range mapping 0 to 604800 seconds to represent a recurring week). Implement EXCLUDE USING GIST constraints to strictly prevent overlapping schedules or double-booking at the transaction level. Handle wrap-around scenarios for cross-week schedules using CASE WHEN logic inside the exclusion constraint.

## Preemptive Task Queues
Implementation: Use PostgreSQL as a queue engine leveraging the FOR UPDATE SKIP LOCKED clause. This allows multiple workers to fetch pending tasks concurrently by skipping rows that are already locked by other transactions, completely eliminating lock contention.

Preemption Logic: Map Kubernetes PriorityClass values to a database table. Write SQL queries (using CTEs and Window Functions like SUM() OVER ()) to find and aggregate lower-priority running tasks ("victims") until sufficient resources (e.g., GPUs) are freed for high-priority tasks.