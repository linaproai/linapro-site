---
slug: '/docs/configuration'
title: 'Service Configuration Management'
hide_title: true
description: 'Overview of LinaPro''s layered configuration system, covering main framework static configuration, main framework dynamic configuration, and plugin business configuration — explaining the division of labor between config.yaml and sys_config, plugin configuration isolation mechanisms, and production environment best practices, helping developers and ops engineers fully understand the configuration management architecture.'
keywords:
  - configuration management
  - config.yaml
  - LinaPro configuration
  - runtime configuration
  - sys_config
  - layered configuration system
  - static configuration
  - dynamic configuration
  - plugin configuration
  - configuration isolation
  - production configuration
  - best practices
  - configuration priority
  - hot update
  - cluster sync
---

## Introduction

`LinaPro` uses a **layered configuration system** that divides configuration into three levels: main framework static configuration, main framework dynamic configuration, and plugin business configuration. This design ensures core framework stability while providing flexibility for runtime adjustments and plugin extensions.

### Configuration Layers

| Layer | Source | Description |
|-------|--------|-------------|
| <span style={{whiteSpace: 'nowrap'}}><strong>Main Framework Static Config</strong></span> | `config.yaml` | Loaded at startup, unchanged during process lifetime; covers service, logging, database, authentication, and other core components |
| <span style={{whiteSpace: 'nowrap'}}><strong>Main Framework Dynamic Config</strong></span> | <span style={{whiteSpace: 'nowrap'}}>`sys_config` data table</span> | Can be hot-updated at runtime, overriding static defaults; cached in-process with `1` hour `TTL`; cluster mode maintains consistency via `Redis` revision sync |
| <span style={{whiteSpace: 'nowrap'}}><strong>Plugin Business Config</strong></span> | Plugin independent config files | Plugins have independent configuration scopes, reading config through a priority mechanism, isolated from main framework config |

### Configuration File Locations

The main framework default configuration file is located at:

```text
apps/lina-core/manifest/config/config.yaml
```

The repository also provides a fully bilingual-annotated configuration template, suitable as a per-field reference:

```text
apps/lina-core/manifest/config/config.template.yaml
```

Plugin configuration files are located in their respective plugin directories. For priority and reading details, see [Plugin Business Configuration](/docs/plugin-configuration).

## Related Documents

import DocCardList from '@theme/DocCardList';

<DocCardList />
