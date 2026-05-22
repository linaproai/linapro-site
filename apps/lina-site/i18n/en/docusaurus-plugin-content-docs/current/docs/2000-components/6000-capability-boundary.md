---
slug: '/docs/capability-boundary'
title: 'Framework and Plugin Capability Boundaries'
hide_title: true
description: 'This page explains how LinaPro defines capability boundaries between the main framework and plugins. It covers the default admin workspace, host control-plane APIs, the unified plugin API namespace, public plugin asset hosting, source-plugin HTTP routes, and the division of responsibility between source plugins and dynamic plugins. It also explains how APIPrefix, route contracts, public_assets, source, mount, and index are implemented across source plugins and dynamic plugins, so developers can understand exactly how the host and plugins collaborate without leaking responsibilities into each other.'
keywords:
  - host capability boundary
  - plugin capability boundary
  - pluginhost
  - pluginbridge
  - APIPrefix
  - dynamic routes
  - route conflicts
  - plugin API prefix
  - /x prefix
  - /x-assets
  - public_assets
  - public_assets index
  - source
  - mount
  - /admin
  - static asset governance
  - Go Embed
  - host services
  - HostServices
  - portal routes
  - admin routes
  - data plane
  - control plane
  - source plugin routes
  - dynamic plugin sandbox
  - admin workspace
  - route groups
  - route namespace
  - capability boundaries
  - plugin extension
  - LinaPro
---

## Design Philosophy

`LinaPro` was designed around one core principle from the start: **the main framework is responsible only for lightweight foundational capabilities and stable extension interfaces, while business capabilities are implemented through plugins**.

The reason is straightforward. If the main framework carries too much business logic, every business change forces changes to the framework itself, raising upgrade cost and weakening stability. When business capabilities move into plugins instead, the main framework can focus on stable infrastructure: authentication, authorization, multi-tenancy, route dispatch, scheduled tasks, static assets, and plugin lifecycle management. Once these foundations are stable, they can support rapid iteration in upper-layer plugins for the long term.

That does not mean the host and plugins are completely isolated. Collaboration itself needs boundaries and rules. If a plugin depends too deeply on the host, for example by importing the host's `internal` packages directly or bypassing published interfaces to call private implementations, the plugin will break whenever host internals evolve. Conversely, if the host exposes one-off interfaces just to accommodate a specific plugin, unnecessary coupling appears between the host and that plugin, harming maintainability across the whole system.

`LinaPro` establishes clear collaboration boundaries between the host and plugins in three ways:

- **Source plugins integrate through the `pluginhost` contract**: every interface in the `pluginhost` package is a stable commitment from the main framework to source plugins. Source plugins can use host capabilities only through these interfaces and must not import the host's `internal/` directory directly.
- **Dynamic plugins communicate through the `pluginbridge` sandbox**: dynamic plugins run inside a `WASM` sandbox and call host services through the `host_call` mechanism. The host validates every call against the `hostServices` authorization snapshot confirmed at installation time.
- **The main framework provides general foundations, not business-specific interfaces**: `pluginhost` exposes general service adapters for authentication, tenancy, configuration, cache, notifications, and similar platform capabilities. It does not expose domain-specific business APIs for a particular plugin; plugins build their own business logic on top of those primitives.

This design keeps host stability and plugin flexibility compatible. The two sides collaborate across explicit boundaries instead of leaking into each other's implementation.

## Host Capability Design

The foundational capabilities provided by the main framework `lina-core` can be divided into two groups: **infrastructure capabilities** and **plugin extension interfaces**.

### Infrastructure Capabilities

Infrastructure capabilities apply to the whole runtime environment. They are not designed for any single business scenario or page:

| Capability Area | Description |
|-----------------|-------------|
| **Authentication and sessions** | `JWT` issuance and validation, session storage, forced sign-out, session activity refresh |
| **Permission management** | `RBAC` model, menu and button permissions, permission identifier enforcement middleware |
| **Multi-tenancy** | Tenant resolution, tenant context injection, tenant filter service |
| **Routing and middleware** | Unified response serialization, `CORS`, request body limits, business context injection |
| **Scheduled tasks** | Task scheduling, distributed leader election, task execution logs |
| **Static assets** | Embedded frontend build artifacts and unified plugin asset serving |
| **Configuration** | Static configuration reading and runtime configuration items |
| **Cache** | Plugin-scoped cache namespaces |
| **Plugin governance** | Plugin directory scanning, dependency checks, lifecycle orchestration, runtime upgrades |

These capabilities are wrapped as stable service adapters and exposed to plugins through the `HostServices` interface. The main framework does not define a dedicated interface for a single business scenario. Instead, it provides composable general primitives, and each plugin decides how to use them in its own logic.

### Plugin Extension Interfaces

The main framework defines two stable extension interfaces for plugins:

```mermaid
flowchart LR
    subgraph Host["Main framework lina-core"]
        PH["pluginhost<br/>Source plugin integration"]
        PB["pluginbridge<br/>Dynamic plugin sandbox"]
    end

    subgraph Source["Source plugin"]
        SP["plugin.go<br/>init() registration"]
    end

    subgraph Dynamic["Dynamic plugin (WASM)"]
        DP["WASM module<br/>host_call"]
    end

    SP -- "Routes / Hooks / Cron / Lifecycle" --> PH
    DP -- "host_call / BridgeEnvelope" --> PB
```

Source plugins use `pluginhost` at compile time to register routes, hooks, scheduled tasks, lifecycle callbacks, and governance logic. Dynamic plugins use `pluginbridge` at runtime to access host capabilities through sandbox communication. The two interfaces differ in capability scope, but both plugin types are governed transparently by the main framework.

## Admin Workspace Responsibility Boundaries

`LinaPro` includes a default admin workspace built on `Vue 3 + Vben5`. By default, it is served under the `/admin` route. The workspace provides visual management for system modules and is the standard `UI` expression layer for host APIs and plugin APIs.

`/admin` is the entry boundary of the default admin workspace, not the boundary of all frontend capability in the system. The workspace `SPA` static asset service and refresh `fallback` serve only `/admin` and its subpaths. They do not occupy the root path `/` by default. Therefore, the root path, portal pages, plugin-managed static assets, and other non-reserved public paths can be registered by source plugins in code. If no host or plugin route matches a request, the request should return an unmatched result instead of falling back to the admin workspace `index.html`.

### Workspace Entry Route Configuration

The admin workspace entry is controlled by the host static configuration key `workspace.basePath`, which defaults to `/admin`. This configuration is evaluated during startup and determines which browser path serves the built-in workspace `SPA` static assets and refresh `fallback`:

```yaml
workspace:
  basePath: "/admin"
```

For a normal single-domain deployment, keeping `/admin` is recommended so the root path and other public paths remain available for portal pages, source-plugin-owned pages, or static assets. When the admin workspace uses a dedicated domain, `workspace.basePath` may also be set to `/`, allowing the whole domain to serve only the admin workspace.

Whichever entry path you use, it must not occupy a host-reserved namespace such as the host control plane `/api`, the unified plugin `API` namespace `/x`, the public plugin asset namespace `/x-assets`, the legacy plugin asset namespace `/plugin-assets`, or any other reserved path. This configuration is a startup-scoped routing boundary, so host restart is required after changing it.

### How the Workspace Connects to Plugins

Plugins connect their admin pages to the workspace navigation by declaring the `menus` field in `plugin.yaml`:

```yaml
menus:
  - key: plugin:linapro-demo-source:example
    name: Example Management
    path: linapro-demo-source-example
    component: system/plugin/dynamic-page
    perms: linapro-demo-source:example:view
    type: M
```

When responding to frontend menu requests, the host merges built-in menus with menu declarations from enabled plugins, filters them by the current user's permissions, and returns the complete menu tree to the workspace. When a plugin is disabled, its menu entries automatically disappear from navigation, and the workspace does not need a code change.

```mermaid
sequenceDiagram
    participant UI as Admin workspace
    participant Core as lina-core
    participant Plugin as Enabled plugin

    UI->>Core: GET /api/v1/menu
    Core->>Core: Read host menus and user permissions
    Core->>Plugin: Project plugin menu declarations
    Plugin-->>Core: Return plugin menus and button permissions
    Core-->>UI: Return complete menu tree
    UI->>UI: Render sidebar and dynamic routes
```

This mechanism lets each plugin own and maintain its admin interface declaration. The main framework does not need to understand the plugin's business page content; it only aggregates menu data and filters permissions.

### Workspace and Host Responsibilities

The workspace is the standard `UI` consumer of the main framework. It does not define business logic. The main framework provides `API` contracts, and the workspace uses those contracts to display data and expose operations. Plugins extend the framework's APIs, and the workspace hosts plugin pages through the unified menu and dynamic page shell mechanism without custom frontend changes for each plugin.

Developers can replace the built-in workspace with a custom frontend. As long as the new frontend follows the host's public `RESTful API` and permission model, it can use the same backend capabilities.

Workspace page paths, host control-plane APIs, and plugin APIs are three separate boundaries:

| Boundary | Default Path | Description |
|----------|--------------|-------------|
| Default admin workspace | `/admin` | Admin workspace `SPA` entry and refresh `fallback` boundary |
| Host control-plane `API` | `/api/v1/...` | Built-in host governance APIs for users, permissions, configuration, and related system capabilities |
| Unified plugin `API` | `/x/{plugin-id}/...` | Shared plugin `API` namespace for source plugins and dynamic plugins; `/api/v1` is the internal version group used by the official examples |

## Plugin Capability Boundaries

### Dynamic Routes

The two plugin types differ clearly in route registration capability, but their plugin APIs must enter the same host-reserved namespace:

```text
/x/{plugin-id}/...
```

Here, `/x` is the unified plugin `API` namespace, `{plugin-id}` isolates each plugin, and the remaining path segments are owned by the plugin. Official plugins usually continue with `/api/v1` as an internal version group, so their public paths commonly look like `/x/{plugin-id}/api/v1/...`. The boundary enforced by the host, however, is `/x/{plugin-id}`, not a fixed `/api/v1` child path.

**Source plugins** can register non-reserved `HTTP` route paths freely. A plugin declares a route registration callback during `init()`, and the host triggers it during startup. A plugin can register `/`, `/portal/...`, `/assets/...`, or other non-reserved public paths for pages, portals, static assets, custom `fallback` behavior, or other `HTTP` responses. This flexibility is useful, but it also creates a risk: **if multiple source plugins register the same route path, route conflicts cause service startup to fail**. For example, if more than one plugin registers the same `GET /` root route, the `HTTP Server` router reports a duplicate route and the process exits.

Plugin APIs for source plugins must not use arbitrary public paths. They must use the unified plugin `API` namespace. `RouteRegistrar` provides `APIPrefix()`, which returns the plugin namespace assigned to the current plugin:

```text
/x/{plugin-id}
```

:::info Tip
`x` stands for `eXtension`. It is the conventional namespace prefix that marks the path as a plugin extension `API`, not a host control-plane endpoint.
:::

Source plugin business APIs should be mounted under that prefix. The recommended pattern is to add an internal API version group, for example:

```text
/x/linapro-demo-source/api/v1/demo-records
/x/linapro-demo-source/api/v1/demo-records/{id}
```

Source plugins must not register another plugin's path under `/x`. To keep boundaries clear, public pages, portal entries, static assets, health checks, and other non-API routes should not live under `/x/{plugin-id}`. Public pages and plugin-managed static assets should continue to use `/`, `/portal/...`, `/assets/...`, or another non-reserved path.

**Dynamic plugins** run inside a `WASM` sandbox and cannot directly access the `HTTP Server` route registration mechanism. Dynamic plugins declare their internal `path`, `method`, `access`, and `permission` through build-time `RegisterRoutes` metadata and the generated `route contract`. At runtime, the host owns only the `/x/{plugin-id}` prefix; the remaining path comes from the dynamic plugin's route contract. The official dynamic plugin example also declares `/api/v1` as its internal route group, producing public paths such as:

```text
/x/linapro-demo-dynamic/api/v1/demo-records
/x/linapro-demo-dynamic/api/v1/demo-records/{id}
/x/linapro-demo-dynamic/api/v1/backend-summary
```

Source plugin requests are handled by the `HTTP Server` handlers registered by source plugins. Dynamic plugin requests are bridged by the host to the matching runtime based on the `route contract`. Host control-plane APIs still use `/api/v1/...`; neither source plugins nor dynamic plugins should mount plugin business APIs under the host control-plane namespace.

| Plugin Type | Route Registration | Path Constraint | Conflict Risk |
|-------------|--------------------|-----------------|---------------|
| Source plugin `API` | `pluginhost.RouteRegistrar` | Must use the `/x/{plugin-id}` namespace returned by `APIPrefix()`; remaining segments are plugin-defined | Isolated by plugin `ID`, so it does not conflict with another plugin's `/x` namespace |
| Source plugin custom routes | `pluginhost.RouteRegistrar` | Can register non-reserved paths such as `/portal/...` or `/assets/...` | Startup fails if they conflict with host, workspace, or other plugin routes |
| Dynamic plugin `API` | Host dispatches by `route contract` | Host combines `/x/{plugin-id}` with the contract's internal path; official examples use `/api/v1/...` | Avoided by host isolation through plugin `ID`, active artifact, and route contract |

### Static Assets

The main framework manages its own static assets through `Go Embed` and provides a declarative `public_assets` model for serving public plugin assets. Both source plugins and dynamic plugins can declare public asset directories in the root of `plugin.yaml`. The host maps those assets to `/x-assets/{plugin-id}/{version}/...`, for example:

```text
/x-assets/linapro-demo-dynamic/v0.1.0/pages/standalone.html
/x-assets/linapro-demo-dynamic/v0.1.0/css/main.css
```

Each declaration uses `source` to point to a plugin-relative directory or a dynamic `artifact frontend asset` prefix, optional `mount` to specify its relative mount path under `/x-assets/{plugin-id}/{version}/`, and optional `index` to select the default file returned when the mount directory itself is requested. If `mount` is empty or `/`, files in the declared directory are mounted directly at the version root. If `index` is omitted, the host defaults to `index.html`.

```yaml
public_assets:
  # source points to the plugin directory that should be publicly exposed.
  - source: frontend/public
    # mount is the path under /x-assets/{plugin-id}/{version}/.
    # / means the files are mounted directly at the version root.
    mount: /
    # index is the default file used when the mount directory itself is requested.
    # When omitted, it defaults to index.html.
    index: index.html
  # Another asset group can be mounted under a subpath for pages, styles, or other groups.
  - source: frontend/pages
    mount: pages
    index: standalone.html
```

When a source plugin registers an `embedded filesystem` through `plugin.Assets().UseEmbeddedFiles(...)`, the host exposes files only from directories declared in `public_assets`. For example:

| source | mount | File inside the plugin | Public URL |
|--------|-------|------------------------|------------|
| `frontend/public` | `/` | `frontend/public/logo.png` | `/x-assets/{plugin-id}/{version}/logo.png` |
| `frontend/pages` | `pages` | `frontend/pages/standalone.html` | `/x-assets/{plugin-id}/{version}/pages/standalone.html` |

Files outside declared directories are not exposed simply because they exist in the `embedded filesystem`.

Dynamic plugin `frontend assets` must also use the same declarative model. The host no longer exposes dynamic `artifact` frontend assets automatically just because they exist. Only asset sets matching `public_assets` declarations are served through `/x-assets/{plugin-id}/{version}/...`.

`/x-assets` is an optional hosted entry, not a mandatory frontend mechanism for source plugins. Source plugins can still return pages, file streams, static assets, or `SPA fallback` behavior through their own `HTTP` routes. Those routes are fully maintained by plugin code. The host does not infer `HTTP` routes from `public_assets`, and it does not infer asset directories from `HTTP` routes.

For dynamic plugins, workspace menus may reference assets hosted under `/x-assets`, but the menu path does not become the workspace route directly. The current embedded-mount mode usually uses `component: system/plugin/dynamic-page`, points the menu `path` at a `.js` or `.mjs` entry asset, and declares `pluginAccessMode: embedded-mount` in `query`. The host then forwards the hosted asset URL to the dynamic page shell as `embeddedSrc`.

```yaml
menus:
  - key: plugin:linapro-demo-dynamic:main-entry
    name: Dynamic Plugin Demo
    path: /x-assets/linapro-demo-dynamic/v0.1.0/mount.js
    component: system/plugin/dynamic-page
    perms: linapro-demo-dynamic:view
    type: M
    query:
      pluginAccessMode: embedded-mount
```

Plugins can also expose static asset access through dynamic routes, but this is **not recommended**. The unified host asset entry provides governance capabilities that would otherwise need to be rebuilt:

| Host-Managed Capability | Problem When Implemented Manually |
|-------------------------|-----------------------------------|
| Plugin enablement coupling, returning `404` automatically when disabled | Plugin must implement enablement checks itself |
| Automatic `MIME` type derivation | Plugin must maintain its own `Content-Type` map |
| In-memory caching and on-demand serving | Plugin must implement caching itself |
| Version path isolation through `/{version}/...` | Plugin must design its own versioning strategy |
| Consistent priority rules with host asset routes | Custom implementation may conflict with host route ordering |

`public_assets` is an explicit publication boundary set by the plugin author. As long as the declaration points to a real directory in the plugin-owned resource set or a dynamic artifact frontend asset prefix, the host treats it as ordinary public static content. For that reason, plugin authors must declare only files that are safe for anonymous access. Do not put governance metadata, installation scripts, private configuration, tenant-specific files, or user-specific files into `public_assets`.

The host still applies strict path validation. The following declarations are rejected:

| Invalid Configuration | Reason |
|-----------------------|--------|
| Empty `source` | No clear publication boundary can be formed |
| Absolute paths or `URL`s | They may escape the plugin-owned resource set |
| Paths that normalize to `../` traversal | They may escape the declared directory or plugin root |
| Paths containing wildcards, query strings, or fragments | They cannot be mapped to a stable static directory |
| Duplicate or overlapping `mount` values | A single public URL could map to multiple sources |
| Missing `source` directory or prefix | Source plugin directories and dynamic frontend asset prefixes must exist |
| Symlinks that escape the plugin root | They may read files outside the plugin |
| `index` values that are not plain file names | Directory defaults must be safe relative file names |

A `public asset URL` uses `{plugin-id, version}` as the cache boundary. Asset content under the same plugin version must remain stable. If asset content changes, the plugin should upgrade its `plugin.yaml` version or introduce an equivalent content versioning mechanism. When a plugin is not installed, not enabled, or not available to the current tenant, `/x-assets/{plugin-id}/{version}/...` returns `404` by default. Dynamic plugins may continue serving installed or active versioned assets while the plugin remains enabled; source plugins resolve declared resources from the plugin assets compiled into the host or from the plugin directory.

### Foundational Capabilities

The main framework exposes foundational capabilities to plugins through stable service adapter interfaces. Plugins can call these interfaces without knowing the host's internal implementation.

**Source plugins** access host foundational capabilities through the `HostServices` interface, provided consistently by `pluginhost.HTTPRegistrar` and `pluginhost.CronRegistrar`:

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    svc := registrar.HostServices()
    // Use host-provided i18n, tenant filtering, cache, and related capabilities.
    _ = svc.I18n()
    _ = svc.TenantFilter()
    _ = svc.Cache()
    _ = svc.Config()
    _ = svc.Notify()
    // ...
    return nil
}
```

Current `HostServices` covers the following service domains:

| Service | Capability |
|---------|------------|
| `APIDoc()` | API documentation localization adapter |
| `Auth()` | Tenant authentication adapter |
| `BizCtx()` | Business context adapter |
| `Cache()` | Plugin-scoped cache adapter |
| `Config()` | Static configuration reader adapter |
| `I18n()` | Runtime translation adapter |
| `Notify()` | Notification sender adapter |
| `PluginState()` | Plugin enablement state adapter |
| `PluginLifecycle()` | Plugin lifecycle orchestration adapter |
| `Route()` | Dynamic route metadata adapter |
| `Session()` | Online session adapter |
| `TenantFilter()` | Tenant filter adapter |

**Dynamic plugins** access host capabilities through the `host_call` mechanism exposed by `pluginbridge`. Every call is validated against the authorization snapshot. A dynamic plugin declares `hostServices` in `plugin.yaml` to request the services and methods it needs. The host records that authorization snapshot at installation time, then validates the service, method, and resource boundary for every runtime `host_call`:

```yaml
hostServices:
  - service: data
    methods: [list, get, create, update, delete]
    resources:
      tables:
        - plugin_demo_dynamic_record
  - service: cache
    methods: [get, set, delete]
  - service: runtime
    methods: [log.write, info.uuid, info.now]
```

Current host service capabilities available to dynamic plugins include:

| Service | Capability |
|---------|------------|
| `runtime` | Runtime logs, plugin-scoped runtime state read/write/delete, host time, `UUID`, node information, and other runtime metadata |
| `cron` | Dynamic plugin scheduled task registration |
| `storage` | File write, read, delete, list, and metadata operations constrained by authorized paths |
| `network` | Outbound `HTTP` requests constrained by authorized `URL` patterns |
| `data` | List, detail, create, update, delete, and transaction operations constrained by authorized data tables |
| `cache` | Plugin cache read, write, delete, increment, and expiry adjustment |
| `lock` | Distributed lock acquire, renew, and release |
| `notify` | Message notification sending |
| `config` | Read-only configuration reads, existence checks, and typed values |

For `storage`, accessible paths are constrained by `paths`; for `data`, accessible tables are constrained by `tables`; for `network`, accessible targets are constrained by declared `URL` resource patterns. Other services continue to be controlled by `methods`.

The main framework's foundational capabilities will continue to expand across versions. Plugins can use only the capabilities they actually need without depending on host internal service implementations.

## Portal Routes and Admin Routes

When implementing business capabilities, plugins often need two kinds of interfaces: **portal routes** for frontend users (the data plane), and **admin routes** for backend management (the control plane). Their callers, permission models, and interface semantics differ significantly.

| Dimension | Portal Routes (Data Plane) | Admin Routes (Control Plane) |
|-----------|----------------------------|------------------------------|
| **Caller** | Frontend users and anonymous visitors | Administrators and backend operators |
| **Authentication** | Can be public or require login as needed | Usually requires login and permissions |
| **Interface semantics** | Content viewing and business operations | Data management and configuration changes |
| **Menu mounting** | Not necessarily mounted in the workspace | Usually accessed through workspace menus |

The main framework does not interpret a plugin's internal route grouping and does not distinguish the plugin's control plane from its data plane. Plugins own this internal grouping entirely. Note that `/x/{plugin-id}` carries the plugin `API` namespace only. If a plugin uses `/api/v1` as its internal version group, portal APIs and admin APIs can live under `/x/{plugin-id}/api/v1/...`. Public pages, portal entries, static assets, or custom `fallback` routes for source plugins are not plugin APIs and should continue to use non-reserved paths such as `/portal/*` or `/assets/*`. The following example shows a source plugin registering a public portal entry, portal APIs, and admin APIs together:

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    var (
        routes      = registrar.Routes()
        middlewares = routes.Middlewares()
        apiPrefix   = registrar.APIPrefix()
    )
    // Plugin APIs: both source plugins and dynamic plugins must enter the
    // /x/{plugin-id} namespace. Here, /api/v1 is the plugin's internal API version group.
    routes.Group(apiPrefix, func(group pluginhost.RouteGroup) {
        group.Group("/api/v1", func(group pluginhost.RouteGroup) {
            group.Middleware(
                middlewares.NeverDoneCtx(),
                middlewares.HandlerResponse(),
                middlewares.CORS(),
                middlewares.RequestBodyLimit(),
                middlewares.Ctx(),
            )

            // Portal API (data plane): for frontend users; some endpoints can be anonymous.
            // Full path example: /x/my-plugin/api/v1/portal/articles.
            group.Group("/portal", func(group pluginhost.RouteGroup) {
                // Public endpoint: no login required.
                group.Bind(
                    portalArticleCtrl,
                )
                // Login endpoint: requires authentication but not admin permission.
                group.Group("/", func(group pluginhost.RouteGroup) {
                    group.Middleware(
                        middlewares.Auth(),
                    )
                    group.POST("/comments", commentCtrl.Create)
                })
            })

            // Admin API (control plane): for the admin workspace; requires full auth and permission checks.
            // Full path example: /x/my-plugin/api/v1/admin/articles.
            group.Group("/admin", func(group pluginhost.RouteGroup) {
                group.Middleware(
                    middlewares.Auth(),
                    middlewares.Tenancy(),
                    middlewares.Permission(),
                )
                group.Bind(
                    adminArticleCtrl,
                    adminCommentCtrl,
                )
            })
        })
    })

    // Public portal entry: plugin-managed pages, static assets, or SPA fallback.
    // These routes are not plugin APIs, do not enter /x, and are not automatically projected as menus, permissions, or OpenAPI entries.
    routes.Group("/portal", func(group pluginhost.RouteGroup) {
        group.Middleware(
            middlewares.NeverDoneCtx(),
            middlewares.CORS(),
            middlewares.RequestBodyLimit(),
        )
        group.Bind(
            portalPageCtrl,
            portalAssetCtrl,
        )
    })

    return nil
}
```

In this example, `/portal` is a plugin-managed public portal entry, `/x/my-plugin/api/v1/portal/articles` is a plugin portal `API`, and `/x/my-plugin/api/v1/admin/articles` is a plugin admin `API`. The admin workspace sees plugin menu entries only through `menus` in `plugin.yaml`; it must not automatically generate workspace routes, menus, or permission nodes just because a plugin registered `HTTP` routes. Permission identifiers for admin APIs are declared through the `permission` field in `g.Meta`, then associated with menu entries and permission items in `plugin.yaml`, and finally shown to administrators through the workspace extension center:

```go
// Example admin endpoint DTO.
type ArticleAdminListReq struct {
    g.Meta   `path:"/articles" method:"get" tags:"My Plugin Admin" summary:"List admin articles" permission:"my-plugin:article:view"`
    Page     int `json:"page"`
    PageSize int `json:"pageSize"`
}
```

## Why These Boundaries Matter

The capability boundary design between the main framework and plugins is not only about today's feature division. It also shapes how the whole ecosystem evolves over time.

**For the main framework**, stable extension interfaces mean internal implementation can evolve safely. The host can refactor concrete `HostServices` implementations, optimize plugin governance flows, and add new foundational capabilities. As long as the public `pluginhost` and `pluginbridge` contracts remain stable, existing plugins are not affected.

**For plugin developers**, clear boundaries mean host capabilities can be used with confidence. Developers do not need to understand host internals or worry that future host upgrades will break plugin logic. If a plugin collaborates with the host strictly through stable contracts, version compatibility is protected by the host's interface stability.

**For the system as a whole**, this layered architecture allows different teams to maintain the main framework and their own plugins independently, reducing coordination friction and leaving enough room for future plugin types or governance capabilities.
