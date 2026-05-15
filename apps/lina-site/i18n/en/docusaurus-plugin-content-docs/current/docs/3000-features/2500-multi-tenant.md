---
slug: '/docs/multi-tenant'
title: 'Multi-Tenancy'
hide_title: true
description: 'An overview of LinaPro multi-tenancy capabilities, including the host built-in bizctx request context, tenant_id filtering seam, default tenant_id = 0 platform tenant mode, the official multi-tenant source plugin for tenant lifecycle and membership management, plugin multi-tenant manifest fields, global vs. tenant-scoped plugin enablement, LifecycleGuard protection, and the boundaries of the current Pool shared-database model.'
keywords:
  - multi-tenancy
  - tenant_id
  - Pool model
  - platform tenant
  - PLATFORM
  - multi-tenant plugin
  - tenant management
  - tenant members
  - tenant resolution
  - bizctx
  - TenantFilter
  - plugin multi-tenancy
  - scope_nature
  - supports_multi_tenant
  - default_install_mode
  - LifecycleGuard
  - plugin governance
  - LinaPro
---

## Overview

`LinaPro` separates multi-tenancy into two layers: the host provides a stable tenant context and data filtering seam, while the official `multi-tenant` source plugin delivers the visible capabilities — tenant management UI, tenant lifecycle, membership, and tenant-scoped plugin governance.

When the `multi-tenant` plugin is not installed or not enabled, the framework still delivers a single-tenant out-of-the-box experience. Both host and plugin data use `tenant_id = 0` to represent the `PLATFORM` tenant; existing single-tenant projects do not need to change their runtime model to prepare for future multi-tenancy evolution.

## Capability Layers

| Layer | Location | Responsibility |
|-------|----------|----------------|
| Host base capabilities | `apps/lina-core` | Per-request `bizctx`, tenant identity snapshot, `tenant_id` filtering seam, plugin multi-tenancy metadata, platform bypass policy |
| Official management plugin | `apps/lina-plugins/multi-tenant` | Tenant principal, tenant members, tenant resolution policy, tenant impersonation, tenant-scoped plugin enablement governance |
| Tenant-aware plugins | `apps/lina-plugins/<plugin-id>` | Declare multi-tenant capabilities in the manifest and isolate tenant data using `tenant_id` in plugin tables |

## Default Platform Tenant

By default, all requests run in the platform context:

- `tenant_id = 0` represents the `PLATFORM` tenant
- Platform requests can bypass tenant filtering per host policy
- Plugin tables that may need multi-tenancy in the future should include a `tenant_id` column
- Single-tenant projects can run normally without enabling the `multi-tenant` plugin

## Pool Shared-Database Model

The current version uses the `Pool` shared-database model: different tenants' data resides in the same database and table structure, distinguished by the `tenant_id` column.

| Model | Current Status | Notes |
|-------|---------------|-------|
| `Pool` shared tables | Supported | Default model, suitable for most internal systems and early-stage `SaaS` |
| Separate schema | Not available | Reserved as a future evolution path |
| Separate database | Not available | Reserved as a future evolution path |

The host publishes a `TenantFilterService` to source plugins. Plugins can obtain the tenant identity from the current request context and append tenant filtering conditions to their own tables. The default tenant filter column name is `tenant_id`.

## Official multi-tenant Plugin

The official `multi-tenant` plugin is a source plugin that provides visible tenant management capabilities. It is a platform-level governance plugin, declared in its manifest as:

```yaml
id: multi-tenant
type: source                   # Source plugin, compiled with the host
scope_nature: platform_only    # Governed only in platform context, does not enter tenant context
supports_multi_tenant: false   # Does not support tenant-scoped installation — it IS the tenant control plane
default_install_mode: global   # Globally unique, not per-tenant
```

Once enabled, platform administrators can use:

- Tenant lifecycle management — create, edit, delete, and status governance
- Tenant membership management — a user can belong to multiple tenants
- Tenant selection and switching — users can enter a specific tenant context after login
- Platform impersonation of tenants — audit records retain the real operator and the impersonated tenant
- Tenant-scoped plugin governance — enable tenant-aware plugins globally or per-tenant

## Plugin Multi-Tenant Manifest Fields

Plugins declare their multi-tenant boundaries in `plugin.yaml`:

| Field | Values | Description |
|-------|--------|-------------|
| `scope_nature` | `platform_only` / `tenant_aware` | Whether the plugin is a platform-level governance capability or can be governed per tenant context |
| `supports_multi_tenant` | `true` / `false` | Whether it supports tenant-scoped installation, provisioning, and data isolation |
| `default_install_mode` | `global` / `tenant_scoped` | Whether it is enabled globally by default or enabled independently per tenant |

Example:

```yaml
id: demo-control
type: source
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: global
```

## Plugin Enablement Modes

| Mode | Use Case | Behavior |
|------|----------|----------|
| `global` | Platform shared capabilities | Plugin is installed and enabled once, effective for the platform or all tenants |
| `tenant_scoped` | Content, audit, business modules that tenants can independently provision | Plugin can be independently enabled or disabled per tenant |

Whether new tenants automatically have a tenant-aware plugin enabled is governed by the platform plugin registry's provisioning policy, not directly declared in `plugin.yaml`.

## LifecycleGuard Protection

Plugins can use `LifecycleGuard` to run protection checks before being disabled or uninstalled. For example, the `multi-tenant` plugin can block uninstallation when tenant data still exists, preventing accidental irreversible data loss.

The `plugin.allowForceUninstall` setting in `config.yaml` controls whether platform administrators can perform an audited force-uninstall after `LifecycleGuard` vetoes:

```yaml
plugin:
  allowForceUninstall: true
```

Set this option cautiously in production, per your organization's governance policy.

## Current Boundaries

The current multi-tenancy capabilities focus on internal BU, multi-team backends, early-stage `SaaS`, and tenant-scoped plugin governance scenarios. The following capabilities are not yet available as default features:

- `Schema per tenant` or `Database per tenant`
- Tenant quotas, billing, and plan management
- Per-tenant branding customization
- Automatic tenant domain generation via `rootDomain`

To build these capabilities, extend the current `Pool` model and plugin governance layer per your business requirements.
