---
slug: '/docs/components'
title: 'Component Design'
hide_title: true
description: 'An overview of the LinaPro component design philosophy: separating a stable platform foundation from pluggable business extensions, replacing implicit coupling with explicit contracts, and delivering capabilities as self-contained, composable units. Authentication, RBAC, multi-tenancy, scheduling, i18n, cluster coordination, and API documentation are built into the host; business capabilities are delivered as independent plugins; the frontend and backend are decoupled through public API contracts.'
keywords:
  - component design
  - LinaPro components
  - stable foundation
  - plugin extensibility
  - explicit contracts
  - self-contained capabilities
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
  - lina-plugins
  - pluginhost
  - pluginbridge
  - component boundaries
  - architecture
  - extensibility
  - runtime governance
  - component collaboration
---

## Design Philosophy

`LinaPro`'s component design centers on one principle: **stable foundation, extend on demand**. Platform-level capabilities are built into the host; business capabilities are delivered as self-contained components; components collaborate through explicit contracts with no hidden coupling.

This approach yields several key characteristics:

**Clear boundaries, independently replaceable.** The frontend workspace, backend host, and plugin system each have a well-defined scope of responsibility. The workspace depends only on the host's public API; plugins depend only on the stable extension interfaces the host publishes. Any layer can be upgraded or replaced without breaking the others.

**Self-contained capabilities, opt-in delivery.** Each business component (plugin) encapsulates its own API routes, database resources, frontend pages, menu permissions, language packs, and scheduled tasks — installed and uninstalled through the plugin lifecycle without touching host code. Official capabilities ship as independent plugins, so unused features never enter the deployment artifact.

**Platform capabilities out of the box.** Authentication, RBAC, multi-tenancy, scheduling, i18n, cluster coordination, and API documentation are all built into the host. Business plugins consume these directly without reimplementing them.

**Dual-mode plugins balance flexibility and performance.** Long-lived business modules use source plugins compiled into the host for native Go performance. When hot-loading or commercial binary distribution is required, WASM dynamic plugins fill that role — both modes share a single governance surface.

## Related Topics

import DocCardList from '@theme/DocCardList';

<DocCardList />
