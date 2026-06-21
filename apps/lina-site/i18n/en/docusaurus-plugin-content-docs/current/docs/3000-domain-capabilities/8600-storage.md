---
slug: '/docs/domain-capability-storage'
title: 'Storage'
hide_title: true
description: '`Storage()` provides a plugin-scoped object storage sandbox for source plugins and dynamic plugins, supporting object read/write, listing, deletion, and metadata queries. The host ensures security through path authorization and plugin isolation, ships a built-in local disk provider, and supports registering custom storage backends via the Provider interface. Supports chunked uploads for large objects, and automatically cleans up all objects under authorized paths when a plugin is uninstalled.'
keywords:
  - Storage capability
  - storagecap
  - hostServices.storage
  - storage.put
  - storage.get
  - storage.list
  - storage.delete
  - storage.stat
  - Provider
  - ProviderStatuses
  - object storage
  - plugin storage
  - path authorization
  - capability-storage
  - plugin capability
  - chunked upload
  - LinaPro
---

## Introduction

Source plugins use object storage through `services.Storage()`. Dynamic plugins declare `service: storage` in `plugin.yaml` and access it through the `pluginbridge.Default().Storage()` client.

The `Storage` capability provides each plugin with an independent object storage sandbox. Plugins read and write objects under their declared authorized path prefixes, while the host handles path validation, plugin isolation, and storage backend management.

**Capability Phase**: Runtime

**Supported Plugin Types**: Source plugins, Dynamic plugins

## Capability Design

### Sandbox Isolation Model

`Storage` uses dual plugin + tenant isolation. Each plugin's objects are automatically scoped under the `plugins/{pluginID}/` prefix, with tenant data further isolated in `tenant/{tenantID}/` sub-paths. Platform-level data uses the `platform/` sub-path:

```mermaid
graph TB
    Source["Source Plugin"] --> Storage["services.Storage()"]
    Dynamic["Dynamic Plugin"] --> Guest["pluginbridge.Default().Storage()"]
    Storage --> Adapter["storageAdapter"]
    Guest --> Adapter
    Adapter --> Key["Object Key Mapping"]
    Key --> TenantKey["plugins/{pluginID}/tenant/{tenantID}/{path}"]
    Key --> PlatformKey["plugins/{pluginID}/platform/{path}"]
    TenantKey --> Provider["Storage Provider"]
    PlatformKey --> Provider
    Provider --> Local["Local Disk"]
    Provider --> Custom["Custom Provider"]
```

### Object Key Mapping

Logical paths used by plugins are automatically mapped to physical storage object keys by `storageAdapter`. The mapping rules are:

| Scope | Object Key Format |
|-------|-------------------|
| Tenant-level | `plugins/{pluginID}/tenant/{tenantID}/{logicalPath}` |
| Platform-level | `plugins/{pluginID}/platform/{logicalPath}` |

Plugins only need to work with logical paths (e.g., `exports/report.csv`) and don't need to worry about the underlying object key structure.

### Object Metadata

Plugin-visible object metadata does not expose physical storage paths, `Provider` keys, or host file management IDs:

| Field | Description |
|-------|-------------|
| `Path` | Logical path |
| `Size` | Object size |
| `ContentType` | Content type |
| `ETag` | Entity tag, computed as `SHA-1(key + size + modtime)` |
| `UpdatedAt` | Last update time |
| `Visibility` | Visibility flag: `private` or `public` |

Object visibility is controlled by the host. The `Visibility` field indicates whether an object can be publicly served. Plugins should not assume that written objects are automatically publicly accessible — the specific access policy is determined by the host's service layer.

### Content Type Detection

When writing an object, `storageAdapter` detects content type in the following priority order:

1. Explicitly specified `contentType` in the request
2. Body sniffing (reads the first `512` bytes for detection)
3. File extension inference
4. Fallback to `application/octet-stream`

### Storage Provider Architecture

`Storage` supports a pluggable storage `Provider` architecture:

| Provider | Description |
|----------|-------------|
| `local` | Built-in local disk `Provider`, stores under the `.capability-storage/` directory |
| Custom | Plugin `Provider` registered via `Provide()`, supporting `OSS`, `S3`, `MinIO`, etc. |

The active `Provider` is selected via the `plugin.storage.activeProviderPluginId` configuration. When unconfigured, the local `Provider` is used. In cluster mode, the local `Provider` refuses service by default — `allowLocalProviderInCluster` must be explicitly set.

#### Provider Registration Mechanism

`Provider`s are managed through a process-level global registry. Source plugins call `storagecap.Provide(pluginID, factory)` to register a `ProviderFactory` function. The host resolves the currently active `Provider` at runtime via `ResolveProvider`:

- When `activeProviderPluginId` is not configured, the built-in local `Provider` is used
- When `activeProviderPluginId` is configured, it must match a registered and available plugin `Provider` — no silent fallback

### Difference from Files Capability

| Dimension | Files | Storage |
|-----------|-------|---------|
| Purpose | Read-only view of host file management system | Plugin-scoped object storage sandbox |
| Database | `sys_file` table, full metadata | No database, pure object storage |
| Isolation | Tenant + data scope | Plugin + tenant path isolation |
| Operations | Read-only view + controlled deletion | Full CRUD |
| Size limit | Configured by `upload.maxSize` | Unlimited |
| `Provider` | Built-in local storage | Pluggable `Provider` registration |

## Interface Definitions

### Source Plugin Interface

| Method | Description |
|--------|-------------|
| `Put` | Write an object, with `ContentType` and `Overwrite` control |
| `Get` | Read object content and metadata |
| `Delete` | Delete an object under an authorized path |
| `DeleteMany` | Batch-delete objects under authorized paths |
| `List` | List objects by prefix |
| `ListCursor` | List objects by prefix with cursor pagination |
| `Stat` | Read object metadata without returning content |
| `BatchStat` | Batch-read object metadata |
| `ProviderStatuses` | Query all registered `Provider` statuses (source plugins only) |

The `Overwrite` parameter of `Put` controls overwrite behavior: when set to `false`, a `PLUGIN_STORAGE_OBJECT_EXISTS` error is returned if the object already exists.

### Dynamic Plugin Interface

| Dynamic Method | Dynamic `SDK` Method | Description |
|----------------|---------------------|-------------|
| `put` | `Storage().Put` | Write an object, with `ContentType` and `Overwrite` control |
| `put.init` | — | Initialize a chunked upload session, returns an upload ID |
| `put.chunk` | — | Write chunk data sequentially by offset |
| `put.commit` | — | Commit chunked upload, merging into the final object |
| `put.abort` | — | Cancel chunked upload, clean up temporary files |
| `get` | `Storage().Get` | Read object content and metadata |
| `delete` | `Storage().Delete` | Delete an object under an authorized path |
| `delete_many` | `Storage().DeleteMany` | Batch-delete objects under authorized paths |
| `list` | `Storage().List` | List objects by prefix |
| `list_cursor` | `Storage().ListCursor` | List objects by prefix with cursor pagination |
| `stat` | `Storage().Stat` | Read object metadata without returning content |
| `batch_stat` | `Storage().BatchStat` | Batch-read object metadata |

The dynamic plugin `Guest SDK` automatically selects the upload mode: objects no larger than `1 MB` use direct upload (single call), while objects exceeding `1 MB` or with unknown size automatically switch to chunked upload (`1 MB` chunks). The host-side maximum chunk size is `4 MB`, and sessions are valid for `15` minutes. On chunk failure, the system automatically attempts `abort` to clean up temporary files.

`ProviderStatuses` cannot be used through the dynamic plugin transport protocol.

## Capability Usage

### Source Plugin Usage

Source plugins operate objects directly through `services.Storage()`:

```go
// Write an object
_, err := services.Storage().Put(ctx, storagecap.PutInput{
    Path:        "exports/report.csv",
    Body:        reader,
    ContentType: "text/csv",
    Overwrite:   true,
})

// Read an object
output, err := services.Storage().Get(ctx, storagecap.GetInput{
    Path: "exports/report.csv",
})

// Delete an object
err := services.Storage().Delete(ctx, storagecap.DeleteInput{
    Path: "exports/report.csv",
})

// Batch-delete objects
err := services.Storage().DeleteMany(ctx, storagecap.DeleteManyInput{
    Paths: []string{"exports/report1.csv", "exports/report2.csv"},
})

// List objects
list, err := services.Storage().List(ctx, storagecap.ListInput{
    Prefix: "exports/",
    Limit:  100,
})

// List objects with cursor pagination
cursorList, err := services.Storage().ListCursor(ctx, storagecap.ListCursorInput{
    Prefix: "exports/",
    Cursor: lastCursor,
    Limit:  100,
})

// Query object metadata
stat, err := services.Storage().Stat(ctx, storagecap.StatInput{
    Path: "exports/report.csv",
})

// Batch-query object metadata
batchStat, err := services.Storage().BatchStat(ctx, storagecap.BatchStatInput{
    Paths: []string{"exports/report1.csv", "exports/report2.csv"},
})

// Query Provider statuses
statuses, err := services.Storage().ProviderStatuses(ctx)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `storage` service and authorized paths in `plugin.yaml`:

```yaml
hostServices:
  - service: storage
    methods:
      - put
      - get
      - delete
      - delete_many
      - list
      - list_cursor
      - stat
      - batch_stat
    resources:
      paths:
        - exports/
        - temp/reports/
```

Authorization granularity is the logical path prefix. All request paths undergo normalization and authorization validation at the `WASM` host service layer, ensuring plugins can only access declared path scopes.

Usage on the dynamic plugin side:

```go
storageSvc := pluginbridge.Default().Storage()

// Write an object (small objects use direct upload)
_, err := storageSvc.Put(ctx, storagecap.PutInput{
    Path:        "exports/report-2024.csv",
    Body:        data,
    ContentType: "text/csv",
})

// Read an object
output, err := storageSvc.Get(ctx, storagecap.GetInput{
    Path: "exports/report-2024.csv",
})

// Delete an object
err := storageSvc.Delete(ctx, storagecap.DeleteInput{
    Path: "exports/report-2024.csv",
})

// Batch-delete objects
err := storageSvc.DeleteMany(ctx, storagecap.DeleteManyInput{
    Paths: []string{"exports/report1.csv", "exports/report2.csv"},
})

// List objects
list, err := storageSvc.List(ctx, storagecap.ListInput{
    Prefix: "exports/",
})

// List objects with cursor pagination
cursorList, err := storageSvc.ListCursor(ctx, storagecap.ListCursorInput{
    Prefix: "exports/",
    Cursor: lastCursor,
    Limit:  100,
})

// Query object metadata
stat, err := storageSvc.Stat(ctx, storagecap.StatInput{
    Path: "exports/report-2024.csv",
})

// Batch-query object metadata
batchStat, err := storageSvc.BatchStat(ctx, storagecap.BatchStatInput{
    Paths: []string{"exports/report1.csv", "exports/report2.csv"},
})
```

## System Constraints

| Constraint | Limit |
|------------|-------|
| Single object size | Unlimited |
| Logical path length | `512` bytes |
| List default limit | `100` items |
| List maximum limit | `1000` items |
| Direct upload threshold | `1 MB` (Guest SDK auto-switches to chunked) |
| Chunk size (Guest) | `1 MB` |
| Chunk size (Host) | `4 MB` |
| Chunk session validity | `15` minutes |

## Design Constraints

- **Paths are not physical paths.** `paths` are logical authorization scopes. Plugins cannot escape the authorized prefix through relative paths. The `WASM` host service layer performs normalization and authorization validation on every request path.
- **Object visibility is controlled by the host.** Whether an object can be publicly served is determined by host metadata and subsequent service policies. Plugins should not assume that written objects are automatically publicly accessible.
- **No underlying details exposed.** Object metadata does not include physical paths, `Provider` keys, or host file management IDs.
- **Automatic cleanup on plugin uninstall.** When a plugin is uninstalled, the host enumerates and batch-deletes all objects under the authorized path prefix.
- **No silent Provider fallback.** Once a custom `Provider` is configured, if that `Provider` is unavailable, operations fail directly rather than falling back to the local `Provider`.

## Related Services

- [Files Capability](/docs/domain-capability-files)
- [Manifest Resources Capability](/docs/domain-capability-manifest)
- [Record Store Capability](/docs/domain-capability-recordstore)
