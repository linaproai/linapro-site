---
slug: '/docs/components'
title: 'Component Design'
hide_title: true
description: 'An overview of LinaPro framework components — the core host service, default management workspace, dual-mode plugin system, source plugins, WASM dynamic plugins, plugin management, native distributed architecture, and host capabilities including configuration, API documentation, multi-tenancy, scheduled tasks, and I18N internationalization.'
keywords:
  - component design
  - LinaPro components
  - core host service
  - default management workspace
  - dual-mode plugin system
  - source plugins
  - WASM dynamic plugins
  - plugin management
  - distributed architecture
  - configuration
  - API documentation
  - multi-tenancy
  - scheduled tasks
  - I18N internationalization
  - lina-core
  - lina-vben
  - pluginhost
  - pluginbridge
  - hostServices
---


## Component Collaboration

`LinaPro` components collaborate around three main threads:

| Thread | Description |
|--------|-------------|
| **API contracts** | The host and plugins declare APIs; the workspace consumes them and supports debugging through the API documentation |
| **Plugin governance** | Plugin manifests drive installation, enablement, upgrades, disablement, uninstallation, menu projection, permission resources, and lifecycle callbacks |
| **Runtime coordination** | The host handles authentication, authorization, tenant context, configuration, scheduling, cache revision, and cluster coordination |

Through these components, `LinaPro` brings the backend host, frontend workspace, business plugins, and deployment runtime under a unified governance model. Developers can extend business capabilities with plugins while keeping the host stable, and get a complete management experience through the default workspace.

## Related Topics

import DocCardList from '@theme/DocCardList';

<DocCardList />
