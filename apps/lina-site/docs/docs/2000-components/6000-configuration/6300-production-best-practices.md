---
slug: '/docs/production-best-practices'
title: '最佳实践与生产建议'
hide_title: true
description: '生产环境部署配置检查清单、安全建议和最佳实践，涵盖静态配置、运行时参数的关键配置项、配置边界和安全要求，帮助运维团队确保系统在生产环境中的稳定性、安全性和可观测性。'
keywords:
  - 生产配置
  - 最佳实践
  - 安全建议
  - 配置检查清单
  - JWT安全
  - 数据库配置
  - 日志配置
  - 集群配置
  - 插件配置
  - 运行时参数
  - IP黑名单
  - 日志保留
  - 水印配置
  - 部署建议
  - 配置边界
---

## 基本介绍

上线前至少检查这些配置，确保系统在生产环境中的稳定性、安全性和可观测性。本节提供配置检查清单和最佳实践建议。

## 静态配置检查清单

| 项目 | 建议 |
|------|------|
| `jwt.secret` | 替换默认值，使用随机强密钥 |
| `database.default.link` | 指向生产`PostgreSQL`，不要使用演示连接串 |
| `database.default.debug` | 保持`false`，避免泄露`SQL`细节 |
| `logger.extensions.structured` | 生产环境建议启用结构化日志 |
| `workspace.basePath` | 保持`/admin`或使用独立后台域名的`/`，不要占用保留路径；本地默认工作台地址是`http://localhost:5666/admin` |
| `scheduler.defaultTimezone` | 按业务所在时区设置 |
| `redis.<group>` | 集群模式使用独立、可靠、带认证的`Redis`分组 |
| `plugin.allowForceUninstall` | 按组织治理要求决定是否允许强制卸载 |
| `plugin.autoEnable` | 生产环境谨慎启用演示数据 |

## 运行时参数检查清单

| 项目 | 建议 |
|------|------|
| `sys.jwt.expire` | 按安全策略设置合理的`Token`有效期，可通过管理后台随时调整 |
| `sys.login.blackIPList` | 根据安全需要配置`IP`黑名单，支持精确地址和`CIDR`段 |
| `sys.log.retentionDays` | 根据磁盘容量和合规要求设置日志保留天数 |
| `cron.log.retention` | 按任务执行频率设置合理的日志清理策略，避免日志表无限增长 |
| `sys.ui.watermark.enabled` | 有数据防泄露需求时开启水印 |

## 安全建议

### 敏感配置管理

- **JWT密钥**：生产环境必须替换默认的`jwt.secret`，使用随机生成的强密钥，且不要将真实密钥提交到源码仓库
- **数据库连接**：确保数据库连接串中的密码安全，使用环境变量或密钥管理服务
- **Redis密码**：集群模式下的`Redis`实例应启用认证，使用强密码

### 访问控制

- **IP黑名单**：合理配置`sys.login.blackIPList`，限制可疑`IP`的访问
- **强制卸载**：根据组织治理要求决定是否启用`plugin.allowForceUninstall`
- **演示数据**：生产环境谨慎启用`plugin.autoEnable`中的`withMockData`

### 日志与审计

- **结构化日志**：生产环境建议启用`logger.extensions.structured`，便于日志收集和分析
- **日志保留**：根据磁盘容量和合规要求设置合理的`sys.log.retentionDays`
- **任务日志**：配置`cron.log.retention`避免定时任务日志无限增长

## 配置边界

### 管理工作台路径

`workspace.basePath`不能使用主框架保留路径，以下路径及其子路径均被禁止：

| 保留路径 | 用途 |
|----------|------|
| `/api` | 主框架`REST API`根路径 |
| `/api/v1` | 主框架`REST API v1`命名空间 |
| `/x` | 插件`API`与扩展路由 |
| `/x-assets` | 插件资源分发 |
| `/plugin-assets` | 插件静态资源 |

启动时如果检测到冲突，主框架会直接`panic`。

### 集群模式要求

启用集群时，`coordination`必须设为`"redis"`（当前唯一支持的协调后端），`redis.address`必须非空，否则主框架启动时会报错。

### 健康探针超时

`health.timeout`直接影响部署运行时的可观测性和编排系统判断。容器化部署时，健康探针超时不宜设置过长，否则故障节点会更晚被摘除。时长字段要求秒对齐（`second-aligned`）且不低于`1s`。

## 最佳实践总结

1. **环境分离**：开发、测试、生产环境使用不同的配置文件
2. **敏感信息**：使用环境变量或密钥管理服务管理敏感配置
3. **配置验证**：上线前验证所有必填配置项
4. **监控告警**：配置合理的健康探针和日志监控
5. **备份策略**：定期备份数据库和配置文件
6. **文档维护**：保持配置文档与实际配置同步
