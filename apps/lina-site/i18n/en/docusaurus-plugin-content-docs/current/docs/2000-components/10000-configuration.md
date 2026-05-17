---
slug: '/docs/configuration'
title: 'Configuration'
hide_title: true
description: 'A component-level guide to LinaPro core host service configuration — how config.yaml drives HTTP server, logging, database, JWT authentication, sessions, monitoring, health probes, scheduling, internationalization, cluster coordination, file uploads, and plugin governance, with production deployment recommendations.'
keywords:
  - configuration
  - config.yaml
  - LinaPro configuration
  - runtime configuration
  - core host service
  - HTTP server
  - logging
  - PostgreSQL
  - SQLite
  - JWT authentication
  - session management
  - scheduling
  - internationalization
  - Redis coordinator
  - cluster configuration
  - file upload
  - plugin governance
  - autoEnable
  - production configuration
  - GoFrame configuration
---

## Introduction

Configuration management is one of the runtime entry points for `lina-core`. When the host starts, it reads `config.yaml` and distributes the settings to the `HTTP` server, logging, database, authentication, sessions, scheduling, internationalization, cluster coordination, file uploads, and plugin governance subsystems.

The default configuration file lives in the main repository:

```text
apps/lina-core/manifest/config/config.yaml
```

This file serves as both the default development template and an index for understanding how the host operates. Business projects can override sensitive settings per environment at delivery time — database connections, `JWT` secrets, log output, and cluster coordination addresses.

## Configuration Groups

| Group | Purpose |
|-------|---------|
| `server` | `HTTP` listen address, route map dump, and API documentation path |
| `logger` | Log directory, filename pattern, level, stdout, structured logging, and `TraceID` output |
| `database` | Default database connection and `SQL` debug logging |
| `jwt` | Authentication token signing secret and expiry |
| `session` | Online session timeout and expired session cleanup interval |
| `monitor` | Service metrics collection interval for monitoring plugins |
| `health` | Health probe timeout when checking database connectivity |
| `shutdown` | Maximum wait time for graceful process shutdown |
| `scheduler` | Default timezone for scheduled tasks |
| `i18n` | Default language, multi-language toggle, and locale list |
| `cluster` | Single-node or cluster mode, `Redis` coordinator, and leader election lease |
| `upload` | Upload file directory and single-file size limit |
| `plugin` | Force-uninstall policy, dynamic plugin artifact directory, and auto-enable plugin list |

## Server and API Documentation

`server.address` sets the host listen address, defaulting to `:8080`. `server.dumpRouterMap` outputs the route table at startup — useful during development but not recommended for production.

`server.extensions.apiDocPath` is a `LinaPro` extension field on the `GoFrame` server configuration, defaulting to `/api.json`:

```yaml
server:
  address: ":8080"
  dumpRouterMap: false
  extensions:
    apiDocPath: "/api.json"
```

After startup, the host publishes the aggregated `OpenAPI` document at this path. See [API Reference](/docs/api-reference) for more on API documentation design.

## Logging

Logging reuses `GoFrame` log capabilities with `LinaPro`-specific extensions:

```yaml
logger:
  path: ""
  file: "{Y-m-d}.log"
  level: "all"
  stdout: true
  extensions:
    structured: false
    traceIDEnabled: false
```

| Setting | Description |
|---------|-------------|
| `path` | Log file directory; when empty, output goes to the terminal only |
| `file` | Log filename pattern |
| `level` | Log level — `all`, `debug`, `info`, `error`, or `off` |
| `stdout` | Whether to also output to standard output |
| `structured` | Enable `JSON` structured logging |
| `traceIDEnabled` | Include `TraceID` in log output |

Production environments should enable structured logging and set a stable log directory to integrate with `ELK`, `Loki`, or cloud logging services.

## Database

The default database configuration points to `PostgreSQL`:

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
```

`PostgreSQL 14+` is the recommended production database. `SQLite` can be used for single-node local demos or smoke testing:

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite` is not suitable for production and does not support cluster deployment. In cluster mode, all `lina-core` nodes must connect to the same shared `PostgreSQL` database.

## Authentication and Sessions

Authentication configuration includes the `JWT` signing secret and token expiry:

```yaml
jwt:
  secret: "lina-jwt-secret-key-change-in-production"
  expire: 24h
```

You must replace `jwt.secret` in production. Use a random strong secret and never commit real secrets to the source repository.

Online sessions are controlled by the `session` group:

```yaml
session:
  timeout: 24h
  cleanupInterval: 5m
```

`session.timeout` determines how long an inactive session lives. `session.cleanupInterval` controls how often the built-in session cleanup task runs. This cleanup task is projected into the host's persistent task system — see [Scheduled Tasks](/docs/scheduled-tasks) for details.

## Monitoring, Health, and Shutdown

`monitor.interval` provides the collection interval for service monitoring plugins. `health.timeout` controls the timeout when the `/health` probe checks database connectivity — exceeding this returns an error status. `shutdown.timeout` controls the maximum wait time for graceful process shutdown.

```yaml
monitor:
  interval: 1m

health:
  timeout: 5s

shutdown:
  timeout: 30s
```

These settings directly affect runtime observability and orchestration system decisions. In containerized deployments, the health probe timeout should not be too long, or failed nodes will be removed later.

## Scheduling

`scheduler.defaultTimezone` defines the default timezone for persistent tasks:

```yaml
scheduler:
  defaultTimezone: "UTC"
```

If your users are primarily in mainland China, change this to `Asia/Shanghai` based on your deployment strategy. Tasks that have explicitly saved their own timezone take precedence.

## Internationalization

I18N configuration determines the runtime default language, whether multi-language is enabled, and the frontend language switch list:

```yaml
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
```

The main repository ships `zh-CN` and `en-US` runtime language resources by default. When adding a new language, you need to provide host and plugin language packs and add the language to `i18n.locales`. See [I18N Internationalization](/docs/i18n) for resource organization details.

## Cluster Coordination

The default single-node mode:

```yaml
cluster:
  enabled: false
```

Single-node mode does not connect to `Redis`. When cluster mode is enabled, the current coordination backend only supports `Redis`:

```yaml
cluster:
  enabled: true
  coordination: redis
  election:
    lease: 30s
    renewInterval: 10s
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: ""
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

`Redis` handles leader election, distributed locks, cache revision, and cross-node events. `PostgreSQL` continues to handle business data, governance data, and plugin state persistence. See [Native Distributed Architecture](/docs/distributed-architecture) for topology details.

## File Uploads

Upload configuration controls where the host saves files and the single-file size limit:

```yaml
upload:
  path: "temp/upload"
  maxSize: 20
```

Plugins that need file storage should use their own namespace, for example `temp/upload/content-notice/`, to avoid mixing with host or other plugin resources.

## Plugin Configuration

Plugin configuration bridges host governance and the plugin runtime:

```yaml
plugin:
  allowForceUninstall: true
  dynamic:
    storagePath: "temp/output"
  autoEnable:
    # - id: "demo-control"
    #   withMockData: false
```

| Setting | Description |
|---------|-------------|
| `allowForceUninstall` | Whether platform administrators can perform an audited force-uninstall after a lifecycle guard vetoes |
| `dynamic.storagePath` | Storage directory for `WASM` dynamic plugin build artifacts and uploads |
| `autoEnable` | Plugin list to automatically install and enable when the host starts |

Each entry in `autoEnable` uses the `{id, withMockData}` structure. `withMockData: true` loads demo data from the plugin's `manifest/sql/mock-data` during auto-installation — not recommended for production.

## Production Checklist

Before going live, review at least these settings:

| Item | Recommendation |
|------|----------------|
| `jwt.secret` | Replace the default with a random strong secret |
| `database.default.link` | Point to production `PostgreSQL`; do not use demo connection strings |
| `database.default.debug` | Keep `false` to avoid leaking `SQL` details |
| `logger.extensions.structured` | Enable structured logging in production |
| `scheduler.defaultTimezone` | Set to your business timezone |
| `cluster.redis` | Use a dedicated, reliable, authenticated `Redis` instance in cluster mode |
| `plugin.allowForceUninstall` | Decide based on organizational governance requirements |
| `plugin.autoEnable` | Use caution with demo data in production |
