---
slug: '/quick/first-plugin'
title: '🛠️ First Plugin Path'
sidebar_position: 3
description: 'This introductory plugin path explains how new LinaPro developers should approach extension work, when to choose a source plugin or a dynamic WASM plugin, which repository directories matter first, and what lifecycle, manifest, frontend, backend, SQL, and upgrade concepts should be understood before writing production plugins.'
keywords:
  - LinaPro
  - first plugin
  - plugin development
  - source plugin
  - dynamic plugin
  - WASM plugin
  - plugin.yaml
  - plugin lifecycle
  - lina-plugins
  - backend API
  - frontend pages
  - SQL manifest
  - menu injection
  - source upgrade
  - extension points
  - sandbox
---

Plugins are the main extension path in `LinaPro`. A plugin can bring its own backend routes, service logic, database resources, frontend pages, menus, hooks, and lifecycle behavior.

## Choose The Plugin Mode

| Mode | Use it when |
| --- | --- |
| Source plugin | The plugin ships with the product source and is wired into the host during development. Official plugins use this mode. |
| Dynamic `WASM` plugin | The plugin is installed, enabled, disabled, or removed at runtime through plugin management. |

For most business modules that live in the same repository as the product, start with a source plugin and use `apps/lina-plugins/plugin-demo-source/` as the reference.

## Files To Read First

| Path | Purpose |
| --- | --- |
| `apps/lina-plugins/README.md` | Plugin workspace rules and official plugin list. |
| `apps/lina-plugins/plugin-demo-source/README.md` | Source plugin structure and registration model. |
| `apps/lina-plugins/plugin-demo-dynamic/README.md` | Dynamic plugin structure and lifecycle model. |
| `apps/lina-plugins/OPERATIONS.md` | Operational plugin workflows. |

## Minimum Mental Model

A production plugin should keep its ownership boundary clear:

1. `plugin.yaml` declares metadata, menu entries, pages, resources, and lifecycle requirements.
2. `backend/` contains routes, controllers, services, and local generated data access artifacts.
3. `frontend/pages/` contains pages mounted by host menus.
4. `manifest/sql/` owns install, mock, and uninstall database resources.
5. Source plugin version upgrades are explicit through `make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>`.

After you understand this shape, continue with the plugin development chapter in the development manual.
