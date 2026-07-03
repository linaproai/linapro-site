---
slug: '/docs/domain-capability-route'
title: 'Route'
hide_title: true
description: '`RouteService` provides a dynamic route metadata view attached to the current request. Source plugins access it through `services.Route()`, while dynamic plugins declare `service: route` in `plugin.yaml` and access it via `pluginbridge.Default().Route()`. It only reads `DynamicRouteMetadata` written by the `pluginbridge` runtime and is not responsible for registering, modifying, or dispatching routes.'
keywords:
  - RouteService
  - routecap
  - DynamicRouteMetadata
  - dynamic routing
  - pluginbridge
  - Wasm plugins
  - route metadata
  - PublicPath
  - ResponseBody
  - ResponseContentType
  - APIDocService
  - audit logging
  - plugin routes
  - capability.Services
  - LinaPro
---

## Overview

`services.Route()` reads dynamic route metadata attached to the current request. Source plugins access it through `services.Route()`, while dynamic plugins declare `service: route` in `plugin.yaml` and access it via `pluginbridge.Default().Route()`. This service primarily supports host audit, operation logging, or source plugin context enrichment when handling dynamic plugin requests.

Source plugins do not need `RouteService` to register their own routes; route registration is done through `pluginhost.Declarations.HTTP().RegisterRoutes`.

**Capability Phase**: Runtime

**Supported Types**: Source plugins, dynamic plugins

## Capability Design

### Metadata View Model

`DynamicRouteMetadata` is a read-only view attached to the current request, containing full descriptive information about a dynamic route:

| Field | Description |
|-------|-------------|
| `PluginID` | The plugin `ID` that owns the current dynamic route |
| `Method` | The `HTTP` method declared by the dynamic route |
| `PublicPath` | The path exposed and matched by the host |
| `Tags` | Tags declared by the dynamic route |
| `Summary` | Summary declared by the dynamic route |
| `Meta` | Plugin-defined custom route metadata |
| `ResponseBody` | The raw response body captured by the dispatcher |
| `ResponseContentType` | The response content type |

### Route Dispatch Flow

```mermaid
graph LR
    Request["/x/{pluginId}/..."] --> Bridge["pluginbridge dispatch"]
    Bridge --> Metadata["Write DynamicRouteMetadata"]
    Metadata --> Route["RouteService read"]
    Route --> Audit["Audit or logging"]
```

### Read-only Data Semantics

Callers cannot modify dynamic routes or responses through this service. Non-dynamic routes return `nil`; callers should perform a nil check before use. `ResponseBody` depends on runtime capture results and should not be treated as an authoritative source of business data.

## Interface Definitions

### Source Plugin Interface

| Method | Description |
|--------|-------------|
| `DynamicRouteMetadata` | Reads dynamic route metadata from `context.Context`; returns `nil` for non-dynamic routes |

### Dynamic Plugin Interface

Dynamic plugins declare authorized methods through `hostServices.route`:

| Dynamic Method | Description |
|----------------|-------------|
| `metadata.get` | Reads dynamic route metadata attached to the current request |

## Usage

### Source Plugin Usage

Source plugins read dynamic route metadata through `services.Route()`, with typical scenarios including audit logging and operation recording:

```go
// Read dynamic route metadata
meta := services.Route().DynamicRouteMetadata(ctx)
if meta != nil {
    // Record audit log
    log.Infof("Dynamic plugin %s route %s %s was accessed", meta.PluginID, meta.Method, meta.PublicPath)
}
```

Source plugins register their own routes using `pluginhost.Declarations.HTTP().RegisterRoutes`:

```go
plugin := pluginhost.NewDeclarations("my-author-my-domain-my-cap")
err := plugin.HTTP().RegisterRoutes(
    pluginhost.ExtensionPointHTTPRouteRegister,
    pluginhost.CallbackExecutionModeBlocking,
    registerRoutes,
)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `route` service in `plugin.yaml`:

```yaml
hostServices:
  - service: route
    methods:
      - metadata.get
```

Dynamic plugins call through `pluginbridge.Default().Route()`:

```go
routeSvc := pluginbridge.Default().Route()
meta := routeSvc.DynamicRouteMetadata(ctx)
```

## Design Constraints

- **Read-only data.** Callers cannot modify dynamic routes or responses through this service.
- **Non-dynamic routes return `nil`.** Callers should perform a nil check before use.
- **Response body may be empty.** `ResponseBody` depends on runtime capture results and should not be treated as an authoritative source of business data.
- **Dynamic plugins read the request envelope directly.** Route information within dynamic plugins comes from `BridgeRequestEnvelopeV1.Route`, not from calling `RouteService` through `hostServices`.

## Related Services

- [APIDoc Capability](/docs/domain-capability-apidoc)
- [Domain Capability Overview](/docs/domain-capabilities)
- [Business Context Capability](/docs/domain-capability-bizctx)
