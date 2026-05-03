---
slug: '/docs/core-concepts/layers'
title: 'Framework Layers'
sidebar_position: 0
description: 'This core concept page explains the four LinaPro layers: the Go core host service, Vue management workspace, official plugin workspace, and OpenSpec AI-native workflow. It clarifies each layer responsibility, how they collaborate, and why the framework keeps host, UI, plugin, and specification ownership separated.'
keywords:
  - LinaPro
  - framework layers
  - core concepts
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - core host
  - management workspace
  - plugin system
  - AI-native workflow
  - ownership boundary
  - loose coupling
  - full-stack architecture
  - sustainable delivery
  - repository layout
---

`LinaPro` is organized around four layers that evolve together but keep clear ownership boundaries.

| Layer | Repository path | Responsibility |
| --- | --- | --- |
| Core host service | `apps/lina-core` | Backend `API` contracts, service governance, authentication, permissions, plugin lifecycle, database migration, and scheduled jobs. |
| Management workspace | `apps/lina-vben` | `Vue 3` workspace and reference `UI` for built-in capabilities. |
| Plugin workspace | `apps/lina-plugins` | Official source plugins, plugin demos, manifests, frontend pages, and plugin-owned resources. |
| `AI` R&D workflow | `openspec/` | Specification-driven change records that align requirements, implementation, review, and archived capability baselines. |

## Why The Layers Matter

The host should stay stable and generic. Business modules should prefer plugin ownership when they can be installed, upgraded, or removed independently. The workspace consumes backend contracts but should not define host behavior by itself. `OpenSpec` sits above implementation work so changes retain context after the original conversation disappears.

## Repository Map

```text
apps/
  lina-core/      Core host service
  lina-vben/      Management workspace
  lina-plugins/   Official and sample plugins
hack/
  scripts/install/ Bootstrap installers
  tests/          Playwright E2E suite
openspec/
  changes/        Active and archived changes
  specs/          Current baseline specifications
```

Keep new work inside the layer that owns the behavior. If a change needs to cross layers, write down the contract first.
