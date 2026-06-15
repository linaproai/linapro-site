---
slug: '/docs/domain-capability-tenant'
title: 'Tenant'
hide_title: true
description: '`Tenant()` is an optional framework tenant capability that provides plugins with the current tenant, platform bypass, tenant visibility, user-tenant membership, and tenant switch validation. Source plugins can also append `tenant_id` filters to plugin-owned tables through `pluginhost.Services.TenantFilter()`. The actual multi-tenant strategy is implemented by a provider plugin; the host is responsible for request context degradation, provider status checks, source plugin query filtering, and dynamic `hostServices.tenant` bridging.'
keywords:
  - tenant capability
  - tenantcap
  - TenantService
  - TenantFilter
  - PluginTableFilterService
  - tenant_id
  - linapro-tenant-core
  - Provider
  - Resolver
  - multi-tenancy
  - current tenant
  - platform bypass
  - tenant visibility
  - tenant filtering
  - tenant switching
  - hostServices.tenant
  - capability.status
  - plugin capability
  - LinaPro
---

## Overview

Source plugins consume the standard tenant capability through `services.Tenant()`. Tenant is an optional framework capability whose official provider plugin is identified as `linapro-tenant-core`. When no active provider is available, the service degrades to single-tenant semantics for platform tenant `0`.

Dynamic plugins can declare `service: tenant` to invoke published tenant capability methods.

**Capability Phase**: Runtime

**Supported Types**: Source plugins, dynamic plugins

## Capability Design

### SPI Pattern

The `Tenant` capability uses an SPI pattern where the actual multi-tenant strategy is implemented by a provider plugin. `tenantcap.Provider` handles tenant resolution, user-tenant relationship validation, user-visible tenant lists, and tenant switch validation. `tenantcap.Resolver` handles tenant identity resolution from HTTP requests and can compose a responsibility chain based on request headers, domains, paths, tokens, or other strategies.

```mermaid
graph TB
    Request["HTTP Request"] --> Resolver["Tenant resolution"]
    Resolver --> Context["Request tenant context"]
    Plugin["Plugin"] --> Tenant["TenantService"]
    Tenant --> Context
    Tenant --> Provider["linapro-tenant-core"]
```

### Graceful Degradation

When no tenant provider is available, the system degrades to single-tenant mode for the platform tenant. The standard `Tenant()` does not expose `RequestResolver`, `ScopeService`, user-tenant membership writes, or startup consistency checks; these belong to the host's internal middleware, database filtering, or governance flows.

### Source Plugin Exclusive: TenantFilter

Source plugins that need to query plugin-owned tables can obtain `tenantcap.PluginTableFilterService` through `pluginhost.Services.TenantFilter()`. It belongs to the tenant domain capability but is not part of the standard `capability.Services.Tenant()` because it directly receives and returns `*gdb.Model` query builders, making it suitable only for source plugins running inside the host process.

| Method | Description |
|--------|-------------|
| `Context` | Returns the current request's tenant, user, real operator, impersonation state, and platform bypass information |
| `Apply` | Appends a `tenant_id` condition to the query model; returns the original model when the current request allows platform bypass |

`TenantFilterContext` contains `UserID`, `TenantID`, `ActingUserID`, `OnBehalfOfTenantID`, `ActingAsTenant`, `IsImpersonation`, and `PlatformBypass`. Among these, `ActingUserID` is suitable for writing audit records, while `PlatformBypass` is determined by host policy and plugins should not construct it themselves.

```mermaid
graph LR
    Model["Plugin query model"] --> Apply["TenantFilter.Apply"]
    Apply --> Bypass{"PlatformBypass?"}
    Bypass -->|"true"| Same["Return original model"]
    Bypass -->|"false"| Filtered["Append tenant_id condition"]
```

Dynamic plugins do not use `TenantFilter()`. When dynamic plugins access plugin-owned tables, they should declare `service: data` and authorized `resources.tables`, with the host `data` service handling tenant, authorization, and table namespace governance.

## Interface Definitions

### Source Plugin Interface

| Method | Description |
|--------|-------------|
| `Available` | Checks whether the tenant capability has an available provider |
| `Status` | Returns capability status, active provider, and conflict reasons |
| `Current` | Returns the current request tenant; returns the platform tenant when missing |
| `PlatformBypass` | Checks whether the current request allows bypassing tenant filtering |
| `EnsureTenantVisible` | Validates that the current user can access the specified tenant |
| `ValidateUserInTenant` | Validates that the specified user belongs to the specified tenant |
| `ListUserTenants` | Lists active tenants visible to the user |
| `SwitchTenant` | Validates that a tenant switch target is legal |

Source plugin exclusive interface (accessed through `pluginhost.Services.TenantFilter()`):

| Method | Description |
|--------|-------------|
| `Context` | Returns the current request's tenant context information |
| `Apply` | Appends a `tenant_id` condition to the query model |

### Dynamic Plugin Interface

| Dynamic Method | Description |
|----------------|-------------|
| `capability.available` | Checks whether the tenant capability has an available provider |
| `capability.status` | Returns capability status and active provider |
| `tenants.current` | Returns the current request tenant |
| `tenants.platform_bypass` | Checks whether bypassing tenant filtering is allowed |
| `tenants.visible.ensure` | Validates that the current user can access the specified tenant |
| `users.tenant_membership.validate` | Validates that the specified user belongs to the specified tenant |
| `users.tenants.list` | Lists active tenants visible to the user |
| `tenants.switch.validate` | Validates that a tenant switch target is legal |

## Usage

### Source Plugin Usage

Source plugins access the standard tenant capability through `services.Tenant()`:

```go
// Check whether tenant capability is available
if !services.Tenant().Available(ctx) {
    // Handle degradation
    return
}

// Get the current tenant
tenant := services.Tenant().Current(ctx)

// Validate tenant visibility
err := services.Tenant().EnsureTenantVisible(ctx, targetTenantID)

// List user-visible tenants
tenants, err := services.Tenant().ListUserTenants(ctx, userID)
```

Source plugins use `TenantFilter()` to append tenant filtering to plugin-owned tables:

```go
model := g.DB().Model("plugin_record")
model = services.TenantFilter().Apply(ctx, model, "")
result, err := model.Where("status", "active").All()
```

The third parameter of `Apply` is a table name or alias qualifier:

| `qualifier` | Result |
|-------------|--------|
| Empty string | Uses `tenant_id` |
| `plugin_record` | Uses `plugin_record.tenant_id` |
| `r` | Uses `r.tenant_id` |

### Dynamic Plugin Usage

Dynamic plugins declare the `tenant` service and authorized methods in `plugin.yaml`:

```yaml
hostServices:
  - service: tenant
    methods:
      - tenants.current
      - tenants.visible.ensure
      - users.tenants.list
```

`tenant` is a `none` resource type and does not declare `paths`, `tables`, `keys`, or `resources`. Usage on the dynamic plugin side:

```go
tenantSvc := pluginbridge.Default().Tenant()

// Get the current tenant
tenant := tenantSvc.Current(ctx)

// Validate tenant visibility
err := tenantSvc.EnsureTenantVisible(ctx, targetTenantID)

// List user-visible tenants
tenants, err := tenantSvc.ListUserTenants(ctx, userID)
```

When dynamic plugins access plugin-owned tables, they should declare `service: data` and authorized `resources.tables`, with the host data service handling tenant boundary governance.

## Design Constraints

- **Capability is optional.** When no tenant provider is available, the system degrades to single-tenant mode for the platform tenant.
- **Query filtering is not in the standard service.** Tenant scope capabilities requiring database query builders belong to the host-internal `ScopeService` or the source-plugin-exclusive `TenantFilter()`.
- **`TenantFilter()` is only for plugin-owned tables.** Do not use it to operate on host core tables, and do not hand-write inconsistent tenant conditions in plugin code.
- **Pass qualifiers for joined queries.** When multiple tables contain `tenant_id`, pass the table name or alias to avoid column name ambiguity.
- **Tenant switching only validates.** `SwitchTenant` validates target legality; token re-issuance is handled by `Auth().Token().SwitchTenant`.
- **Platform bypass is determined by the host.** Plugins should not construct cross-tenant access states themselves.

## Related Services

- [Auth Capability](/docs/domain-capability-auth)
- [Org Capability](/docs/domain-capability-org)
- [RecordStore Capability](/docs/domain-capability-recordstore)
