---
slug: '/docs/configuration'
title: '配置管理'
hide_title: true
description: '本文详细介绍 LinaPro 核心宿主服务的配置文件 config.yaml，提供与当前源码一致的配置模板和配置项说明，涵盖 HTTP 服务、日志、PostgreSQL 默认数据库、JWT 认证、会话管理、定时调度、仅内置中英文的国际化配置、Redis 集群协调、文件上传和插件治理等配置分组，帮助开发者理解并调整框架的运行时行为。'
keywords:
  - 配置管理
  - config.yaml
  - LinaPro配置
  - HTTP配置
  - 日志配置
  - 数据库配置
  - JWT配置
  - 会话配置
  - 集群配置
  - 国际化配置
  - 插件配置
  - 上传配置
  - GoFrame配置
  - 运行时配置
  - 环境配置
  - 生产配置
  - PostgreSQL
  - Redis协调器
  - 多租户插件治理
---

## 配置文件位置

核心宿主服务的运行时配置文件位于：

```text
apps/lina-core/manifest/config/config.yaml
```

## 完整配置模板

以下是带完整注释的配置模板，可直接用于生产部署，按需调整各项参数：

```yaml
# HTTP 服务配置
# https://goframe.org/docs/web/server-config-file-template
server:
  # 服务监听地址，格式为 ":端口号"
  address: ":8080"
  # 是否在启动时输出路由表（开发环境建议开启，生产环境建议关闭）
  dumpRouterMap: false
  # LinaPro 对 GoFrame server 组件的扩展配置
  extensions:
    # 宿主自建 OpenAPI JSON 文档访问路径
    apiDocPath: "/api.json"

# 日志配置
# https://goframe.org/docs/core/glog-config
logger:
  # 日志文件目录；留空表示仅输出到终端，不写文件
  path: ""
  # 日志文件名模式，支持时间占位符
  file: "{Y-m-d}.log"
  # 日志级别：all | debug | info | notice | warning | error | critical | off
  level: "all"
  # 是否同时输出到标准输出
  stdout: true
  # LinaPro 扩展配置
  extensions:
    # 是否开启 JSON 结构化日志（生产环境建议开启，便于日志采集系统解析）
    structured: false
    # 是否在日志中输出 TraceID；默认 false，仅显式开启时输出
    traceIDEnabled: false

# 数据库配置
# https://goframe.org/docs/core/gdb-config-file
database:
  default:
    # 默认数据库连接串
    # PostgreSQL 14+: pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable
    # SQLite: sqlite::@file(./temp/sqlite/linapro.db)
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    # 是否开启 SQL 调试日志；默认关闭，排查 SQL 执行明细时再临时开启
    debug: false

# JWT 认证配置
jwt:
  # JWT 签名密钥（生产环境务必替换为随机强密码，至少 32 位）
  secret: "lina-jwt-secret-key-change-in-production"
  # Token 有效期，支持 duration 字符串格式（如 24h、7d、30m）
  expire: 24h

# 在线会话配置
session:
  # 会话无活动超时时间，超时后用户需要重新登录
  timeout: 24h
  # 过期会话清理任务执行间隔
  cleanupInterval: 5m

# 服务监控配置（用于 monitor-server 插件采集服务器指标）
monitor:
  # 服务指标采集间隔
  interval: 1m

# 健康探针配置
health:
  # 数据库探活超时时间，超时后 /health 端点返回 503
  timeout: 5s

# 优雅关停配置
shutdown:
  # 进程关停完整流程超时时间，超时后强制终止进程
  timeout: 30s

# 定时调度配置
scheduler:
  # 默认调度时区，未配置时使用 UTC
  # 支持 IANA 时区名称，如 Asia/Shanghai、America/New_York
  defaultTimezone: "UTC"

# 国际化配置
i18n:
  # 请求未携带语言标识或不支持时使用的默认语言
  default: zh-CN
  # 是否开启多语言能力；关闭时仅使用默认语言，前端隐藏语言切换按钮
  enabled: true
  # 运行时语言列表，决定展示顺序和原生名称
  # 注释掉某个语言即可停用（停用后该语言不再出现在切换列表中）
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文

# 集群部署配置
cluster:
  # 是否开启集群多节点模式
  # 单机部署设为 false；多节点水平扩展设为 true
  enabled: false
  # 集群模式协调后端；cluster.enabled=true 时必填
  # 当前版本仅支持 redis，单机模式不会连接 Redis
  coordination: redis
  election:
    # 主节点选举锁租约时长
    # 主节点下线后，其他节点最多等待此时长后完成重新选主
    lease: 30s
    # 租约续约间隔，建议为 lease 的 1/3
    renewInterval: 10s
  redis:
    # 集群协调使用的 Redis 地址
    address: "127.0.0.1:6379"
    # Redis 逻辑数据库
    db: 0
    # Redis 密码；Redis 未启用认证时留空
    password: ""
    # Redis 连接超时时间
    connectTimeout: 3s
    # Redis 读取超时时间
    readTimeout: 2s
    # Redis 写入超时时间
    writeTimeout: 2s

# 文件上传配置
upload:
  # 上传文件保存目录（相对于可执行文件的路径）
  path: "temp/upload"
  # 单个上传文件大小上限（单位：MB）
  maxSize: 20

# 插件配置
plugin:
  # 是否允许平台管理员在 LifecycleGuard 否决后强制卸载插件
  # 强制操作必须审计；生产环境应结合组织治理策略谨慎开启
  allowForceUninstall: true
  # 动态插件相关配置
  dynamic:
    # 动态插件（.wasm 文件）存储目录
    storagePath: "temp/output"
  # 启动时自动安装并启用的插件清单
  # 适用于新环境初始化时自动激活预置插件
  # 每个条目使用 {id, withMockData} 格式
  # withMockData: true 表示同时加载该插件的 manifest/sql/mock-data 下的演示数据
  # 注意：演示数据不建议在生产环境启用
  autoEnable:
    # - id: "demo-control"
    #   withMockData: false
    # - id: "plugin-demo-source"
    # - id: "plugin-demo-dynamic"
    #   withMockData: true
```

## 配置分组说明

| 配置分组 | 说明 |
|---------|------|
| `server` | `HTTP`服务监听地址、路由表输出、接口文档路径。<br/>更多配置项请参考：[GoFrame Server](https://goframe.org/docs/web/server-config-file-template) |
| `logger` | 日志文件路径、日志级别、结构化日志开关。<br/>更多配置项请参考：[GoFrame Logger](https://goframe.org/docs/core/glog-config) |
| `database` | 数据库连接串、`SQL`调试日志。<br/>更多配置项请参考：[GoFrame Database](https://goframe.org/docs/core/gdb-config-file) |
| `jwt` | 签名密钥、`Token`有效期 |
| `session` | 会话超时时间、清理任务间隔 |
| `monitor` | 服务器指标采集间隔 |
| `health` | 健康探针数据库超时 |
| `shutdown` | 优雅关停超时时间 |
| `scheduler` | 定时任务默认时区 |
| `i18n` | 默认语言、多语言开关、语言列表 |
| `cluster` | 集群模式开关、`Redis`协调器、选主租约配置 |
| `upload` | 上传目录、文件大小限制 |
| `plugin` | 强制卸载策略、动态插件目录、自动启用插件清单 |

## 数据库说明

`PostgreSQL 14+`是`LinaPro`默认数据库，也是生产环境推荐数据库。运行`make init`、`make dev`或跨平台`linactl`同名命令前，需要先准备可连接的`PostgreSQL`实例；这些命令不会自动启动或托管数据库。

如需单节点本地演示或冒烟验证，可以将`database.default.link`改为`SQLite`连接串：

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite`模式会强制保持单节点运行，不支持集群模式，也不建议用于生产部署。

## 集群协调说明

单机模式下，`cluster.enabled: false`，宿主不会连接`Redis`，只依赖`PostgreSQL`和进程内协调即可运行。

集群模式下，`cluster.enabled: true`时必须配置`cluster.coordination: redis`和`cluster.redis`端点。当前版本仅支持`Redis`协调器，配置项采用稳定的`coordination`标量形式，为后续扩展其他协调后端预留空间。

`Redis`协调器承担选主、分布式锁、热态会话和集群感知缓存等跨节点协调能力；`PostgreSQL`仍负责业务数据、治理数据和插件状态的持久化。

## 生产环境注意事项

**必须修改的配置项：**

1. `jwt.secret`：必须替换为随机生成的强密码，至少 32 位，切勿使用默认值
2. `database.default.link`：替换为生产数据库连接信息
3. `database.default.debug`：设为`false`，避免`SQL`日志泄露

**建议调整的配置项：**

1. `logger.structured`：生产环境设为`true`，便于日志采集系统（如`ELK`、`Loki`）解析
2. `logger.path`：设置日志文件目录，避免日志丢失
3. `scheduler.defaultTimezone`：根据业务实际所在时区调整
4. `cluster.redis`：集群部署时使用独立、可靠、带认证的`Redis`实例
5. `plugin.allowForceUninstall`：根据组织治理要求确认是否允许平台管理员强制卸载被`LifecycleGuard`否决的插件
6. `plugin.autoEnable`：按需列出生产环境需要自动启用的插件
