---
slug: '/docs/architecture/runtime-architecture'
title: '🏗️ Runtime Architecture'
sidebar_position: 0
description: 'This architecture page describes how LinaPro runs at runtime, showing how the Vue management workspace consumes the Go core host APIs, how the host coordinates governance and plugin lifecycle, how official plugins are loaded, and how OpenSpec keeps requirements, implementation, tests, and archived specifications aligned.'
keywords:
  - LinaPro
  - runtime architecture
  - architecture design
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - management workspace
  - core host
  - plugin lifecycle
  - API contracts
  - governance
  - database
  - Mermaid diagram
  - AI-native workflow
  - full-stack architecture
---

The runtime architecture keeps the host generic, the workspace contract-driven, and plugins independently owned.

```mermaid
graph TB
    subgraph Workflow["AI R&D Workflow"]
        Explore["Explore"] --> Propose["Propose"] --> Implement["Implement"] --> Review["Review"] --> Archive["Archive"]
    end

    subgraph Frontend["Management Workspace  lina-vben"]
        UI["Vue 3 + Ant Design Vue"]
    end

    subgraph Host["Core Host Service  lina-core"]
        API["API contracts"]
        Service["Domain services"]
        Governance["Auth · RBAC · audit · jobs"]
        Runtime["Plugin runtime"]
        API --> Service
        Service --> Governance
        Service --> Runtime
    end

    subgraph Plugins["Plugin Workspace  lina-plugins"]
        Source["Source plugins"]
        Dynamic["Dynamic WASM plugins"]
    end

    DB[("Database")]

    Workflow -.-> Frontend
    Workflow -.-> Host
    UI --> API
    Runtime --> Source
    Runtime --> Dynamic
    Service --> DB
```

## Runtime Contracts

| Contract | Owner |
| --- | --- |
| Backend routes and DTOs | `apps/lina-core/api` and plugin-local `backend/api`. |
| Management pages | `apps/lina-vben` and plugin-local `frontend/pages`. |
| Permissions and menus | Host governance data plus plugin manifests. |
| Database resources | Host SQL manifests or plugin-local `manifest/sql`. |
| Change intent | `openspec/changes` before implementation and `openspec/specs` after archive. |

## Design Rule

Do not make the host depend on plugin internals. Plugins should integrate through manifests, public host packages, lifecycle hooks, and published extension points.
