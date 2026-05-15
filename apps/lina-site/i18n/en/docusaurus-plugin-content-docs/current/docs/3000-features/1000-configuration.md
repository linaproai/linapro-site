---
slug: '/docs/configuration'
title: 'Configuration'
hide_title: true
description: 'A detailed guide to the LinaPro core host service configuration file config.yaml, covering HTTP server, logging, PostgreSQL default database, JWT authentication, session management, scheduled tasks, internationalization (built-in en-US and zh-CN only), Redis cluster coordination, file uploads, and plugin governance — everything you need to tune the framework runtime behavior.'
keywords:
  - configuration
  - config.yaml
  - LinaPro configuration
  - HTTP config
  - logging config
  - database config
  - JWT config
  - session config
  - cluster config
  - i18n config
  - plugin config
  - upload config
  - GoFrame config
  - runtime config
  - environment config
  - production config
  - PostgreSQL
  - Redis coordinator
  - multi-tenant plugin governance
---

## Configuration File Location

The runtime configuration file for the core host service is located at:

```text
apps/lina-core/manifest/config/config.yaml
```

## Complete Configuration Template

Below is the fully annotated configuration template, ready for production deployment — adjust each parameter as needed:

```yaml
# HTTP server configuration
# https://goframe.org/docs/web/server-config-file-template
server:
  # Listen address in ":port" format
  address: ":8080"
  # Whether to dump the route table at startup (enable in dev, disable in prod)
  dumpRouterMap: false
  # LinaPro extensions for the GoFrame server component
  extensions:
    # Path to the host's auto-generated OpenAPI JSON document
    apiDocPath: "/api.json"

# Logging configuration
# https://goframe.org/docs/core/glog-config
logger:
  # Log file directory; leave empty to output only to the terminal
  path: ""
  # Log file name pattern, supports time placeholders
  file: "{Y-m-d}.log"
  # Log level: all | debug | info | notice | warning | error | critical | off
  level: "all"
  # Whether to also write to stdout
  stdout: true
  # LinaPro extensions
  extensions:
    # Enable JSON structured logging (recommended for prod; aids log collectors)
    structured: false
    # Whether to include TraceID in log output; defaults to false
    traceIDEnabled: false

# Database configuration
# https://goframe.org/docs/core/gdb-config-file
database:
  default:
    # Default database connection string
    # PostgreSQL 14+: pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable
    # SQLite: sqlite::@file(./temp/sqlite/linapro.db)
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    # Enable SQL debug logging; off by default — enable temporarily when troubleshooting
    debug: false

# JWT authentication configuration
jwt:
  # JWT signing secret (replace with a random strong secret in production, at least 32 chars)
  secret: "lina-jwt-secret-key-change-in-production"
  # Token validity period, supports duration strings (e.g., 24h, 7d, 30m)
  expire: 24h

# Online session configuration
session:
  # Session inactivity timeout; users must re-authenticate after this period
  timeout: 24h
  # Cleanup interval for expired session records
  cleanupInterval: 5m

# Service monitor configuration (used by the monitor-server plugin to collect metrics)
monitor:
  # Metrics collection interval
  interval: 1m

# Health probe configuration
health:
  # Database liveness probe timeout; the /health endpoint returns 503 when exceeded
  timeout: 5s

# Graceful shutdown configuration
shutdown:
  # Full shutdown timeout; the process is force-terminated when exceeded
  timeout: 30s

# Scheduled task configuration
scheduler:
  # Default scheduling timezone; UTC if not set
  # Supports IANA timezone names, e.g., Asia/Shanghai, America/New_York
  defaultTimezone: "UTC"

# Internationalization configuration
i18n:
  # Default language when a request carries no language identifier or the language is unsupported
  default: zh-CN
  # Enable multi-language; when off, only the default language is used and the switcher is hidden
  enabled: true
  # Runtime language list; controls display order and native names
  # Comment out a language to disable it
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文

# Cluster deployment configuration
cluster:
  # Enable multi-node cluster mode
  # Set to false for single-node deployments; true for horizontal scaling
  enabled: false
  # Cluster coordination backend; required when cluster.enabled=true
  # Current version only supports redis; single-node mode does not connect to Redis
  coordination: redis
  election:
    # Leader election lock lease duration
    # After the primary goes down, other nodes wait at most this long before re-electing
    lease: 30s
    # Lease renewal interval; recommended as 1/3 of lease
    renewInterval: 10s
  redis:
    # Redis address used for cluster coordination
    address: "127.0.0.1:6379"
    # Redis logical database
    db: 0
    # Redis password; leave empty if Redis auth is not enabled
    password: ""
    # Redis connection timeout
    connectTimeout: 3s
    # Redis read timeout
    readTimeout: 2s
    # Redis write timeout
    writeTimeout: 2s

# File upload configuration
upload:
  # Upload file storage directory (relative to the executable)
  path: "temp/upload"
  # Maximum single file upload size (in MB)
  maxSize: 20

# Plugin configuration
plugin:
  # Allow platform administrators to force-uninstall a plugin after LifecycleGuard vetoes
  # Force operations must be audited; enable cautiously in production per governance policy
  allowForceUninstall: true
  # Dynamic plugin configuration
  dynamic:
    # Storage directory for dynamic plugin .wasm files
    storagePath: "temp/output"
  # Plugins to auto-install and enable at startup
  # Useful for bootstrapping new environments with preset plugins
  # Each entry uses the {id, withMockData} format
  # withMockData: true loads the plugin's manifest/sql/mock-data demo records
  # Note: demo data is not recommended for production
  autoEnable:
    # - id: "demo-control"
    #   withMockData: false
    # - id: "plugin-demo-source"
    # - id: "plugin-demo-dynamic"
    #   withMockData: true
```

## Configuration Groups

| Group | Description |
|-------|-------------|
| `server` | HTTP listen address, route table dump, API doc path.<br/>See also: [GoFrame Server](https://goframe.org/docs/web/server-config-file-template) |
| `logger` | Log file path, log level, structured logging toggle.<br/>See also: [GoFrame Logger](https://goframe.org/docs/core/glog-config) |
| `database` | Database connection string, SQL debug logging.<br/>See also: [GoFrame Database](https://goframe.org/docs/core/gdb-config-file) |
| `jwt` | Signing secret, token validity period |
| `session` | Session timeout, cleanup interval |
| `monitor` | Server metrics collection interval |
| `health` | Health probe database timeout |
| `shutdown` | Graceful shutdown timeout |
| `scheduler` | Default timezone for scheduled tasks |
| `i18n` | Default language, multi-language toggle, language list |
| `cluster` | Cluster mode toggle, Redis coordinator, leader election lease settings |
| `upload` | Upload directory, file size limit |
| `plugin` | Force-uninstall policy, dynamic plugin directory, auto-enable plugin list |

## Database Notes

`PostgreSQL 14+` is the default database for `LinaPro` and the recommended choice for production. Before running `make init`, `make dev`, or the cross-platform `linactl` equivalents, make sure a reachable `PostgreSQL` instance is available — these commands do not start or manage a database for you.

For single-node local demos or smoke tests, you can switch `database.default.link` to a `SQLite` connection string:

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite` mode enforces single-node operation, does not support cluster mode, and is not recommended for production deployments.

## Cluster Coordination Notes

In single-node mode (`cluster.enabled: false`), the host does not connect to `Redis` and relies solely on `PostgreSQL` and in-process coordination.

In cluster mode (`cluster.enabled: true`), you must configure `cluster.coordination: redis` and provide `cluster.redis` endpoints. The current version only supports the `Redis` coordinator; the configuration uses a stable `coordination` scalar to leave room for future coordination backends.

The `Redis` coordinator handles leader election, distributed locks, hot session state, and cluster-aware caching across nodes. `PostgreSQL` remains responsible for persisting business data, governance data, and plugin state.

## Production Checklist

**Must change:**

1. `jwt.secret` — replace with a randomly generated strong secret (at least 32 characters). Never use the default value.
2. `database.default.link` — replace with your production database connection string.
3. `database.default.debug` — set to `false` to prevent SQL log leakage.

**Recommended adjustments:**

1. `logger.structured` — set to `true` in production for easier parsing by log collectors (ELK, Loki, etc.).
2. `logger.path` — set a log file directory to prevent log loss.
3. `scheduler.defaultTimezone` — adjust to match your business timezone.
4. `cluster.redis` — use a dedicated, reliable, authenticated `Redis` instance for cluster deployments.
5. `plugin.allowForceUninstall` — confirm whether platform administrators may force-uninstall plugins vetoed by `LifecycleGuard`, per your organization's governance policy.
6. `plugin.autoEnable` — list the plugins that should auto-enable in production.
