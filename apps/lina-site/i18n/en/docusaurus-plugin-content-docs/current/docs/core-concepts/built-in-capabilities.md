---
slug: '/docs/core-concepts/built-in-capabilities'
title: 'Built-in Capabilities'
sidebar_position: 1
description: 'This page summarizes the LinaPro built-in capabilities available before custom business development starts, including access control, system settings, file management, job scheduling, plugin governance, developer tools, official organization, notice, online user, server monitor, operation log, and login log plugins.'
keywords:
  - LinaPro
  - built-in capabilities
  - access control
  - user management
  - role management
  - menu management
  - data dictionary
  - parameter settings
  - file management
  - job scheduling
  - plugin governance
  - developer tools
  - organization plugin
  - notice plugin
  - monitor plugins
  - RBAC
---

`LinaPro` includes a set of production-oriented capabilities so teams can start from business value instead of infrastructure wiring.

## Host Capabilities

| Area | Modules |
| --- | --- |
| Access control | User management, role management, menu management, and button-level permission authorization. |
| System settings | Data dictionary, parameter settings, and file management. |
| Job scheduling | Job management, job groups, execution logs, handler registry, `Cron` preview, and controlled shell execution. |
| Extension governance | Plugin management, lifecycle state, and version visibility. |
| Developer center | Online `API` documentation and runtime system information. |

## Official Plugins

| Plugin | Capability |
| --- | --- |
| `org-center` | Department and position management. |
| `content-notice` | Notice and announcement management. |
| `monitor-online` | Online users and force logout. |
| `monitor-server` | Server monitoring and runtime resource inspection. |
| `monitor-operlog` | Operation log persistence and audit queries. |
| `monitor-loginlog` | Login log persistence and audit queries. |
| `demo-control` | Read-only protection for demo environments. |

## Practical Rule

Before creating a new module, check whether the requirement fits an existing host capability, official plugin, or extension point. Reusing the built-in model keeps permissions, menus, logs, and lifecycle behavior consistent.
