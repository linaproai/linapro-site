---
slug: '/docs/production-best-practices'
title: 'Best Practices and Production Recommendations'
hide_title: true
description: 'Production deployment configuration checklist, security recommendations, and best practices covering static configuration, runtime parameter key items, configuration boundaries, and security requirements — helping ops teams ensure system stability, security, and observability in production environments.'
keywords:
  - production configuration
  - best practices
  - security recommendations
  - configuration checklist
  - JWT security
  - database configuration
  - logging configuration
  - cluster configuration
  - plugin configuration
  - runtime parameters
  - IP blacklist
  - log retention
  - watermark configuration
  - deployment recommendations
  - configuration boundaries
---

## Introduction

Before going live, at least check these configurations to ensure system stability, security, and observability in the production environment. This section provides a configuration checklist and best practice recommendations.

## Static Configuration Checklist

| Item | Recommendation |
|------|----------------|
| `jwt.secret` | Replace default value with a randomly generated strong key |
| `database.default.link` | Point to production `PostgreSQL`; do not use demo connection strings |
| `database.default.debug` | Keep `false` to avoid leaking `SQL` details |
| `logger.extensions.structured` | Enable structured logging in production |
| `workspace.basePath` | Keep `/admin` or use `/` for independent admin domain; do not occupy reserved paths; default workspace address locally is `http://localhost:5666/admin` |
| `scheduler.defaultTimezone` | Set according to business timezone |
| `redis.<group>` | Use an independent, reliable, authenticated Redis group in cluster mode |
| `plugin.allowForceUninstall` | Decide based on organizational governance requirements |
| `plugin.autoEnable` | Carefully enable demo data in production |

## Runtime Parameter Checklist

| Item | Recommendation |
|------|----------------|
| `sys.jwt.expire` | Set reasonable token expiration per security policy; adjustable anytime via admin console |
| `sys.login.blackIPList` | Configure `IP` blacklist based on security needs; supports exact addresses and `CIDR` ranges |
| `sys.log.retentionDays` | Set log retention days based on disk capacity and compliance requirements |
| `cron.log.retention` | Set reasonable log cleanup policy based on task execution frequency; avoid unbounded log table growth |
| `sys.ui.watermark.enabled` | Enable watermark when data leakage prevention is needed |

## Security Recommendations

### Sensitive Configuration Management

- **JWT Key**: Production must replace the default `jwt.secret` with a randomly generated strong key; do not commit real keys to the source repository
- **Database Connection**: Ensure database connection string passwords are secure; use environment variables or key management services
- **Redis Password**: `Redis` instances in cluster mode should enable authentication with strong passwords

### Access Control

- **IP Blacklist**: Reasonably configure `sys.login.blackIPList` to restrict suspicious IP access
- **Force Uninstall**: Decide whether to enable `plugin.allowForceUninstall` based on organizational governance requirements
- **Demo Data**: Carefully enable `withMockData` in `plugin.autoEnable` for production

### Logging and Auditing

- **Structured Logging**: Enable `logger.extensions.structured` in production for log collection and analysis
- **Log Retention**: Set reasonable `sys.log.retentionDays` based on disk capacity and compliance requirements
- **Task Logs**: Configure `cron.log.retention` to avoid unbounded scheduled task log growth

## Configuration Boundaries

### Admin Workspace Path

`workspace.basePath` cannot use main framework reserved paths. The following paths and their sub-paths are prohibited:

| Reserved Path | Purpose |
|---------------|---------|
| `/api` | Main framework `REST API` root path |
| `/api/v1` | Main framework `REST API v1` namespace |
| `/x` | Plugin `API` and extension routes |
| `/x-assets` | Plugin asset distribution |
| `/plugin-assets` | Plugin static assets |

If a conflict is detected at startup, the main framework will `panic` directly.

### Cluster Mode Requirements

When cluster is enabled, `coordination` must be set to `"redis"` (the only currently supported coordination backend), and `redis.address` must be non-empty, otherwise the main framework will error at startup.

### Health Probe Timeout

`health.timeout` directly affects deployment observability and orchestration system decisions. In containerized deployments, health probe timeouts should not be set too long, otherwise faulty nodes will be removed later. Duration fields require second-alignment and must be at least `1s`.

## Best Practices Summary

1. **Environment Separation**: Use different configuration files for development, testing, and production environments
2. **Sensitive Information**: Use environment variables or key management services for sensitive configuration
3. **Configuration Validation**: Validate all required configuration items before going live
4. **Monitoring Alerts**: Configure reasonable health probes and log monitoring
5. **Backup Strategy**: Regularly back up databases and configuration files
6. **Documentation Maintenance**: Keep configuration documentation in sync with actual configuration
