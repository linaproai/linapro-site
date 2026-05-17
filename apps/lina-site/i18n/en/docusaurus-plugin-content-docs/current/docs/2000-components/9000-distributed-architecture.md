---
slug: '/docs/distributed-architecture'
title: 'Native Distributed Architecture'
hide_title: true
description: 'How LinaPro natively supports distributed deployment — single-node vs. cluster mode, Redis-based distributed coordination and leader election, node roles and horizontal scaling, permission topology version synchronization, distributed locks and cluster-aware key-value cache, and how to enable cluster mode with zero code changes.'
keywords:
  - distributed architecture
  - cluster mode
  - leader election
  - horizontal scaling
  - high availability
  - LinaPro
  - distributed locks
  - key-value cache
  - permission topology sync
  - node election
  - cluster-aware
  - single-node deployment
  - cluster deployment
  - lease
  - primary node
  - replica node
  - Redis coordinator
---

## Overview

`LinaPro` treats distributed capability as a native feature of the framework — not as a bolt-on extension added after the fact. The framework supports both single-node and distributed deployment modes at the infrastructure level. Switching between them requires only a configuration change; no changes to business code are needed.

## Deployment Modes

### Single-node mode (default)

By default, `LinaPro` runs in single-node mode, suitable for development environments and small-scale production deployments. Single-node mode does not require `Redis` — it relies only on `PostgreSQL` and in-process cache coordination:

```yaml
cluster:
  enabled: false  # default — single-node mode
```

```text
          ┌──────────────────────────────────┐
          │             lina-core            │
          │    (single node, all-in-one)     │
          └─────────────────┬────────────────┘
                            │
                       ┌────▼─────┐
                       │PostgreSQL│
                       └──────────┘
```

### Cluster mode

Setting `cluster.enabled` to `true` activates the distributed coordination mechanism and enables horizontal scaling across multiple nodes. **Cluster mode requires a distributed coordinator** — the current version only supports `redis`:

```yaml
cluster:
  enabled: true
  coordination: redis        # Required — currently only redis is supported
  election:
    lease: 30s               # Election lock lease duration
    renewInterval: 10s       # Lease renewal interval (recommended: 1/3 of lease)
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: ""
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

```text
       ┌───────────────────────────────────┐
       │           Load Balancer           │
       └───────┬──────────────┬────────────┘
               │              │
         ┌─────▼─────┐  ┌─────▼─────┐
         │   Node 1  │  │   Node 2  │  ...
         │ (primary) │  │ (replica) │
         └─────┬─────┘  └─────┬─────┘
               │    Redis     │
               └──────┬───────┘
                      │
              ┌───────▼──────┐
              │    Redis     │  ← Cluster coordination (election, caching, distributed locks)
              └──────────────┘
                      │
               ┌──────▼──────┐
               │ PostgreSQL  │  ← Data persistence
               └─────────────┘
```

In cluster mode, all nodes share the same `PostgreSQL` database for data storage, and coordinate through `Redis` for distributed election, distributed locks, cluster-aware caching, and more.

## Distributed Leader Election

Cluster mode uses `Redis` for leader election — lightweight and reliable:

```mermaid
sequenceDiagram
    participant N1 as Node 1
    participant N2 as Node 2
    participant R as Redis

    N1->>R: Attempt to acquire election lock (SET NX)
    R-->>N1: Success, becomes primary
    N2->>R: Attempt to acquire election lock
    R-->>N2: Failed (lock held)
    Note over N2: Becomes replica, awaits primary failure

    loop Lease renewal (every renewInterval)
        N1->>R: Renew election lock (EXPIRE refreshes lease)
        R-->>N1: Renewal successful
    end

    Note over N1: Node 1 goes down, stops renewing

    N2->>R: Detects lease expired beyond lease duration
    N2->>R: Attempt to acquire election lock
    R-->>N2: Success, becomes new primary
```

**Lease configuration recommendations:**

- `lease` (lease duration): 30s is recommended — after the primary goes down, re-election completes within at most 30 seconds
- `renewInterval` (renewal interval): 1/3 of `lease` is recommended (i.e., 10s) to ensure a comfortable renewal window

## Primary and Replica Responsibilities

`LinaPro` primary and replica nodes have distinct responsibilities:

**Primary-only responsibilities:**

- Execute host-level periodic maintenance tasks
- Permission topology version broadcast (notifying all nodes to refresh their cache)
- Dynamic plugin lifecycle coordination

**Shared responsibilities (all nodes):**

- Handle `HTTP` requests (business API, plugin API)
- Read the permission topology (from local cache or database)
- Execute tasks in the persistent scheduled task subsystem (competing for a lock before execution to prevent duplicate runs)

## Permission Topology Version Synchronization

Changes to the permission topology (user roles, role menus, menu permissions) must propagate to all nodes. `LinaPro` achieves this through a version number mechanism:

```mermaid
sequenceDiagram
    participant Admin as Admin
    participant N1 as Node 1 (Primary)
    participant N2 as Node 2 (Replica)
    participant DB as PostgreSQL
    participant R as Redis

    Admin->>N1: Modify role permissions
    N1->>DB: Update permission data, increment version
    N1->>R: Publish cache version change
    Note over N2: Observes version change
    N2->>DB: Reload permission topology
    Note over N2: Local cache updated<br/>New requests use new permissions immediately
```

This mechanism ensures permission changes take effect across all nodes quickly — within at most 3 seconds — with no user re-login required.

## Distributed Task Scheduling

In cluster mode, the persistent scheduled task subsystem uses distributed locks to prevent duplicate task execution:

- Each time a task fires, nodes compete to acquire a distributed lock
- The node that acquires the lock executes the task; all other nodes skip
- After the task completes, the lock is released and the result is written to the database
- All nodes can view the task execution history

## Distributed Locks

The framework exposes a unified distributed lock interface (`pkg/locker`) for use by both the host and plugins:

```go
// Example: using a distributed lock to protect a critical section
err := locker.TryLock(ctx, "my-lock-key", 30*time.Second, func() error {
    // Critical section — only one node in the cluster will execute this
    return doSomething()
})
```

In single-node mode, distributed locks degrade to local mutexes with identical behavior.

## Key-Value Cache

The framework provides a cluster-aware key-value cache interface (`pkg/kvcache`) that supports cache invalidation broadcasts in cluster mode:

- Permission topology version cache
- Online session cache
- i18n resource cache

## Scaling Up

When business growth calls for more capacity, the steps are:

1. Update `config.yaml` — set `cluster.enabled` to `true`
2. Configure `cluster.coordination: redis` and a reachable `cluster.redis` endpoint
3. Start a new host node instance pointing to the same `PostgreSQL` database
4. Add the new node to the load balancer

No changes to business code are required. The framework automatically handles node discovery, leader election, and permission synchronization.
