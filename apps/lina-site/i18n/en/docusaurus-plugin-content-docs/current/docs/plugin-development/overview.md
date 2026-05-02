---
slug: '/docs/plugin-development'
title: '🔧 Plugin Development'
sidebar_position: 0
description: 'This plugin development manual explains the LinaPro plugin ownership model, source plugin and dynamic WASM plugin choices, plugin directory structure, manifest responsibilities, backend and frontend boundaries, SQL resource ownership, lifecycle behavior, and explicit source plugin upgrade workflow.'
keywords:
  - LinaPro
  - plugin development
  - plugin system
  - source plugin
  - dynamic plugin
  - WASM plugin
  - plugin.yaml
  - lifecycle hooks
  - plugin manifest
  - frontend pages
  - backend service
  - SQL resources
  - source upgrade
  - lina-plugins
  - extension points
  - sandbox
---

Plugins are self-contained extension units. They should own their routes, services, pages, database resources, menus, and lifecycle behavior.

## Plugin Types

| Type | Characteristics |
| --- | --- |
| Source plugin | Lives under `apps/lina-plugins/<plugin-id>/`, is explicitly wired into the product source, and uses the source plugin upgrade workflow. |
| Dynamic `WASM` plugin | Uses runtime upload, install, enable, disable, and uninstall lifecycle management. |

## Standard Source Plugin Shape

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/pages/
  manifest/sql/
  manifest/sql/mock-data/
  manifest/sql/uninstall/
  plugin.yaml
  plugin_embed.go
  README.md
  README.zh_CN.md
```

The `backend/internal/service/` directory is the legal home for source plugin business services. Do not create a parallel `backend/service/` tree.

## Upgrade Workflow

When an installed source plugin has a newer `plugin.yaml` version, the host will not silently switch to it during startup. Run an explicit upgrade:

```bash
make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>
```

Use `plugin=all` only when you intentionally want to upgrade every installed source plugin with a prepared higher version.

## Extension Discipline

Keep plugin code dependent on published host packages and extension points. If a plugin requires a new host capability, define the host contract first and keep the plugin implementation behind that contract.
