---
slug: '/docs/admin-workspace'
title: 'Management Workspace'
hide_title: true
description: 'A comprehensive look at lina-vben, the LinaPro default management workspace — its functional modules, technology stack, and design characteristics, including access control, system settings, job scheduling, extension center, developer center, official plugin extensions, and the dynamic menu injection mechanism.'
keywords:
  - lina-vben
  - management workspace
  - Vue3
  - Vben5
  - Ant Design Vue
  - access control
  - user management
  - role management
  - menu management
  - dictionary management
  - scheduled tasks
  - plugin management
  - API documentation
  - dynamic menus
  - RBAC
  - LinaPro
---

## Overview

`lina-vben` is `LinaPro`'s default management workspace, built on `Vue 3 + Vben5 + Ant Design Vue + TypeScript`. It is the standard UI expression layer for the `lina-core` host `API` and all plugin `API`s.

Developers can build business applications directly on top of the workspace, extend or replace any module on demand through the plugin mechanism, or replace it entirely with a custom frontend workspace.

## Technology Stack

| Technology | Version | Description |
|------------|---------|-------------|
| `Vue 3` | `3.x` | Frontend framework |
| `Vben5` | `5.x` | Enterprise Vue admin framework |
| `Ant Design Vue` | `4.x` | UI component library |
| `TypeScript` | `5.x` | Type system |
| `pnpm` | `8.x` | Package manager (monorepo) |

## Functional Modules

### Access Control

**User Management**

| Feature | Description |
|---------|-------------|
| **User list** | Paginated list with filters for username, phone number, and status |
| **Create user** | Set username, password, roles, phone number, email, and other basic info |
| **Edit user** | Modify user info, reset password, adjust role assignments |
| **Status management** | Enable or disable user accounts; disabled accounts take effect immediately |
| **Bulk export** | Export the user list to an `Excel` file |

**Role Management**

| Feature | Description |
|---------|-------------|
| **Role list** | Display all roles with their description and user count |
| **Create role** | Define role name and description |
| **Permission assignment** | Tree-based selection of menu permissions (three levels: directory / menu / button) |
| **Permission propagation** | Changes to the permission topology take effect quickly (immediately on single-node, within 3 seconds on cluster) — no user re-login required |

**Menu Management**

| Feature | Description |
|---------|-------------|
| **Menu tree** | Display the complete menu hierarchy |
| **Menu types** | Supports directory (`D`), menu (`M`), and button (`B`) |
| **Dynamic configuration** | Menu paths, components, icons, and sort order are all adjustable from the management UI |
| **Permission identifiers** | Button-type menus bind permission identifiers used for `RBAC` authorization |

### System Settings

**Dictionary Management**

| Feature | Description |
|---------|-------------|
| **Dictionary types** | Manage all dictionary categories in the system, such as status codes, types, and flags |
| **Dictionary data** | Maintain the option values and display labels for each dictionary type |
| **Import/Export** | Bulk import dictionary data from `Excel` and export backups |

**Parameter Settings**

| Feature | Description |
|---------|-------------|
| **Parameter list** | Maintain configurable runtime parameters |
| **Parameter editing** | Modify parameter values; changes take effect immediately |
| **Import/Export** | Bulk import and export parameter configurations |

**File Management**

| Feature | Description |
|---------|-------------|
| **File upload** | Supports single and multi-file uploads |
| **File list** | View uploaded files including name, size, and upload time |
| **File download** | Download stored files |
| **Storage isolation** | Host and plugin files are stored in separate namespaces |

### Job Scheduling

**Task Management**

| Feature | Description |
|---------|-------------|
| **Task list** | View all scheduled tasks with their status and next execution time |
| **Create task** | Configure a `Cron` expression, task type (`Go` handler or `Shell` command), and timeout |
| **Trigger now** | Manually trigger a task for immediate execution |
| **Pause/Resume** | Pause a running task or resume a paused one |
| **Execution history** | View execution records including duration and result |

**Group Management**

Group scheduled tasks by business domain for easier navigation and permission control.

**Execution Logs**

| Feature | Description |
|---------|-------------|
| **Log list** | View historical execution logs with filtering by task and time range |
| **Log details** | View the full output and error information for each execution |

### Extension Center

**Plugin Management**

Plugin management is the control panel for `LinaPro`'s extensibility:

| Feature | Description |
|---------|-------------|
| **Plugin list** | Display all discovered plugins including their type (source/dynamic), version, and status |
| **Install** | Transition a plugin from "discovered" to "installed" — runs install SQL and initialization |
| **Enable/Disable** | Toggle a plugin's runtime state; menus and routes are hidden immediately on disable |
| **Uninstall** | Remove a plugin completely, with options to keep or clean up plugin data |
| **Upload dynamic plugin** | Upload a `WASM` dynamic plugin package (`.wasm` file) |
| **Version management** | View the plugin's current version and any available upgrades |

### Developer Center

**API Documentation**

At startup, the host automatically scans the host and all enabled plugins' interfaces to generate a unified online `API` documentation:

- Browse all interface request parameters and response structures in the browser
- Send `API` debug requests directly from the page
- Documentation content always reflects the currently enabled plugins

**System Information**

| Item | Description |
|------|-------------|
| `CPU` usage | Current server `CPU` load |
| Memory usage | Used memory and total memory |
| Disk information | Usage for each disk partition |
| Runtime information | `Go` runtime version, goroutine count, etc. |

## Official Plugin Extensions

The following modules are provided by official plugins. Once a plugin is installed and enabled, its menus are automatically injected into the management workspace — no changes to workspace code required:

| Plugin | Mount location | Menu modules | Key features |
|--------|---------------|--------------|--------------|
| `org-center` | Organization | Department management, position management | Hierarchical org chart, position definitions |
| `content-notice` | Content | Notices & announcements | Announcement CRUD with multiple notice types |
| `monitor-online` | System monitor | Online users | Real-time session view, force logout |
| `monitor-server` | System monitor | Server monitor | CPU, memory, and disk metrics |
| `monitor-operlog` | System monitor | Operation logs | User action audit with request params and duration |
| `monitor-loginlog` | System monitor | Login logs | Login record query with IP and device fingerprint |

## How Dynamic Menu Injection Works

Plugin menus are not hardcoded in the workspace — they are injected dynamically through the following mechanism:

```mermaid
sequenceDiagram
    participant WorkSpace as Management Workspace
    participant Core as lina-core
    participant Plugin as Source Plugin

    WorkSpace->>Core: GET /api/v1/menu (request menu)
    Core->>Plugin: Trigger menu.filter hook
    Note over Plugin: Check if plugin is enabled
    Plugin-->>Core: Inject plugin menu items into menu tree
    Core-->>WorkSpace: Return full menu tree (with plugin menus)
    WorkSpace->>WorkSpace: Render sidebar
    Note over WorkSpace: Plugin menu component path<br/>loaded via dynamic-page wrapper
```

When a plugin is disabled, the menu API response no longer includes that plugin's menu items, and the corresponding entry disappears from the UI automatically — no redeployment or config refresh needed.

## Default Account

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `admin123` |
| Access URL | `http://localhost:5666` |
