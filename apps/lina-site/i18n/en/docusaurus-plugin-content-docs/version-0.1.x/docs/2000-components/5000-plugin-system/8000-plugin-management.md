---
slug: '/docs/plugin-management'
title: 'Plugin Management'
hide_title: true
description: 'From a component design and operational flow perspective, this page explains LinaPro plugin management — plugin source configuration, the local plugin workspace, plugins.init, plugins.install, plugins.update, plugins.status, admin workspace installation and enablement, disablement and uninstallation, dynamic plugin upload, and runtime upgrade — covering the full chain from code acquisition to runtime activation.'
keywords:
  - plugin management
  - plugin workspace
  - plugins.init
  - plugins.install
  - plugins.update
  - plugins.status
  - hack/config.yaml
  - plugin source
  - plugin installation
  - plugin upgrade
  - plugin status
  - dynamic plugin upload
  - runtime upgrade
  - plugin disable
  - plugin uninstall
  - apps/lina-plugins
  - LinaPro plugins
---

## Component Positioning

Plugin management connects two chains:

- **Development chain**: Reads plugin sources from `hack/config.yaml` and synchronizes source plugins to the `apps/lina-plugins/` workspace.
- **Runtime chain**: After the host scans plugin manifests, the admin workspace handles discovery, installation, enablement, disablement, uninstallation, and upgrade.

These two chains have clear responsibilities. Code synchronization only means plugin files appear in the local workspace; whether a plugin is installed, enabled, or upgraded successfully is still determined by the host's runtime governance records.

```mermaid
flowchart LR
    Config["hack/config.yaml<br/>Plugin sources"]
    Workspace["apps/lina-plugins<br/>Local plugin workspace"]
    Host["Host scans plugin.yaml"]
    Govern["Plugin governance records<br/>Discovery, installation, enablement"]
    Runtime["Runtime projection<br/>Menus, routes, hooks, cron"]

    Config -->|"plugins.install / update"| Workspace
    Workspace --> Host
    Host --> Govern
    Govern --> Runtime
```

## Plugin Source Configuration

Plugin sources are configured in the `plugins.sources` section of `hack/config.yaml` at the repository root:

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "*"
```

| Field | Description |
|-------|-------------|
| `repo` | Plugin source repository URL |
| `root` | Root path within the source repository where plugins live |
| `ref` | Branch, tag, or `commit` |
| `items` | List of plugins to install; `"*"` means all plugins |

You can configure multiple sources, e.g. an official plugin source and an internal enterprise source. Commands support filtering via `source=<name>`.

:::info Note
The current version only supports open repositories as plugin sources. Private repository and dynamic plugin source installation will be supported in the future.
:::

## Plugin Workspace

`apps/lina-plugins/` is the fixed plugin workspace. Official plugins are typically mounted as a `Git submodule`; if the project wants to manage plugin source code independently, it can first convert to an ordinary directory:

```bash
make plugins.init
```

This command disassociates the submodule while preserving existing plugin code. If the directory does not exist, it creates an empty workspace.

:::info Note
This command is optional — the directory conversion is performed automatically during installation. After execution, `apps/lina-plugins/` becomes an ordinary directory where you can directly add, modify, or delete plugin source code.
:::

## Installing and Updating Source Plugins

Install plugins:

```bash
make plugins.install
```

Update plugins:

```bash
make plugins.update
```

Before updating, the tool checks for uncommitted local changes in the plugin directory. By default, local modifications block the update to prevent overwriting developer changes. If you need to force the update:

```bash
make plugins.update force=1
```

Check status:

```bash
make plugins.status
```

The status output shows plugin `ID`, source, version, local installation state, local modification state, and remote sync status.

## Admin Workspace Lifecycle

After plugin code enters the workspace, the host scans `plugin.yaml` at startup and presents the plugin as "Discovered". The administrator then performs runtime lifecycle operations in the Extension Center:

| Operation | Runtime behavior |
|-----------|-----------------|
| **Install** | Check dependencies, run installation `SQL`, write plugin governance records |
| **Enable** | Project menus, permissions, routes, hooks, scheduled tasks, and frontend resources |
| **Disable** | Hide menus and business routes; preserve data and governance records |
| **Uninstall** | Clean up governance records; optionally preserve or clean up plugin-owned data |
| **Upgrade** | Preview, confirm, migrate, switch version, and invalidate cache for the target plugin |

The distinction between disable and uninstall is important: disable only removes the runtime entry point while data remains; uninstall enters a cleanup flow, and if data cleanup is chosen, plugin-owned data may be unrecoverable.

## Dynamic Plugin Upload

`WASM` dynamic plugins do not depend on the source workspace for delivery. The build artifact is a `.wasm` file; after the administrator uploads it in the Extension Center, the host validates the artifact and reads the embedded manifest, routes, resources, and authorization declarations.

During dynamic plugin installation, the administrator must confirm `hostServices` authorization. Only after authorization confirmation does the host allow the plugin to access the corresponding host services and resource scope through `pluginbridge`.

## Runtime Upgrade

After plugin files are updated, the host may detect that the "effective version" and the "discovered version" are inconsistent. At this point the plugin enters a runtime upgrade state:

| State | Description |
|-------|-------------|
| `normal` | Effective version matches discovered version |
| `pending_upgrade` | A higher version has been discovered, awaiting explicit administrator upgrade |
| `upgrade_running` | Upgrade in progress |
| `upgrade_failed` | Upgrade failed; the old effective version is retained with diagnostics recorded |
| `abnormal` | File version is lower than the effective version or state is anomalous, requiring manual intervention |

Before upgrading, you can view a preview including version diff, dependency check, `SQL` count, `hostServices` diff, and risk warnings. During upgrade execution, the host performs confirmation validation, acquires the runtime upgrade lock, runs lifecycle callbacks, executes upgrade `SQL`, synchronizes governance resources, switches the effective version, and refreshes the cache.

## Multi-Tenant Governance

Tenant-aware plugins can choose between global enablement or tenant-scoped enablement:

| Mode | Behavior |
|------|----------|
| `global` | Plugin is installed and enabled once, effective for the platform or all tenants |
| `tenant_scoped` | Plugin can be enabled or disabled per tenant |

The specific enablement strategy is determined by the host governance records and the `multi-tenant` plugin, not by the frontend alone.

## Best Practices

- After synchronizing plugin code during development, still install and enable the plugin in the admin workspace.
- Check for local changes before updating plugin source code — avoid accidentally using `force=1` to overwrite development work.
- After uploading a higher version of a dynamic plugin, do not assume the new version is active — a runtime upgrade must be performed.
- Before uninstalling a production plugin, confirm whether to preserve data and check for reverse dependencies.
- Pin plugin sources to stable branches, tags, or `commits` — avoid using uncontrolled floating versions in production.
