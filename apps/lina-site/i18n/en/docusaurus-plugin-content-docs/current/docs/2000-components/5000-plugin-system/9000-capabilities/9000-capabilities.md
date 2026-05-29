---
slug: '/docs/plugin-capability-services'
title: 'Plugin Capability Services Overview'
hide_title: true
description: "the design principles behind the Services interface, service classification, access patterns, and a scenario-based guide for choosing the right service, helping plugin developers understand each capability service's positioning, boundaries, and collaboration patterns."
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

## Service Quick Reference

| Category | Service | Contract Type | Description |
|----------|---------|---------------|-------------|
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

## How to Obtain Services

### Source Plugins

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

### Dynamic Plugins

Dynamic plugins access core framework capabilities through the `guest` side SDK in `pluginbridge`. All service calls are bridged through `hostServices` and must be declared in `plugin.yaml` for authorization:

```yaml
hostServices:
  - service: config
    methods: [get]
  - service: hostConfig
    methods: [get]
    resources:
      keys:
        - workspace.basePath
        - i18n.default
  - service: manifest
    methods: [get]
    resources:
      paths:
        - profile.yaml
  - service: cache
    methods: [get, set, delete, incr, expire]
  - service: i18n
    methods: [translate, getLocale]
  - service: notify
    methods: [send]
```

The convenience functions provided by the `guest` side SDK (`Config.String`, `Config.Duration`, `I18n.Translate`, `Manifest.Scan`, etc.) are automatically translated into the corresponding `hostServices` calls at runtime — you do not need to declare these convenience function names as `methods` in `plugin.yaml`. For example, declaring `methods: [get]` is sufficient to use `Config.String`, `Config.Bool`, `Config.Int`, `Config.Duration`, `Config.Scan`, and all other read functions.

In the WASM entry function or controller, obtain services through the `guest` side capabilities:

```go
// Read configuration via guest side SDK
endpoint, err := guestsdk.Config.String(ctx, "sync.endpoint", "")

// Read host configuration via guest side SDK
workspaceBase, err := guestsdk.HostConfig.String(ctx, "workspace.basePath", "/admin")

// Read manifest resources via guest side SDK
var profile struct {
    Category string `yaml:"category"`
}
err := guestsdk.Manifest.Scan(ctx, "profile.yaml", "", &profile)

// Read translations via guest side SDK
message := guestsdk.I18n.Translate(ctx, "plugin.record.created", "Record created")

// Operate cache via hostServices
err := guestsdk.Cache.Set(ctx, "stats", "visit-count", "100", 0)

// Extract context from BridgeRequestEnvelopeV1
userID := requestEnvelope.UserID
tenantID := requestEnvelope.TenantID
```

For dynamic plugins, `BizCtx` context information is extracted from the `BridgeRequestEnvelopeV1` request envelope rather than obtained through `services.BizCtx()`. Fields like `UserID` and `TenantID` in the request envelope are injected by the core framework before entering the WASM sandbox, and their semantics are consistent with the source plugin's `BizCtxService.Current()`.

## Code Examples

The following examples demonstrate common patterns for using each capability service in both source plugins and dynamic plugins.

### Reading Configuration

#### Source Plugin

Source plugins call `services.Config()` and `services.HostConfig()` directly:

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.Services()

    syncEndpoint, err := services.Config().String(ctx, "sync.endpoint", "")
    if err != nil {
        return err
    }

    timeout, err := services.Config().Duration(ctx, "sync.timeout", 30*time.Second)
    if err != nil {
        return err
    }

    workspaceBase, err := services.HostConfig().String(ctx, "workspace.basePath", "/admin")
    if err != nil {
        return err
    }

    _ = syncEndpoint
    _ = timeout
    _ = workspaceBase
    return nil
}
```

#### Dynamic Plugin

Dynamic plugins read through the `guest` side SDK and require `config` and `hostConfig` authorization in `plugin.yaml`:

```go
syncEndpoint, err := guestsdk.Config.String(ctx, "sync.endpoint", "")
if err != nil {
    return err
}

timeout, err := guestsdk.Config.Duration(ctx, "sync.timeout", 30*time.Second)
if err != nil {
    return err
}

workspaceBase, err := guestsdk.HostConfig.String(ctx, "workspace.basePath", "/admin")
if err != nil {
    return err
}
```

### Cache Operations

#### Source Plugin

Source plugins operate the cache through `services.Cache()`:

```go
func (s *recordService) GetVisitCount(ctx context.Context, recordID int) (int, error) {
    services := s.services
    val, err := services.Cache().Get(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    if err != nil {
        return 0, err
    }
    count, _ := strconv.Atoi(val)
    return count, nil
}

func (s *recordService) IncrVisitCount(ctx context.Context, recordID int) error {
    services := s.services
    _, err := services.Cache().Incr(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    return err
}

func (s *recordService) CacheSummary(ctx context.Context, recordID int, summary string) error {
    services := s.services
    return services.Cache().Set(ctx, "summary", fmt.Sprintf("record:%d", recordID), summary, 10*time.Minute)
}
```

#### Dynamic Plugin

Dynamic plugins operate the cache through `hostServices` with `cache` authorization declared in `plugin.yaml`:

```go
func GetVisitCount(ctx context.Context, recordID int) (int, error) {
    val, err := guestsdk.Cache.Get(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    if err != nil {
        return 0, err
    }
    count, _ := strconv.Atoi(val)
    return count, nil
}

func IncrVisitCount(ctx context.Context, recordID int) error {
    _, err := guestsdk.Cache.Incr(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    return err
}
```

### I18n Translation

#### Source Plugin

Source plugins translate text through `services.I18n()`:

```go
func (s *recordService) CreateRecord(ctx context.Context, req *CreateRecordReq) error {
    services := s.services

    // Translation key "plugin.record.created" defined in plugin language pack manifest/i18n/zh-CN/plugin.json
    msg := services.I18n().Translate(ctx, "plugin.record.created", "Record created")

    // Use translated text when publishing notifications
    return services.Notify().SendNoticePublication(ctx, &contract.NoticePublication{
        SourceType: "plugin",
        Title:      msg,
        Content:    fmt.Sprintf("Record ID: %d", req.ID),
    })
}
```

#### Dynamic Plugin

Dynamic plugins translate through the `guest` side SDK and require `service: i18n` authorization in `plugin.yaml`:

```go
func CreateRecord(ctx context.Context, req *CreateRecordReq) error {
    msg := guestsdk.I18n.Translate(ctx, "plugin.record.created", "Record created")

    // Publish notification
    return guestsdk.Notify.SendNoticePublication(ctx, &notify.Publication{
        SourceType: "plugin",
        Title:      msg,
        Content:    fmt.Sprintf("Record ID: %d", req.ID),
    })
}
```

### Publishing Notifications

#### Source Plugin

Source plugins publish notifications through `services.Notify()`:

```go
func (s *recordService) NotifyRecordCreated(ctx context.Context, record *Record) error {
    services := s.services

    title := services.I18n().Translate(ctx, "plugin.record.created.title", "New Record")
    content := services.I18n().Translate(
        ctx,
        "plugin.record.created.content",
        fmt.Sprintf("Record \"%s\" has been created", record.Title),
    )

    return services.Notify().SendNoticePublication(ctx, &contract.NoticePublication{
        SourceType:   "plugin",
        CategoryCode: "record_created",
        Title:        title,
        Content:      content,
    })
}
```

#### Dynamic Plugin

Dynamic plugins publish notifications through the `guest` side SDK and require `service: notify` authorization in `plugin.yaml`:

```go
func NotifyRecordCreated(ctx context.Context, record *Record) error {
    title := guestsdk.I18n.Translate(ctx, "plugin.record.created.title", "New Record")
    content := guestsdk.I18n.Translate(
        ctx,
        "plugin.record.created.content",
        fmt.Sprintf("Record \"%s\" has been created", record.Title),
    )

    return guestsdk.Notify.SendNoticePublication(ctx, &notify.Publication{
        SourceType:   "plugin",
        CategoryCode: "record_created",
        Title:        title,
        Content:      content,
    })
}
```

### Tenant-Aware Data Operations

#### Source Plugin

Source plugins implement tenant-aware data operations through `TenantFilterService`. This service is extended in `pluginhost.Services` and is not part of `capability.Services`. The DAO layer is generated by `make dao`, providing `dao.Xxx` singletons and `do.Xxx` / `entity.Xxx` model types.

**Injecting `TenantFilterService` in the service layer:**

```go
type serviceImpl struct {
    tenantFilter plugincontract.TenantFilterService
}

func New(tenantFilter plugincontract.TenantFilterService) Service {
    return &serviceImpl{tenantFilter: tenantFilter}
}
```

Obtain and inject it in the plugin registration entry:

```go
services := registrar.Services()
recordSvc := recordsvc.New(services.TenantFilter())
```

**Querying lists** — use `Apply()` to append `tenant_id` conditions:

```go
func (s *serviceImpl) List(ctx context.Context, pageNum, pageSize int) ([]*entitymodel.Record, int, error) {
    model := s.tenantFilter.Apply(ctx, dao.Record.Ctx(ctx), "")

    total, err := model.Count()
    if err != nil {
        return nil, 0, err
    }

    var items []*entitymodel.Record
    err = model.OrderDesc(dao.Record.Columns().UpdatedAt).
        Page(pageNum, pageSize).
        Scan(&items)
    if err != nil {
        return nil, 0, err
    }

    return items, total, nil
}
```

For join queries, use the `qualifier` parameter to add a table qualifier to the `tenant_id` column, avoiding column name ambiguity:

```go
func (s *serviceImpl) ListWithAuthor(ctx context.Context) ([]*RecordWithAuthor, error) {
    model := dao.Record.Ctx(ctx).
        LeftJoin("sys_user u", "u.id = "+dao.Record.Columns().AuthorId)

    // qualifier is "r", generating r.tenant_id condition
    model = s.tenantFilter.Apply(ctx, model, "r")

    var items []*RecordWithAuthor
    err := model.Scan(&items)
    return items, err
}
```

**Creating records** — obtain the tenant ID through `Context()` and write it explicitly:

```go
func (s *serviceImpl) Create(ctx context.Context, in *CreateRecordInput) (int64, error) {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    recordID, err := dao.Record.Ctx(ctx).Data(do.Record{
        TenantId: tenantID,
        Title:    strings.TrimSpace(in.Title),
        Content:  in.Content,
    }).InsertAndGetId()
    if err != nil {
        return 0, err
    }

    return recordID, nil
}
```

**Updating and deleting** — manually add `TenantFilterColumn` conditions alongside the primary key to scope operations:

```go
func (s *serviceImpl) Update(ctx context.Context, in *UpdateRecordInput) error {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    _, err := dao.Record.Ctx(ctx).
        Where(plugincontract.TenantFilterColumn, tenantID).
        Where(do.Record{Id: in.Id}).
        Data(do.Record{
            Title:   strings.TrimSpace(in.Title),
            Content: in.Content,
        }).
        Update()
    return err
}

func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    _, err := dao.Record.Ctx(ctx).
        Where(plugincontract.TenantFilterColumn, tenantID).
        Where(do.Record{Id: id}).
        Delete()
    return err
}
```

#### Dynamic Plugin

Dynamic plugins work differently — due to sandbox permission controls, they perform database queries through `hostServices` with `data` authorization. Tenant filtering is automatically injected by the core framework at the data access layer:

```yaml
# plugin.yaml
hostServices:
  - service: data
    methods: [list, get, create, update, delete]
    resources:
      tables:
        - plugin_demo_dynamic_record
```

Example database operations in a dynamic plugin:

```go
type recordDAO struct{}

func (d *recordDAO) ListByTenant(ctx context.Context, page, pageSize int) ([]*Record, error) {
    result, err := guestsdk.Data.List(ctx, &data.ListRequest{
        Table:    "plugin_demo_dynamic_record",
        Page:     page,
        PageSize: pageSize,
    })
    if err != nil {
        return nil, err
    }

    var records []*Record
    if err := result.Scan(&records); err != nil {
        return nil, err
    }
    return records, nil
}
```

## Related Content

import DocCardList from '@theme/DocCardList';

<DocCardList />
