---
slug: '/docs/static-configuration'
title: 'Framework Static Configuration'
hide_title: true
description: 'Detailed explanation of the main framework config.yaml static configuration, covering configuration items, default values, and usage instructions for server, logger, database, jwt, session, health, scheduler, workspace, i18n, cluster, upload, and plugin groups, helping developers quickly understand and customize main framework runtime behavior.'
keywords:
  - static configuration
  - config.yaml
  - LinaPro configuration
  - main framework configuration
  - service configuration
  - HTTP service
  - logging configuration
  - database configuration
  - PostgreSQL
  - JWT authentication
  - session management
  - health probe
  - scheduled task
  - admin workspace
  - i18n configuration
  - cluster configuration
  - file upload
  - plugin management
  - GoFrame configuration
---

## Introduction

The main framework uses a **dual-layer configuration system**, where static configuration is loaded from `config.yaml` at startup and remains unchanged during the process lifetime. Static configuration provides base runtime parameters for the entire application, covering core components including `HTTP` service, logging, database, authentication, sessions, health probes, scheduled tasks, frontend display, internationalization, cluster coordination, file upload, and plugin governance.

The default configuration file is located in the main repository:

```text
apps/lina-core/manifest/config/config.yaml
```

The repository also provides a fully bilingual-annotated configuration template, suitable as a per-field reference:

```text
apps/lina-core/manifest/config/config.template.yaml
```

This file serves as both the development environment default template and an index for understanding how the main framework operates. Business projects can override sensitive configuration by environment at delivery time, such as database connections, `JWT` secrets, log output, and cluster coordination addresses.

## Configuration Groups

| Group | Purpose |
|-------|---------|
| `server` | Configure `HTTP` listen address, route table output, and API documentation access path |
| `logger` | Configure log directory, file name, level, stdout, structured logging, and `TraceID` output |
| `database` | Configure default database connection and `SQL` debug logging |
| `jwt` | Configure authentication token signing key and expiration |
| `session` | Configure online session timeout and expired session cleanup interval |
| `health` | Configure health probe database access timeout |
| `shutdown` | Configure process graceful shutdown max wait time |
| `scheduler` | Configure default timezone for scheduled tasks |
| `workspace` | Configure default admin workspace frontend route base path, returned to `lina-vben` via public frontend config |
| `i18n` | Configure default language, multi-language toggle, and language list |
| `cluster` | Configure standalone or cluster mode, `Redis` coordinator, and leader election lease |
| `upload` | Configure upload file save directory and single file size limit |
| `plugin` | Configure force uninstall policy, dynamic plugin artifact directory, and auto-enable plugins at startup |

## Service Configuration

`server.address` determines the main framework listen address, defaulting to `:9120`. `server.dumpRouterMap` outputs the route table at startup, suitable for development troubleshooting; not recommended to leave enabled in production.

`server.extensions.apiDocPath` is `LinaPro`'s extension field on the `GoFrame` service configuration, defaulting to `/api.json`:

```yaml
server:
  address: ":9120"
  dumpRouterMap: false
  extensions:
    apiDocPath: "/api.json"
```

> For more `server` group configuration items, see: [GoFrame Web Server Configuration](https://goframe.org/docs/web/server-config-file-template)

## Logging Configuration

Logging configuration reuses `GoFrame` logging capabilities with `LinaPro`'s own extensions:

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

| Config Item | Description |
|-------------|-------------|
| `path` | Log file directory; when empty, outputs only to terminal |
| `file` | Log file name pattern |
| `level` | Log level, e.g., `all`, `debug`, `info`, `error`, `off` |
| `stdout` | Whether to also output to stdout |
| `structured` | Whether to enable `JSON` structured logging |
| `traceIDEnabled` | Whether to output `TraceID` in logs |

Production environments typically recommend enabling structured logging and setting a stable log directory for integration with `ELK`, `Loki`, or cloud log services.

> For more `logger` group configuration items, see: [GoFrame Log Component Configuration](https://goframe.org/docs/core/glog-config)

## Database Configuration

Database configuration currently only supports `PostgreSQL`:

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
```

> For more `database` group configuration items, see: [GoFrame Database Configuration](https://goframe.org/docs/core/gdb-config-file)

## Authentication and Sessions

Authentication configuration includes the `JWT` signing key and token expiration:

```yaml
jwt:
  secret: "lina-jwt-secret-key-change-in-production"
  expire: 24h
```

Production environments must replace `jwt.secret` with a randomly generated strong key, and should not commit real keys to the source repository. `jwt.expire` can be overridden at runtime via the `sys.jwt.expire` parameter without restarting the process — see [Framework Dynamic Configuration](/docs/dynamic-configuration).

Online sessions are controlled by the `session` group:

```yaml
session:
  timeout: 24h
  cleanupInterval: 5m
```

`session.timeout` determines inactive session expiration time; `session.cleanupInterval` determines the built-in session cleanup task's run interval. `session.timeout` can be overridden at runtime via the `sys.session.timeout` parameter without restarting the process. The cleanup task is registered into the main framework's persistent task system for unified scheduling — see [Scheduled Tasks](/docs/cron-tasks).

## Health Probe

`health.timeout` controls the `/health` probe's database access timeout. Exceeding this time returns an abnormal status.

```yaml
health:
  timeout: 5s
```

This configuration directly affects deployment observability and orchestration system decisions. In containerized deployments, health probe timeouts should not be set too long, otherwise faulty nodes will be removed later. Duration fields require second-alignment and must be at least `1s`.

## Scheduled Tasks

`scheduler.defaultTimezone` defines the default timezone for persistent tasks:

```yaml
scheduler:
  defaultTimezone: "UTC"
```

If the business primarily serves users in mainland China, it can be changed to `Asia/Shanghai` per deployment strategy. Tasks that have explicitly saved their own timezone use the task configuration.

## Admin Workspace

`workspace.basePath` defines the default admin workspace's frontend route base path, defaulting to `/admin`:

```yaml
workspace:
  basePath: "/admin"
```

During local development, the default workspace address is `http://localhost:5666/admin` and the main framework API address is `http://localhost:9120`. `workspace.basePath` is returned to `lina-vben` via the public frontend configuration for `Vue Router` base path, resource URL resolution, and login redirect. It is not the main framework control plane `API` prefix, nor does it indicate default workspace routes exist under the main framework API address.

In standard deployments, it is recommended to keep `/admin`, allowing the root path `/`, portal pages, and source plugin self-managed public routes to remain usable. Only in independent admin backend domain scenarios should `workspace.basePath` be set to `/`. This configuration cannot use main framework reserved paths — the following paths and their sub-paths are prohibited:

| Reserved Path | Purpose |
|---------------|---------|
| `/api` | Main framework `REST API` root path |
| `/api/v1` | Main framework `REST API v1` namespace |
| `/x` | Plugin `API` and extension routes |
| `/x-assets` | Plugin asset distribution |
| `/plugin-assets` | Plugin static assets |

If a conflict is detected at startup, the main framework will `panic` directly.

## i18n Internationalization

Internationalization configuration determines the runtime default language, whether multi-language is enabled, and the frontend language switch list:

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

The main repository currently provides `zh-CN` and `en-US` runtime language resources by default. When adding a new language, language packs for both the main framework and plugins need to be completed, and the language should be added to `i18n.locales`. Duplicate `locale` values are silently deduplicated, keeping the first occurrence. See [i18n Internationalization](/docs/i18n) for more resource organization details.

## Cluster Coordination

Default standalone mode:

```yaml
cluster:
  enabled: false
```

Standalone mode does not connect to `Redis`. When cluster is enabled, the current version's coordination backend only supports `Redis`:

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

When cluster is enabled, `coordination` must be set to `"redis"` (the only currently supported coordination backend), and `redis.address` must be non-empty, otherwise the main framework will error at startup.

`Redis` handles leader election, distributed locks, cache revision, and cross-node events; `PostgreSQL` continues to handle business data, governance data, and plugin state persistence. See [Native Distributed Architecture](/docs/distributed-architecture) for more topology details.

## File Upload

Upload configuration controls the main framework's file save path and single file size limit:

```yaml
upload:
  path: "temp/upload"
  maxSize: 100
```

`upload.maxSize` can be overridden at runtime via the `sys.upload.maxSize` parameter without restarting the process. Plugins that need to store files should use their own namespace, e.g., `temp/upload/content-notice/`, to avoid mixing with main framework or other plugin resources.

## Plugin Management

Plugin configuration connects main framework governance capabilities with the plugin runtime:

```yaml
plugin:
  allowForceUninstall: true
  dynamic:
    storagePath: "temp/output"
  autoEnable:
    # - id: "demo-control"
    #   withMockData: false
```

| Config Item | Description |
|-------------|-------------|
| `allowForceUninstall` | Whether to allow platform admins to force-uninstall with audit after lifecycle protection vetoes |
| `dynamic.storagePath` | Storage directory for `WASM` dynamic plugin build artifacts and upload artifacts |
| `autoEnable` | List of plugins to auto-install and enable at main framework startup |

Each entry in `autoEnable` uses the `{id, withMockData}` structure. `withMockData: true` means loading demo data from the plugin's `manifest/sql/mock-data` during auto-install; `mock` data is not recommended for production.
