---
slug: '/docs/multi-tenant'
title: 'Native Multi-Tenancy'
hide_title: true
description: 'A component-level guide to LinaPro multi-tenancy — how the core framework bizctx, TenantFilterService, tenant_id filtering seam, default platform tenant, official multi-tenant source plugin, tenant impersonation, plugin multi-tenant manifest fields, and the current Pool shared-database model work together.'
keywords:
  - multi-tenancy
  - tenant_id
  - Pool model
  - platform tenant
  - PLATFORM
  - bizctx
  - TenantFilterService
  - TenantFilterColumn
  - CurrentContext
  - PlatformBypass
  - tenant impersonation
  - multi-tenant plugin
  - tenant management
  - tenant members
  - tenant plugin governance
  - scope_nature
  - supports_multi_tenant
  - default_install_mode
  - LifecycleGuard
---

## Introduction

`LinaPro` splits multi-tenancy into two layers: core framework foundation seams and an official tenant control plane. The core framework provides a stable foundation across request context, plugin service contracts, and data filtering. The official `multi-tenant` source plugin handles tenant entities, membership, tenant resolution, tenant impersonation, and tenant-scoped plugin governance.

When the `multi-tenant` plugin is not installed or not enabled, the framework runs in single-tenant mode out of the box. The core framework uses `tenant_id = 0` to represent the platform tenant, so existing projects do not need to change their deployment model in anticipation of future multi-tenant evolution.

## Capability Layers

| Layer | Location | Responsibility |
|-------|----------|---------------|
| **Core framework foundation** | `apps/lina-core` | Request-scoped `bizctx`, identity snapshots, `tenant_id` filtering seam, platform bypass policy, plugin multi-tenant metadata |
| **Tenant control plane** | `apps/lina-plugins/multi-tenant` | Tenant lifecycle, membership, tenant resolution, tenant switching, tenant impersonation, tenant plugin governance |
| **Tenant-aware plugins** | `apps/lina-plugins/<plugin-id>` | Declare multi-tenant capabilities and isolate data by `tenant_id` in their own tables |

This layering ensures the core framework remains stable while tenant business rules can be extended through official or business plugins.

## Default Platform Tenant

In the default context, `tenant_id = 0` represents the `PLATFORM` tenant:

- Single-tenant projects can run directly on the platform tenant.
- Both core framework base tables and plugin tables can retain the `tenant_id` column, leaving room for future multi-tenant evolution.
- Whether platform requests bypass tenant filtering is controlled by core framework policy — plugins should not hardcode this.
- When the `multi-tenant` plugin is enabled, users can enter a specific tenant context, and platform administrators can perform tenant impersonation.

## Request Context

The core framework exposes a read-only request snapshot to source plugins through `bizctx`. Key fields include:

| Field | Description |
|-------|-------------|
| `UserID` | Current authenticated user ID |
| `Username` | Current authenticated username |
| `TenantID` | Tenant ID for the current request |
| `ActingUserID` | Real platform operator during tenant impersonation |
| `ActingAsTenant` | Whether the current request runs in a tenant view |
| `IsImpersonation` | Whether the current token represents an impersonation request |
| `PlatformBypass` | Whether the current request is under a platform bypass policy |

Plugins do not need to access core framework internal context objects — they read the necessary information through the contracts published by `pluginservice`.

## Tenant Filter Service

The core framework publishes `TenantFilterService` to source plugins for appending tenant filter conditions to plugin-owned tables. The default filter column is defined by `TenantFilterColumn`:

```go
const TenantFilterColumn = "tenant_id"
```

The service contract provides two core methods:

| Method | Description |
|--------|-------------|
| `Context(ctx)` | Returns the current request's tenant, operator, impersonation status, and platform bypass metadata as a `TenantFilterContext` |
| `Apply(ctx, model, qualifier)` | Appends a `tenant_id` filter condition to the model; returns the original model when platform bypass is active; `qualifier` is the table name or alias for joined queries, or an empty string for single-table queries |

Single-table list query example:

```go
func ListArticles(ctx context.Context, filter contract.TenantFilterService) ([]*entity.Article, error) {
    model := filter.Apply(ctx, dao.ContentArticle.Ctx(ctx), "")

    var rows []*entity.Article
    err := model.Scan(&rows)
    return rows, err
}
```

For joined queries, use `qualifier` to specify the table name or alias and avoid column ambiguity:

```go
// Qualify the tenant filter column with the table name, generating content_article.tenant_id = ?
model = filter.Apply(ctx, dao.ContentArticle.Ctx(ctx), "content_article")
```

When you need the tenant ID manually for write operations, use the `Context` method:

```go
tenantID := filter.Context(ctx).TenantID
```

Tenant-aware plugins should reuse this service rather than hand-writing `Where("tenant_id", ...)` everywhere. Otherwise, platform bypass and tenant impersonation scenarios may produce inconsistent behavior.

## Shared-Database Pool Model

The current version uses the `Pool` shared-database model: different tenants' data lives in the same database and the same set of tables, distinguished by the `tenant_id` column.

| Model | Current Status | Description |
|-------|---------------|-------------|
| `Pool` shared database | Supported | Default model, suitable for internal multi-team backends and early `SaaS` scenarios |
| `Schema per tenant` | Not built-in | Can be extended on top of the existing governance model |
| `Database per tenant` | Not built-in | Can be extended for stronger isolation requirements |

Plugin-owned tables that need multi-tenant support should include a `tenant_id` column at creation time and build appropriate indexes for tenant filtering.

## Official Multi-Tenant Plugin

The official `multi-tenant` plugin is a source plugin declared as a platform-level governance plugin:

```yaml
id: multi-tenant
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
```

These fields indicate it is not a tenant-scoped business plugin but a tenant control plane:

| Field | Value | Meaning |
|-------|-------|---------|
| `scope_nature` | `platform_only` | Governs tenants only in the platform context |
| `supports_multi_tenant` | `false` | The plugin itself is not installed per-tenant |
| `default_install_mode` | `global` | Single global installation and enablement |

Once enabled, platform administrators can manage tenants, membership, tenant impersonation access, and tenant-aware plugin enablement status.

## Plugin Multi-Tenant Declarations

Business plugins declare their multi-tenant boundaries in `plugin.yaml`:

| Field | Values | Description |
|-------|--------|-------------|
| `scope_nature` | `platform_only` / `tenant_aware` | Whether the plugin belongs to the platform context or can enter tenant contexts |
| `supports_multi_tenant` | `true` / `false` | Whether it supports tenant-scoped installation, enablement, and data isolation |
| `default_install_mode` | `global` / `tenant_scoped` | Whether it is enabled globally by default or independently per tenant |

Example:

```yaml
id: content-article
type: source
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: tenant_scoped
```

The core framework uses these manifest fields in startup consistency checks and plugin governance flows. When the field combination is invalid, the core framework exposes a clear error at startup or during plugin scanning.

## Plugin Enablement Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| `global` | Installed and enabled once, effective for the platform or all tenants | Platform-wide capabilities, global monitoring, unified notifications |
| `tenant_scoped` | Can be independently enabled or disabled per tenant | Content, audit, business modules, optional tenant capabilities |

Whether a new tenant automatically gets a plugin is not decided by `plugin.yaml` directly — it is maintained by core framework governance records and the `multi-tenant` plugin's provisioning policy.

## Tenant Impersonation

Tenant impersonation allows platform administrators to enter a tenant's view for troubleshooting or assisted operations. Impersonation requests must preserve two identity types:

| Identity | Purpose |
|----------|---------|
| Current tenant identity | Determines data access scope and tenant view |
| Real operator identity | Written to audit records to avoid misattributing platform admin actions to tenant users |

When recording audit information, business plugins should prefer `ActingUserID`, `ActingAsTenant`, and `IsImpersonation` from `CurrentContext` rather than only reading the current user ID.

## Lifecycle Guard

Plugins can use lifecycle guards to block dangerous operations. For example, the `multi-tenant` plugin can prevent disablement or uninstallation when tenant data or critical governance relationships still exist.

`plugin.allowForceUninstall` in `config.yaml` controls whether platform administrators can perform an audited force-uninstall after a lifecycle guard vetoes:

```yaml
plugin:
  allowForceUninstall: true
```

Production environments should use force-uninstall capability cautiously and incorporate pre-uninstall data retention, cleanup, and reverse-dependency checks into operational procedures.

## Current Boundaries

The current multi-tenancy capabilities focus on internal `BU`, multi-team backend, early `SaaS`, and tenant-scoped plugin governance scenarios. The following capabilities are not yet provided as default features:

- `Schema per tenant` or `Database per tenant`.
- Tenant quotas, billing, and plan management.
- Tenant-specific branding customization.
- Automatic tenant domain generation via `rootDomain`.

These capabilities can be extended on top of the existing `Pool` model, plugin governance, and tenant context based on business needs.
