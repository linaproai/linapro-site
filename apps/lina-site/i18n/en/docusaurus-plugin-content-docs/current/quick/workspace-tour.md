---
slug: '/quick/workspace-tour'
title: 'Workspace Tour'
sidebar_position: 2
description: 'This quick tour introduces the LinaPro management workspace after the first login, including access control, system settings, job scheduling, extension management, developer tools, and official plugin modules, so new users can quickly understand which capabilities are provided out of the box.'
keywords:
  - LinaPro
  - workspace tour
  - management workspace
  - lina-vben
  - user management
  - role management
  - menu management
  - data dictionary
  - parameter settings
  - file management
  - job scheduling
  - plugin management
  - API docs
  - system info
  - official plugins
  - RBAC
---

After signing in to `http://localhost:5666`, use the workspace to confirm what `LinaPro` already provides before you add business-specific modules.

## Built-in Areas

| Area | What to check |
| --- | --- |
| Access control | Users, roles, menus, and button-level permissions wired to backend `RBAC`. |
| System settings | Dictionaries, runtime parameters, and file management. |
| Job scheduling | Job definitions, groups, execution history, and controlled shell execution. |
| Extension center | Plugin installation state, enablement, disablement, and version visibility. |
| Developer center | Online `API` documentation and runtime system information. |

## Official Plugin Modules

The repository includes official source plugins under `apps/lina-plugins/`. When installed, they inject menus and routes through the plugin lifecycle instead of scattering feature code inside the host.

| Plugin | Workspace capability |
| --- | --- |
| `org-center` | Department and position management. |
| `content-notice` | Notice and announcement management. |
| `monitor-online` | Online session inspection and force logout. |
| `monitor-server` | Server resource and runtime monitoring. |
| `monitor-operlog` | Operation audit logs. |
| `monitor-loginlog` | Login audit logs. |

## First Checks

1. Open **Access Control** and confirm the default administrator role has menus and operations assigned.
2. Open **Extension Center** and review the official plugin list.
3. Open the `API` documentation page and inspect one protected endpoint.
4. Change no production-facing setting until you have read the development manual.
