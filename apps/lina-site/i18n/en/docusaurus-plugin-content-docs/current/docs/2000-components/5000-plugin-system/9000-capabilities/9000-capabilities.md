---
slug: '/docs/plugin-capability-services'
title: 'Plugin Capability Services Overview'
hide_title: true
description: "An architectural overview of LinaPro framework capability services (capability.Services) exposed to plugins — the design principles behind the Services interface, service classification, access patterns, and a scenario-based guide for choosing the right service, helping plugin developers understand each capability service's positioning, boundaries, and collaboration patterns."
keywords:
  - plugin capability
  - capability.Services
  - pluginhost.Services
  - capability services
  - plugin development
  - service architecture
  - authentication service
  - cache service
  - configuration service
  - i18n service
  - tenant capability
  - organization capability
  - plugin lifecycle
  - notification service
  - session service
  - LinaPro
---

## Introduction

The `LinaPro` core framework exposes a set of stable capability services to plugins through the `capability.Services` interface. These services cover the most common cross-cutting concerns in plugin development: authentication and context, configuration and resources, data and storage, plugin governance, notifications, and framework-level capabilities such as organization and tenant management.

Source plugins obtain the full service catalog through `pluginhost.Services`, which extends `capability.Services` by adding `TenantFilter()`, providing a separate entry point for capabilities that carry database query builders.

This service architecture follows several core design principles:

- **Explicit contracts, stable boundaries.** Each service has a clear contract definition (in the `contract` package). Plugins depend only on stable contracts, not on host-internal implementations.
- **Plugin-scoped isolation.** Configuration, cache, manifest resources, and similar services are automatically bound to the current plugin ID, preventing interference between plugins.
- **Optional capabilities, safe degradation.** Framework-level capabilities such as organization and tenant management automatically degrade when their providers are unavailable. Plugins check availability through `Available()`.
- **Read-only consumption, minimal exposure.** Ordinary plugins receive read-only consumer interfaces; write operations and database query builders are not exposed through `capability.Services`.

## How to Obtain Services

Source plugins obtain the full service catalog through `registrar.Services()` during route registration, hook callbacks, and scheduled task registration:

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.Services()

    // Access capability services through services
    config := services.Config()
    tenantFilter := services.TenantFilter()
    i18n := services.I18n()
    // ...
    return nil
}
```

The `pluginhost.Services` returned by `registrar.Services()` embeds all 16 services from `capability.Services` and additionally provides `TenantFilter()`. In hook callbacks and scheduled task registration, services are obtained the same way via `payload.Services()` or `registrar.Services()`.

## Service Classification Quick Reference

| Category | Service | Contract Type | One-Line Description |
|----------|---------|---------------|---------------------|
| <span style={{whiteSpace: 'nowrap'}}>Auth & Context</span> | `APIDoc()` | `contract.APIDocService` | API documentation localization — parses route operation keys and translates text |
| <span style={{whiteSpace: 'nowrap'}}>Auth & Context</span> | `Auth()` | `contract.AuthService` | Tenant token issuance, switching, and impersonation token management |
| <span style={{whiteSpace: 'nowrap'}}>Auth & Context</span> | `BizCtx()` | `contract.BizCtxService` | Reads business context snapshots for the current request — user, tenant, impersonation state, etc. |
| <span style={{whiteSpace: 'nowrap'}}>Auth & Context</span> | `I18n()` | `contract.I18nService` | Runtime translation, request locale retrieval, and translation key search |
| <span style={{whiteSpace: 'nowrap'}}>Config & Resources</span> | `Config()` | `contract.ConfigService` | Reads the current plugin's own static configuration |
| <span style={{whiteSpace: 'nowrap'}}>Config & Resources</span> | `HostConfig()` | `contract.HostConfigService` | Reads the host's publicly whitelisted configuration keys |
| <span style={{whiteSpace: 'nowrap'}}>Config & Resources</span> | `Manifest()` | `contract.ManifestService` | Reads original resource files under the current plugin's `manifest/` directory |
| <span style={{whiteSpace: 'nowrap'}}>Data & Storage</span> | `Cache()` | `contract.CacheService` | Plugin-scoped runtime cache |
| <span style={{whiteSpace: 'nowrap'}}>Data & Storage</span> | `Session()` | `contract.SessionService` | Online session management — paginated queries and session kick-out |
| <span style={{whiteSpace: 'nowrap'}}>Data & Storage</span> | `Route()` | `contract.RouteService` | Retrieves metadata for the current dynamic route |
| <span style={{whiteSpace: 'nowrap'}}>Plugin Governance</span> | `PluginLifecycle()` | `contract.PluginLifecycleService` | Plugin lifecycle orchestration — pre-checks and notifications for tenant-level disable and delete |
| <span style={{whiteSpace: 'nowrap'}}>Plugin Governance</span> | `PluginState()` | `contract.PluginStateService` | Queries plugin enablement state |
| <span style={{whiteSpace: 'nowrap'}}>Notifications</span> | `Notify()` | `contract.NotifyService` | Publishes notifications to the host inbox |
| <span style={{whiteSpace: 'nowrap'}}>Capability Provider</span> | `Org()` | `orgcap.Service` | Organization capability consumption — read-only projections of user department, position, etc. |
| <span style={{whiteSpace: 'nowrap'}}>Capability Provider</span> | `Tenant()` | `tenantcap.Service` | Tenant capability consumption — current tenant, visibility validation, tenant switching |
| <span style={{whiteSpace: 'nowrap'}}>Data & Storage</span> | `TenantFilter()` | `contract.TenantFilterService` | Injects `tenant_id` filter conditions into plugin-owned tables |


## Related Content

import DocCardList from '@theme/DocCardList';

<DocCardList />
