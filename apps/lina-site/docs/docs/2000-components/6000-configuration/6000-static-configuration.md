---
slug: '/docs/static-configuration'
title: '框架静态配置'
hide_title: true
description: '主框架 config.yaml 静态配置详解，涵盖 server、logger、database、jwt、session、health、scheduler、workspace、i18n、cluster、upload 和 plugin 等分组的配置项、默认值与使用说明，帮助开发者快速理解并定制主框架运行时行为。'
keywords:
  - 静态配置
  - config.yaml
  - LinaPro配置
  - 主框架配置
  - 服务配置
  - HTTP服务
  - 日志配置
  - 数据库配置
  - PostgreSQL
  - JWT认证
  - 会话管理
  - 健康探针
  - 定时调度
  - 管理工作台
  - 国际化配置
  - 集群配置
  - 文件上传
  - 插件管理
  - GoFrame配置
---

## 基本介绍

主框架采用**双层配置体系**，其中静态配置在启动时从`config.yaml`加载，进程生命周期内保持不变。静态配置为整个应用提供基础运行参数，涵盖`HTTP`服务、日志、数据库、认证、会话、健康探针、定时调度、前端展示、国际化、集群协调、文件上传和插件治理等核心组件。

默认配置文件位于主仓库：

```text
apps/lina-core/manifest/config/config.yaml
```

仓库同时提供了一份完整双语注释的配置模板，适合作为逐字段参考：

```text
apps/lina-core/manifest/config/config.template.yaml
```

这个文件既是开发环境的默认模板，也是理解主框架运行方式的索引。业务项目可以在交付时按环境覆盖敏感配置，例如数据库连接、`JWT`密钥、日志输出和集群协调地址。

## 配置分组

| 分组 | 作用 |
|------|------|
| `server` | 配置`HTTP`监听地址、路由表输出和接口文档访问路径 |
| `logger` | 配置日志目录、文件名、级别、标准输出、结构化日志和`TraceID`输出 |
| `database` | 配置默认数据库连接和`SQL`调试日志 |
| `jwt` | 配置认证`Token`签名密钥和有效期 |
| `session` | 配置在线会话超时和过期会话清理间隔 |
| `health` | 配置健康探针访问数据库时的超时时间 |
| `shutdown` | 配置进程优雅关停的最长等待时间 |
| `scheduler` | 配置定时调度默认时区 |
| `workspace` | 配置默认管理工作台的前端路由基准路径，并通过公开前端配置返回给`lina-vben` |
| `i18n` | 配置默认语言、多语言开关和语言列表 |
| `cluster` | 配置单机或集群模式、`Redis`协调器和选主租约 |
| `upload` | 配置上传文件保存目录和单文件大小上限 |
| `plugin` | 配置强制卸载策略、动态插件产物目录和启动自动启用插件 |

## 服务配置

`server.address`决定主框架监听地址，默认值为`:9120`。`server.dumpRouterMap`用于启动时输出路由表，适合开发排查，不建议在生产环境长期打开。

`server.extensions.apiDocPath`是`LinaPro`在`GoFrame`服务配置上的扩展字段，默认值为`/api.json`：

```yaml
server:
  address: ":9120"
  dumpRouterMap: false
  extensions:
    apiDocPath: "/api.json"
```

> 关于`server`分组的更多配置项请参考：[GoFrame Web Server配置](https://goframe.org/docs/web/server-config-file-template)


## 日志配置

日志配置复用`GoFrame`日志能力，并增加`LinaPro`自己的扩展项：

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

| 配置项 | 说明 |
|--------|------|
| `path` | 日志文件目录；为空时仅输出到终端 |
| `file` | 日志文件名模式 |
| `level` | 日志级别，例如`all`、`debug`、`info`、`error`、`off` |
| `stdout` | 是否同时输出到标准输出 |
| `structured` | 是否启用`JSON`结构化日志 |
| `traceIDEnabled` | 是否在日志中输出`TraceID` |

生产环境通常建议开启结构化日志，并设置稳定的日志目录，方便接入`ELK`、`Loki`或云日志服务。

> 关于`logger`分组的更多配置项请参考：[GoFrame日志组件配置](https://goframe.org/docs/core/glog-config)

## 数据库配置

数据库配置目前仅支持`PostgreSQL`：

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
```

> 关于`database`分组的更多配置项请参考：[GoFrame数据库配置](https://goframe.org/docs/core/gdb-config-file)

## 认证与会话

认证配置包含`JWT`签名密钥和`Token`有效期：

```yaml
jwt:
  secret: "lina-jwt-secret-key-change-in-production"
  expire: 24h
```

生产环境必须替换`jwt.secret`，建议使用随机强密钥，且不要把真实密钥提交到源码仓库。`jwt.expire`可通过运行时参数`sys.jwt.expire`在不重启进程的情况下覆盖，详见[框架动态配置](/docs/dynamic-configuration)。

在线会话由`session`分组控制：

```yaml
session:
  timeout: 24h
  cleanupInterval: 5m
```

`session.timeout`决定无活动会话的过期时间，`session.cleanupInterval`决定内置会话清理任务的运行间隔。`session.timeout`可通过运行时参数`sys.session.timeout`在不重启进程的情况下覆盖。该清理任务会被注册到主框架持久化任务系统中统一调度，更多说明见[定时任务](/docs/cron-tasks)。

## 健康探针

`health.timeout`控制`/health`探针访问数据库时的超时，超过该时间会返回异常状态。
```yaml
health:
  timeout: 5s
```

该配置直接影响部署运行时的可观测性和编排系统判断。容器化部署时，健康探针超时不宜设置过长，否则故障节点会更晚被摘除。时长字段要求秒对齐（`second-aligned`）且不低于`1s`。

## 定时调度

`scheduler.defaultTimezone`定义持久化任务的默认时区：

```yaml
scheduler:
  defaultTimezone: "UTC"
```

如果业务主要面向中国大陆用户，可以按部署策略改为`Asia/Shanghai`。已有任务若显式保存了自己的时区，则以任务配置为准。

## 管理工作台

`workspace.basePath`定义默认管理工作台的前端路由基准路径，默认值为`/admin`：

```yaml
workspace:
  basePath: "/admin"
```

本地开发时，默认工作台地址是`http://localhost:5666/admin`，主框架接口地址是`http://localhost:9120`。`workspace.basePath`会通过公开前端配置返回给`lina-vben`，用于`Vue Router`基准路径、资源地址解析和登录跳转；它不是主框架控制面`API`前缀，也不表示主框架接口地址下存在默认工作台路由。

在普通部署中，建议保留`/admin`，让根路径`/`、门户页面和源码插件自管公开路由继续可用。只有在独立管理后台域名场景下，才建议把`workspace.basePath`设置为`/`。该配置不能使用主框架保留路径，以下路径及其子路径均被禁止：

| 保留路径 | 用途 |
|----------|------|
| `/api` | 主框架`REST API`根路径 |
| `/api/v1` | 主框架`REST API v1`命名空间 |
| `/x` | 插件`API`与扩展路由 |
| `/x-assets` | 插件资源分发 |
| `/plugin-assets` | 插件静态资源 |

启动时如果检测到冲突，主框架会直接`panic`。

## I18N国际化

国际化配置决定运行时默认语言、是否启用多语言能力，以及前端语言切换列表：

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

当前主仓库默认提供`zh-CN`和`en-US`两套运行时语言资源。新增语言时，需要补齐主框架和插件语言包，并把该语言加入`i18n.locales`。重复的`locale`会静默去重，保留第一次出现的配置。更多资源组织方式见[I18N国际化](/docs/i18n)。

## 集群协调

默认单机模式如下：

```yaml
cluster:
  enabled: false
```

单机模式不连接`Redis`。启用集群后，当前版本的协调后端只支持`Redis`：

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

启用集群时，`coordination`必须设为`"redis"`（当前唯一支持的协调后端），`redis.address`必须非空，否则主框架启动时会报错。

`Redis`负责选主、分布式锁、缓存修订和跨节点事件；`PostgreSQL`仍负责业务数据、治理数据和插件状态持久化。更多拓扑说明见[原生分布式架构](/docs/distributed-architecture)。

## 文件上传

上传配置控制主框架保存文件的路径和单文件大小上限：

```yaml
upload:
  path: "temp/upload"
  maxSize: 100
```

`upload.maxSize`可通过运行时参数`sys.upload.maxSize`在不重启进程的情况下覆盖。插件如需存储文件，应使用插件自己的命名空间，例如`temp/upload/content-notice/`，避免和主框架或其他插件资源混用。

## 插件管理

插件配置连接主框架治理能力和插件运行时：

```yaml
plugin:
  allowForceUninstall: true
  dynamic:
    storagePath: "temp/output"
  autoEnable:
    # - id: "demo-control"
    #   withMockData: false
```

| 配置项 | 说明 |
|--------|------|
| `allowForceUninstall` | 是否允许平台管理员在生命周期防护否决后执行带审计的强制卸载 |
| `dynamic.storagePath` | `WASM`动态插件构建产物和上传产物的存储目录 |
| `autoEnable` | 主框架启动时自动安装并启用的插件清单 |

`autoEnable`中的每个条目使用`{id, withMockData}`结构。`withMockData: true`表示自动安装时同时加载插件`manifest/sql/mock-data`下的演示数据，`mock`数据不建议生产环境启用。
