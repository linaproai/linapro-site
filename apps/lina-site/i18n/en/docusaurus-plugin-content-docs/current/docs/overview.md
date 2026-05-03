---
slug: '/docs'
title: 'Development Manual'
sidebar_position: 0
description: 'The LinaPro development manual is the long-form reference for developers building production applications with the framework. It explains the core host, management workspace, plugin system, AI-native OpenSpec workflow, built-in capabilities, architecture boundaries, local commands, testing strategy, deployment preparation, and contribution-oriented engineering practices.'
keywords:
  - LinaPro
  - development manual
  - developer guide
  - AI-native framework
  - full-stack framework
  - lina-core
  - lina-vben
  - lina-plugins
  - OpenSpec
  - architecture
  - plugin development
  - RBAC
  - testing
  - deployment
  - best practices
  - sustainable delivery
---

The development manual is the reference section for building real products on `LinaPro`. Use it after the quick-start checks are complete and you need to understand structure, boundaries, extension points, and delivery workflow.

## Manual Structure

| Area | What it covers |
| --- | --- |
| Core concepts | The four framework layers, built-in capabilities, and ownership boundaries. |
| Architecture | Runtime relationships between the management workspace, core host, plugins, database, and workflow. |
| Plugin development | Source plugins, dynamic `WASM` plugins, manifests, lifecycle, and upgrade behavior. |
| Testing and deployment | Local commands, `E2E` expectations, build preparation, and operational checks. |

## Recommended Order

1. Read [Framework Layers](/docs/core-concepts/layers) to understand the main repository areas.
2. Read [Built-in Capabilities](/docs/core-concepts/built-in-capabilities) before deciding what to implement yourself.
3. Read [Runtime Architecture](/docs/architecture/runtime-architecture) before changing shared contracts.
4. Read [Plugin Development](/docs/plugin-development) when a feature should live outside the host.
5. Read [Testing and Deployment](/docs/testing-deployment) before preparing a change for review.

## Source Of Truth

Official site documentation should stay aligned with the `linapro` source repository. When source behavior changes, update the Chinese documentation first, then synchronize the English version with natural wording.
