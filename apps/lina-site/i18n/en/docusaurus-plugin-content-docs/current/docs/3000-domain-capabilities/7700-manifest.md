---
slug: '/docs/domain-capability-manifest'
title: 'Manifest'
hide_title: true
description: '`ManifestService` provides read-only access to raw resources under the current plugin's `manifest/` directory, supporting byte reading, existence checking, and `YAML` resource scanning. Source plugins consume it through `services.Manifest()`, while dynamic plugins consume it through `hostServices.manifest` after declaring authorized paths.'
keywords:
  - ManifestService
  - manifestcap
  - manifest directory
  - manifest.get
  - manifest resources
  - profile.yaml
  - config.example.yaml
  - i18n resources
  - SQL scripts
  - plugin resources
  - embedded files
  - Wasm artifact
  - hostServices
  - plugin capability
  - LinaPro
---

## Introduction

`services.Manifest()` returns the current plugin-scoped `manifest` resource reading service. Paths are always relative to the `manifest/` directory — for example, to read `manifest/profile.yaml`, pass `profile.yaml`.

Dynamic plugins use the same semantics but must declare the `manifest` service and allowed `paths` in `plugin.yaml`.

**Capability Phase**: Runtime

**Supported Plugin Types**: Source plugins, Dynamic plugins

## Capability Design

### Resource Binding Mechanism

Resources are bound to the current plugin: for source plugins they come from the embedded filesystem, for dynamic plugins from the published artifact or host-bound artifact resources. Paths must be canonical, using forward slashes, and escaping the current plugin's `manifest/` root through relative paths is prohibited.

### Resource Types

| Path Example | Description |
|--------------|-------------|
| `profile.yaml` | Plugin profile or display metadata |
| `config/config.example.yaml` | Config template, not used for runtime default reads |
| `config/config.yaml` | One of the plugin's default config sources |
| `i18n/zh-CN/plugin.json` | Chinese language resources |
| `sql/install.sql` | Install script raw resource |

```mermaid
graph TB
    Plugin["Plugin Code"] --> Service["ManifestService"]
    Service --> Scope["Current plugin manifest/"]
    Scope --> Profile["profile.yaml"]
    Scope --> Config["config/"]
    Scope --> I18n["i18n/"]
    Scope --> SQL["sql/"]
```

### Read-Only Raw Resource Semantics

Reading resources does not mean executing `SQL`, registering language packs, or applying configuration. Configuration reading has dedicated capabilities. Runtime config values use `Plugins().Config()` or dynamic `config.get`; do not read `config/config.yaml` directly to bypass priority.

## Interface Definitions

### Source Plugin Interface

| Method | Description |
|--------|-------------|
| `Get` | Reads the raw byte content at a specified path |
| `GetMany` | Batch-reads raw resource content for a set of specified paths |
| `List` | Returns resource metadata list under a specified prefix |
| `Exists` | Checks whether a resource at a specified path exists |
| `Scan` | Scans a `YAML` resource or nested key within it into the target struct |

### Dynamic Plugin Interface

| Dynamic Method | Dynamic `SDK` Method | Description |
|----------------|---------------------|-------------|
| `get` | `Manifest().Get`, `Manifest().GetMany`, `Manifest().List`, `Manifest().Exists`, `Manifest().Scan` | Reads raw resources under authorized paths |

## Capability Usage

### Source Plugin Usage

Source plugins read resources shipped with the plugin through `services.Manifest()`:

```go
// Read plugin profile
content, err := services.Manifest().Get(ctx, "profile.yaml")

// Batch-read multiple resources
result, err := services.Manifest().GetMany(ctx, manifestcap.GetManyInput{
    Paths: []string{"profile.yaml", "config/config.yaml"},
})

// List resource metadata
listResult, err := services.Manifest().List(ctx, manifestcap.ListInput{
    Prefix: "i18n/",
    Limit:  50,
})

// Check if a resource exists
exists, err := services.Manifest().Exists(ctx, "i18n/zh-CN/plugin.json")

// Scan YAML resource into struct
var config PluginConfig
err := services.Manifest().Scan(ctx, "config/config.yaml", "", &config)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `manifest` service and authorized paths in `plugin.yaml`:

```yaml
hostServices:
  - service: manifest
    methods:
      - get
    resources:
      paths:
        - profile.yaml
        - i18n/zh-CN/plugin.json
```

When `manifest` omits `methods`, it defaults to `get`. `resources.paths` must be paths relative to `manifest/`, not `manifest/profile.yaml`. Usage on the dynamic plugin side:

```go
// Read plugin profile
content, err := pluginbridge.Default().Manifest().Get(ctx, "profile.yaml")

// Check if a resource exists
exists, err := pluginbridge.Default().Manifest().Exists(ctx, "i18n/zh-CN/plugin.json")
```

## Design Constraints

- **Read-only raw resources.** Reading resources does not mean executing `SQL`, registering language packs, or applying configuration.
- **Configuration reading has dedicated capabilities.** Runtime config values use `Plugins().Config()` or dynamic `config.get`; do not read `config/config.yaml` directly to bypass priority.
- **Resources are bound to the current plugin.** For source plugins they come from the embedded filesystem, for dynamic plugins from the published artifact or host-bound artifact resources.
- **Paths must be canonical.** Use forward slashes; escaping the current plugin's `manifest/` root through relative paths is prohibited.

## Related Services

- [Host Config Capability](/docs/domain-capability-hostconfig)
- [i18n Capability](/docs/domain-capability-i18n)
- [Domain Capabilities Overview](/docs/domain-capabilities)
