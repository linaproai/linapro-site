---
slug: '/docs/configuration'
title: 'Configuration'
hide_title: true
description: 'A complete reference for the LinaPro core host service config.yaml — a fully annotated configuration template covering HTTP server, logging, database, JWT authentication, session management, scheduled tasks, internationalization, cluster deployment, file uploads, and plugin management.'
keywords:
  - configuration
  - config.yaml
  - LinaPro configuration
  - HTTP configuration
  - logging configuration
  - database configuration
  - JWT configuration
  - session configuration
  - cluster configuration
  - i18n configuration
  - plugin configuration
  - upload configuration
  - GoFrame configuration
  - runtime configuration
  - environment configuration
  - production configuration
---

## Configuration File Location

The core host service's runtime configuration file is located at:

```text
apps/lina-core/manifest/config/config.yaml
```

## Full Configuration Template

The following is a fully annotated configuration template ready for production use. Adjust each parameter as needed:

```yaml
# HTTP server configuration
# https://goframe.org/docs/web/server-config-file-template
server:
  # Listen address, format: ":port"
  address: ":8080"
  # Whether to print the route table on startup (recommended in development, disable in production)
  dumpRouterMap: false
  # LinaPro extensions to the GoFrame server component
  extensions:
    # Path for the host-generated OpenAPI JSON document
    apiDocPath: "/api.json"

# Logging configuration
# https://goframe.org/docs/core/glog-config
logger:
  # Log file directory; leave empty to output to terminal only (no file)
  path: ""
  # Log file name pattern, supports time placeholders
  file: "{Y-m-d}.log"
  # Log level: all | debug | info | notice | warning | error | critical | off
  level: "all"
  # Whether to also write to stdout
  stdout: true
  # LinaPro extensions
  extensions:
    # Enable JSON structured logging (recommended in production for log collection systems)
    structured: false
    # Include TraceID in logs; disabled by default, only output when explicitly enabled
    traceIDEnabled: false

# Database configuration
# https://goframe.org/docs/core/gdb-config-file
database:
  default:
    # Default database connection string
    # PostgreSQL 14+: pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable
    # SQLite: sqlite::@file(./temp/sqlite/linapro.db)
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    # Whether to enable SQL debug logs; keep disabled by default and enable temporarily for SQL diagnostics
    debug: false

# JWT authentication configuration
jwt:
  # JWT signing secret (replace with a random strong secret in production, at least 32 characters)
  secret: "lina-jwt-secret-key-change-in-production"
  # Token validity period, supports duration string format (e.g., 24h, 7d, 30m)
  expire: 24h

# Online session configuration
session:
  # Session inactivity timeout; user must re-login after this period
  timeout: 24h
  # Interval for the expired session cleanup task
  cleanupInterval: 5m

# Service monitoring configuration (used by the monitor-server plugin to collect metrics)
monitor:
  # Server metric collection interval
  interval: 1m

# Health probe configuration
health:
  # Database liveness check timeout; /health returns 503 after this
  timeout: 5s

# Graceful shutdown configuration
shutdown:
  # Maximum time for the full shutdown sequence; process is force-killed after this
  timeout: 30s

# Scheduled task configuration
scheduler:
  # Default scheduling timezone; UTC is used when not set
  # Supports IANA timezone names, e.g., Asia/Shanghai, America/New_York
  defaultTimezone: "UTC"

# Internationalization configuration
i18n:
  # Default language used when the request does not carry a language identifier
  default: zh-CN
  # Whether to enable multi-language support; when disabled, only the default language is used
  # and the frontend hides the language switcher
  enabled: true
  # Runtime language list — determines the display order and native name in the switcher
  # Comment out a locale to disable it (it will no longer appear in the switcher)
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
    - locale: zh-TW
      nativeName: 繁體中文

# Cluster deployment configuration
cluster:
  # Whether to enable multi-node cluster mode
  # Set to false for single-node; set to true for horizontal scaling
  enabled: false
  election:
    # Primary node election lock lease duration
    # After the primary goes down, other nodes wait at most this long before re-electing
    lease: 30s
    # Lease renewal interval — recommended to be 1/3 of lease
    renewInterval: 10s

# File upload configuration
upload:
  # Upload file storage directory (relative to the executable)
  path: "temp/upload"
  # Maximum file size per upload (in MB)
  maxSize: 20

# Plugin configuration
plugin:
  # Dynamic plugin settings
  dynamic:
    # Storage directory for dynamic plugin (.wasm) files
    storagePath: "temp/output"
  # List of plugins to automatically install and enable on startup
  # Useful for bootstrapping a fresh environment with pre-configured plugins
  # Each entry uses the {id, withMockData} format
  # withMockData: true also loads the plugin's manifest/sql/mock-data demo data
  # Note: demo data is not recommended for production environments
  autoEnable:
    # - id: "org-center"
    # - id: "content-notice"
    # - id: "monitor-online"
    # - id: "monitor-server"
    # - id: "monitor-operlog"
    # - id: "monitor-loginlog"
    # - id: "plugin-demo-dynamic"
    #   withMockData: true
```

## Configuration Group Reference

| Group | Description |
|-------|-------------|
| `server` | HTTP server listen address, route table output, API doc path.<br/>More options: [GoFrame Server](https://goframe.org/docs/web/server-config-file-template) |
| `logger` | Log file path, log level, structured logging toggle.<br/>More options: [GoFrame Logger](https://goframe.org/docs/core/glog-config) |
| `database` | Database DSN, SQL debug logging.<br/>More options: [GoFrame Database](https://goframe.org/docs/core/gdb-config-file) |
| `jwt` | Signing secret, token validity period |
| `session` | Session timeout, cleanup task interval |
| `monitor` | Server metric collection interval |
| `health` | Health probe database timeout |
| `shutdown` | Graceful shutdown timeout |
| `scheduler` | Default scheduled task timezone |
| `i18n` | Default language, multi-language toggle, locale list |
| `cluster` | Cluster mode toggle, election lease configuration |
| `upload` | Upload directory, file size limit |
| `plugin` | Dynamic plugin directory, auto-enable plugin list |

## Production Checklist

**Must change before going to production:**

1. `jwt.secret`: Replace with a randomly generated strong secret, at least 32 characters. Never use the default value.
2. `database.default.link`: Replace with production database connection details.
3. `database.default.debug`: Set to `false` to avoid leaking SQL logs.

**Recommended adjustments:**

1. `logger.structured`: Set to `true` in production for log collection systems (`ELK`, `Loki`, etc.).
2. `logger.path`: Set a log file directory to avoid losing logs.
3. `scheduler.defaultTimezone`: Adjust to the actual business timezone.
4. `plugin.autoEnable`: List the plugins that should be automatically enabled in production.
