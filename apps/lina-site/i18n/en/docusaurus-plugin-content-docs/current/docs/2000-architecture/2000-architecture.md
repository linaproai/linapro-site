---
slug: '/docs/architecture'
title: 'Architecture Overview'
hide_title: true
description: 'A detailed breakdown of LinaPro architecture — the responsibilities, runtime relationships, and design principles of the five layers: core host service, management workspace, dual-mode plugin system, AI development workflow, and AI skill system. Understand the layered model, loose-coupling approach, and how each component collaborates to deliver business features.'
keywords:
  - LinaPro architecture
  - four-layer architecture
  - core host service
  - lina-core
  - lina-vben
  - plugin system
  - AI development workflow
  - AI skill system
  - architecture design
  - loose coupling
  - GoFrame
  - Vue3
  - layered design
  - plugin runtime
  - RBAC
  - design principles
  - extension points
---



## Architecture Overview

`LinaPro` uses a four-layer separation model. Each layer is independently designed around loose-coupling principles, and layers interact through well-defined interfaces rather than depending directly on each other's internal implementations. The key components are:

```mermaid
graph TB
    subgraph Workflow["AI Development Workflow  openspec/"]
        direction LR
        Explore["🔍 Explore"] --> Propose["📋 Propose"] --> Implement["⚙️ Implement"] --> Review["🔎 Review"] --> Archive["📦 Archive"]
    end

    subgraph Frontend["Management Workspace  lina-vben"]
        UI["Vue 3 + Vben5 + Ant Design"]
    end

    subgraph Host["Core Host  lina-core"]
        direction TB
        API["API Layer\n(g.Meta route definitions + DTO)"]
        Ctrl["Controller Layer\n(HTTP request handling)"]
        Svc["Service Layer\n(Core business logic)"]
        Plugin["Plugin Runtime\n(Lifecycle orchestration · Sandbox isolation)"]
        Gov["Governance\n(JWT · RBAC · Logs · Sessions)"]
        API --> Ctrl --> Svc
        Svc --> Plugin
        Svc --> Gov
    end

    subgraph Plugins["Plugin Layer  lina-plugins"]
        direction LR
        Source["Source Plugins\nCompiled with host"]
        Dynamic["WASM Dynamic Plugins\nHot-loaded at runtime"]
    end

    DB[("Data Store\nPostgreSQL")]

    Workflow -.->|Spec-driven| Frontend
    Workflow -.->|Spec-driven| Host
    UI -->|HTTP| API
    Plugin -->|Load| Source
    Plugin -->|Sandbox execution| Dynamic
    Svc --> DB
    Gov --> DB
```

## Directory Structure

```text
/
├── AGENTS.md                         # AI agent collaboration spec
├── hack/                             # Development utility scripts
├── apps/                             # Application workspace
│    ├── lina-core/                   # Core host service (Go)
│    ├── lina-vben/                   # Management workspace (Vue 3)
│    └── lina-plugins/                # Plugin system
└── openspec/                         # AI development workflow
```

## Core Host Service (lina-core)

`lina-core` is the framework's stable foundation, built in `Go`. It handles all general platform capabilities. The host service has no knowledge of specific business domains — it provides only general infrastructure.

**Primary responsibilities:**

- Expose `RESTful API` covering system management, plugin governance, user authentication, and access control
- Own database migration scripts, seed data, and runtime configuration
- Load and manage the complete lifecycle of source plugins and `WASM` dynamic plugins
- Run host-level scheduled tasks and the persistent scheduled task subsystem
- Provide the i18n runtime, aggregating language resources from the host and all plugins


## Management Workspace (lina-vben)

`lina-vben` is the standard consumer of `lina-core`, built on `Vue 3 + Vben5 + Ant Design Vue`. The workspace contains no business logic of its own — it is purely the UI expression layer for the host `API` and plugin `API`.

**Key characteristics:**

- Fully aligned with the host on interface contracts, the permission model, and design conventions
- Plugin menus are dynamically injected via the host's dynamic routing — no changes to workspace code required
- Developers can replace any module through the plugin mechanism, including replacing the entire workspace

## Dual-Mode Plugin System (lina-plugins)

The plugin layer is the heart of `LinaPro`'s extensibility. It supports two fundamentally different plugin delivery modes. The host exposes stable extension seams through the `pluginhost` package; plugins interact with the host only through these seams and never reach into the host's internal implementation.

See [Dual-Mode Plugin System](/docs/plugin-system).

## AI Development Workflow

The AI development workflow sits above all layers. It is the connective tissue that keeps specs, code, and tests in sync. The workflow does not belong to any runtime component — it acts as a governance mechanism that constrains the entire development process. `OpenSpec` is the recommended tool for implementing this workflow.

See [AI Spec-Driven Development](/docs/spec-driven-development).

## AI Skill System (.agents/skills/)

`LinaPro` ships built-in AI skills covering the full development lifecycle. These skills are embedded in the framework's AI collaboration spec (`.agents/skills/`) as domain knowledge. They don't belong to any runtime component, but they are what AI tools draw on to make framework-compliant decisions in every specific work context.

The skills span backend development, frontend design, test coverage, code review, performance auditing, version upgrades, and more than ten other domains. No installation is required — AI tools activate the relevant skill automatically when handling the corresponding scenario.

See [AI-Native Design](/docs/ai-native).

## Layer Boundary Principles

Understanding the following boundary principles helps avoid common architectural mistakes:

### The host has no knowledge of business domains

`lina-core` contains no specific business logic (e.g., product management, order processing). In the ideal design, such capabilities are delivered through plugins, while the host provides only stable, general infrastructure.

### Plugins do not modify the host's internals

Plugins interact with the host through the stable interfaces exposed by the `pluginhost` package — route registration, event hooks, menu filtering, permission filtering. Plugins never call the host's `internal/` packages directly and never access the host's database tables directly.

### The host owns top-level menu directories

The host owns stable top-level menu directory keys (`dashboard`, `iam`, `setting`, `scheduler`, `extension`, `developer`). Plugin menus can only be mounted under directories the host has published, or under the plugin's own menu tree.

### The workspace is a consumer, not a definer

`lina-vben` consumes the host `API` and plugin `API`, but does not define their behavior. Workspace UI changes should not affect backend behavior; backend API changes are communicated to the workspace through interface contract documents.

## Data Interactions

```mermaid
sequenceDiagram
    participant User as User Browser
    participant Vben as lina-vben
    participant Core as lina-core
    participant Plugin as Source Plugin
    participant DB as PostgreSQL

    User->>Vben: Access management workspace
    Vben->>Core: GET /api/v1/menu (fetch menu)
    Core->>Plugin: Trigger menu.filter hook
    Plugin-->>Core: Inject plugin menus
    Core-->>Vben: Return full menu tree
    Vben-->>User: Render sidebar (with plugin menus)

    User->>Vben: Click plugin page
    Vben->>Core: Request plugin API
    Core->>Plugin: Route to plugin controller
    Plugin->>DB: Access plugin namespace data
    DB-->>Plugin: Return data
    Plugin-->>Core: Return response
    Core-->>Vben: Return API response
    Vben-->>User: Render page data
```

## Deployment Models

`LinaPro` supports two deployment modes, switchable at any time as the business grows:

### Single-node mode (default)

```mermaid
graph TB
    Nginx["Nginx"]
    Core["lina-core :8080"]
    DB[("PostgreSQL")]

    Nginx --> Core
    Core --> DB
```

Suitable for development environments and small-scale production deployments. Simple configuration, low resource usage.

### Cluster mode

```mermaid
graph TB
    LB["Load Balancer"]
    N1["lina-core Node 1"]
    N2["lina-core Node 2\n(with distributed election)"]
    NN["lina-core Node N"]
    DB[("PostgreSQL")]

    LB --> N1
    LB --> N2
    LB --> NN
    N1 --> DB
    N2 --> DB
    NN --> DB
```

Cluster mode is enabled by setting `cluster.enabled: true`. The framework implements distributed leader election, distributed locks, and permission topology caching on top of `PostgreSQL` (`sys_locker`, `sys_kv_cache` tables) — no additional middleware required, and no changes to business code. See [Native Distributed Architecture](/docs/distributed-architecture).

## Related Documents

import DocCardList from '@theme/DocCardList';

<DocCardList />
