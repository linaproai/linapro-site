---
slug: '/docs/dynamic-configuration'
title: 'Framework Dynamic Configuration'
hide_title: true
description: 'Detailed explanation of main framework runtime hot-updatable parameters, covering the sys_config data table-driven system built-in parameters and business plugin custom parameter classification management, backend runtime parameters, public frontend display parameters, and TZ environment variable — explaining parameter caching mechanisms, cluster sync strategies, and typical use cases, helping ops engineers dynamically adjust system behavior without restarting processes.'
keywords:
  - dynamic configuration
  - runtime parameters
  - sys_config
  - hot update
  - cache mechanism
  - cluster sync
  - system built-in parameters
  - business custom parameters
  - JWT expiration
  - session timeout
  - upload size limit
  - IP blacklist
  - log retention
  - scheduled tasks
  - frontend display configuration
  - application name
  - Logo configuration
  - login page configuration
  - theme configuration
  - watermark configuration
  - timezone configuration
---

## Introduction

In addition to the `config.yaml` static configuration, the main framework maintains a set of **runtime hot-updatable parameters** through the database `sys_config` table. These parameters take effect immediately after modification without restarting the process, making them suitable for direct adjustment in the admin console.

Runtime parameters use in-process memory caching with a `TTL` of `1` hour. Cache refresh uses a **revision-driven mechanism** rather than relying solely on `TTL` expiration: when parameters are modified, the revision number is immediately advanced, and subsequent reads detect the revision change to auto-refresh the cache, with a freshness budget of `10` seconds.

In standalone mode, changes are detected via in-process revision numbers; in cluster mode, `Redis`-coordinated revision sync maintains consistency across nodes — modifications immediately publish cross-node revision advances, and other nodes sync on the next read.

> For the framework's cache design, see: [Cache Architecture and Consistency Design](/docs/cache-design).

## Parameter Classification and Management

Runtime parameters in the `sys_config` table are divided into two major categories by governance attributes: **system built-in parameters** and **business custom parameters**.

### System Built-in Parameters

System built-in parameters are predefined by the main framework with the following governance characteristics:

- **Protected key names**: Cannot be renamed or deleted; `CRUD` interfaces reject related operations
- **Built-in defaults**: Each parameter has a framework-preset default value
- **Value validation rules**: Format and range validation automatically executed by the framework on write

There are `19` system built-in parameters total, divided into two sub-categories by purpose:

| Sub-Category | Count | Purpose |
|--------------|-------|---------|
| Backend runtime parameters | `7` | Control process runtime behavior, e.g., `JWT` expiration, session timeout |
| Public frontend display parameters | `12` | Control branding, login page, theme layout, watermark, etc. |

### Business Custom Parameters

Business custom parameters are freely created by plugins or administrators for storing business configuration. Their characteristics:

- **Free key names**: Can use any key name (including `sys.` prefix), e.g., `sys.index.skinName`
- **No built-in defaults**: Value must be explicitly specified at creation
- **No automatic validation**: Framework does not validate their value format
- **Fully manageable**: Supports free creation, modification, and deletion

> Both types share the same `sys_config` table; the difference is only in governance policy. Business plugins read and write runtime configuration through the `hostconfigcap` capability without maintaining a separate parameter registry.

## Runtime Parameter Overview

### Backend Runtime Parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `sys.jwt.expire` | Duration string | `"24h"` | Overrides `jwt.expire` in `config.yaml`, controlling token expiration |
| `sys.session.timeout` | Duration string | `"24h"` | Overrides `session.timeout` in `config.yaml`, controlling session timeout |
| `sys.upload.maxSize` | Positive integer (`MB`) | `"100"` | Overrides `upload.maxSize` in `config.yaml`, controlling upload size limit |
| `sys.login.blackIPList` | Semicolon-separated `IP` or `CIDR` | Empty | Login `IP` blacklist, supporting exact addresses and `CIDR` ranges, e.g., `192.168.1.100;10.0.0.0/8` |
| `sys.log.retentionDays` | Positive integer (days) | `"90"` | Max log file retention days; expired logs auto-deleted by cleanup task |
| `sys.cron.shell.enabled` | Boolean string | `"true"` | Global `Shell` type scheduled task toggle; force-disabled on `Windows` regardless of this value |
| `sys.cron.log.retention` | `JSON` object | `{"mode":"days","value":30}` | Scheduled task log default cleanup policy |

`sys.cron.log.retention` supports three cleanup modes:

| Mode | Description |
|------|-------------|
| `days` | Delete logs older than `N` days; `value` is a positive integer |
| `count` | Keep only the most recent `N` logs; `value` is a positive integer |
| `none` | Disable auto-cleanup; `value` is forced to `0` |

### Frontend Display Runtime Parameters

The following parameters are returned to the admin workspace through the public frontend configuration API, used for branding, login page, and interface layout:

| Key | Default | Description |
|-----|---------|-------------|
| `sys.app.name` | `"LinaPro.AI"` | Application name, max `120` characters |
| `sys.app.logo` | `"/logo.webp"` | Light mode `Logo` path |
| `sys.app.logoDark` | `"/logo.webp"` | Dark mode `Logo` path |
| `sys.user.defaultAvatar` | `"/avatar.webp"` | User default avatar path |
| `sys.auth.pageTitle` | Framework description text | Login page title, max `120` characters |
| `sys.auth.pageDesc` | Framework description text | Login page description, max `500` characters |
| `sys.auth.loginSubtitle` | Login guidance text | Login form subtitle, max `120` characters |
| `sys.auth.loginPanelLayout` | `"panel-right"` | Login panel layout: `panel-left`, `panel-center`, `panel-right` |
| `sys.ui.theme.mode` | `"light"` | Theme mode: `light`, `dark`, `auto` |
| `sys.ui.layout` | `"sidebar-nav"` | Global layout: `sidebar-nav`, `sidebar-mixed-nav`, `header-nav`, `header-sidebar-nav`, `header-mixed-nav`, `mixed-nav`, `full-content` |
| `sys.ui.watermark.enabled` | `"false"` | Global watermark toggle |
| `sys.ui.watermark.content` | `"LinaPro"` | Watermark text content, max `120` characters |

## Parameter Management Methods

### Management Interface

System built-in and business custom parameters are uniformly managed through the `sysconfig` service's `CRUD` interface:

| Operation | Description |
|-----------|-------------|
| `List` | List query with pagination and filtering |
| `GetByKey` | Query single parameter by key name |
| `Create` | Create parameter; system built-in parameters auto-marked `isBuiltin=1` |
| `Update` | Update parameter value; system built-in parameters prohibit key name renaming |
| `Delete` | Delete parameter; system built-in parameters prohibit deletion |
| `Export` / `Import` | Batch export/import; imports execute value validation for system built-in parameters |

### Tenant Override

Configuration values support tenant-level overrides. When tenant context exists, the framework prioritizes the tenant-level value, falling back to platform level. `API` responses include fallback metadata:

- `isFallback`: Whether the current value is from platform-level fallback
- `canEdit`: Whether the current tenant can edit
- `canOverride`: Whether tenant-level override is supported
- `overrideMode`: Override mode

### Plugin Access

Business plugins access runtime configuration through the `hostconfigcap` capability:

- **Read-only access**: `hostconfigcap.Service` provides `Get`, `Exists`, `String`, `Bool`, `Int`, `Duration` methods
- **Admin access**: `hostconfigcap.AdminService` provides `BatchGetRuntimeConfig` and `SetRuntimeConfigJSON` methods

Plugin config writes automatically trigger revision advancement, ensuring cluster cache consistency.

## TZ Environment Variable

The main framework reads the `TZ` environment variable when detecting the frontend timezone. The fallback chain is: `TZ` environment variable → process timezone → `Asia/Shanghai`. In containerized deployments, set `TZ` to ensure frontend display matches the business timezone.
