---
slug: '/docs/configuration'
title: '服务配置管理'
hide_title: true
description: '本文从组件设计角度介绍 LinaPro 主框架服务的配置管理能力，说明 config.yaml 如何统一驱动 HTTP服务、日志、数据库、JWT认证、会话、监控、健康探针、定时调度、公开前端工作台配置、国际化、集群协调、文件上传和插件治理，并给出生产环境需要重点关注的配置边界。'
keywords:
  - 配置管理
  - config.yaml
  - LinaPro配置
  - 运行时配置
  - 主框架服务
  - HTTP服务
  - 日志配置
  - PostgreSQL
  - SQLite
  - JWT认证
  - 会话管理
  - 工作台前端配置
  - 定时调度
  - 国际化配置
  - Redis协调器
  - 集群配置
  - 文件上传
  - 插件治理
  - autoEnable
  - 生产配置
  - GoFrame配置
---

## 基本介绍

配置管理是`lina-core`的运行时入口之一。主框架启动时读取`config.yaml`，再把其中的配置分发给`HTTP`服务、日志、数据库、认证、会话、调度、公开前端工作台配置、国际化、集群协调、文件上传和插件治理等组件。

默认配置文件位于主仓库：

```text
apps/lina-core/manifest/config/config.yaml
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
| `monitor` | 配置服务指标采集间隔，供监控插件使用 |
| `health` | 配置健康探针访问数据库时的超时时间 |
| `shutdown` | 配置进程优雅关停的最长等待时间 |
| `scheduler` | 配置定时调度默认时区 |
| `workspace` | 配置默认管理工作台的前端路由基准路径，并通过公开前端配置返回给`lina-vben` |
| `i18n` | 配置默认语言、多语言开关和语言列表 |
| `cluster` | 配置单机或集群模式、`Redis`协调器和选主租约 |
| `upload` | 配置上传文件保存目录和单文件大小上限 |
| `plugin` | 配置强制卸载策略、动态插件产物目录和启动自动启用插件 |

## 服务与接口文档

`server.address`决定主框架监听地址，默认值为`:9120`。`server.dumpRouterMap`用于启动时输出路由表，适合开发排查，不建议在生产环境长期打开。

`server.extensions.apiDocPath`是`LinaPro`在`GoFrame`服务配置上的扩展字段，默认值为`/api.json`：

```yaml
server:
  address: ":9120"
  dumpRouterMap: false
  extensions:
    apiDocPath: "/api.json"
```

主框架启动后会在这个路径输出聚合后的`OpenAPI`文档。更多接口文档设计见[接口文档](/docs/api-reference)。

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

## 数据库配置

默认数据库配置指向`PostgreSQL`：

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
```

`PostgreSQL 14+`是生产推荐数据库。`SQLite`可以用于单节点本地演示或冒烟验证：

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite`不适合生产环境，也不支持集群部署。启用集群时，所有`lina-core`节点必须连接同一套共享`PostgreSQL`数据库。

## 认证与会话

认证配置包含`JWT`签名密钥和`Token`有效期：

```yaml
jwt:
  secret: "lina-jwt-secret-key-change-in-production"
  expire: 24h
```

生产环境必须替换`jwt.secret`，建议使用随机强密钥，且不要把真实密钥提交到源码仓库。

在线会话由`session`分组控制：

```yaml
session:
  timeout: 24h
  cleanupInterval: 5m
```

`session.timeout`决定无活动会话的过期时间，`session.cleanupInterval`决定内置会话清理任务的运行间隔。该清理任务会投影到主框架持久化任务系统中，更多说明见[定时任务](/docs/cron-tasks)。

## 监控、健康与关停

`monitor.interval`为服务监控插件提供采集间隔。`health.timeout`控制`/health`探针访问数据库时的超时，超过该时间会返回异常状态。`shutdown.timeout`控制进程优雅关停的最长等待时间。

```yaml
monitor:
  interval: 1m

health:
  timeout: 5s

shutdown:
  timeout: 30s
```

这些配置直接影响部署运行时的可观测性和编排系统判断。容器化部署时，健康探针超时不宜设置过长，否则故障节点会更晚被摘除。

## 定时调度

`scheduler.defaultTimezone`定义持久化任务的默认时区：

```yaml
scheduler:
  defaultTimezone: "UTC"
```

如果业务主要面向中国大陆用户，可以按部署策略改为`Asia/Shanghai`。已有任务若显式保存了自己的时区，则以任务配置为准。

## 管理工作台前端配置

`workspace.basePath`定义默认管理工作台的前端路由基准路径，默认值为`/admin`：

```yaml
workspace:
  basePath: "/admin"
```

本地开发时，默认工作台地址是`http://localhost:5666/admin`，主框架接口地址是`http://localhost:9120`。`workspace.basePath`会通过公开前端配置返回给`lina-vben`，用于`Vue Router`基准路径、资源地址解析和登录跳转；它不是主框架控制面`API`前缀，也不表示主框架接口地址下存在默认工作台路由。

在普通部署中，建议保留`/admin`，让根路径`/`、门户页面和源码插件自管公开路由继续可用。只有在独立管理后台域名场景下，才建议把`workspace.basePath`设置为`/`。该配置不能使用主框架保留路径，例如`/api`、`/x`、`/x-assets`。

## 国际化

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

当前主仓库默认提供`zh-CN`和`en-US`两套运行时语言资源。新增语言时，需要补齐主框架和插件语言包，并把该语言加入`i18n.locales`。更多资源组织方式见[I18N国际化](/docs/i18n)。

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

`Redis`负责选主、分布式锁、缓存修订和跨节点事件；`PostgreSQL`仍负责业务数据、治理数据和插件状态持久化。更多拓扑说明见[原生分布式架构](/docs/distributed-architecture)。

## 文件上传

上传配置控制主框架保存文件的路径和单文件大小上限：

```yaml
upload:
  path: "temp/upload"
  maxSize: 100
```

插件如需存储文件，应使用插件自己的命名空间，例如`temp/upload/content-notice/`，避免和主框架或其他插件资源混用。

## 插件配置

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

`autoEnable`中的每个条目使用`{id, withMockData}`结构。`withMockData: true`表示自动安装时同时加载插件`manifest/sql/mock-data`下的演示数据，不建议生产环境启用。

`autoEnable`不再接受裸字符串条目；以下写法是无效的：

```yaml
plugin:
  autoEnable:
    - "linapro-demo-source"
```

`withMockData`缺省为`false`，只在启动自动安装阶段生效；已经安装的插件不会因为配置改为`true`而重复导入演示数据。重复插件`ID`会保留第一次出现的配置。

### 插件自身业务配置

主框架`config.yaml`只保存主框架运行时配置，不承载插件自身业务配置。插件应通过插件作用域配置服务读取自己的配置，优先级如下：

| 优先级 | 路径 | 说明 |
|--------|------|------|
| 1 | `plugins/<plugin-id>/config.yaml` | 生产配置根下的插件覆盖配置；配置根由`GF_GCFG_PATH`或默认配置路径决定 |
| 2 | `apps/lina-plugins/<plugin-id>/manifest/config/config.yaml` | 开发期插件默认配置 |
| 3 | 动态插件产物中的`manifest/config/config.yaml` | 动态插件随发布版本携带的默认配置 |

`manifest/config/config.example.yaml`只是模板文件，不作为运行时默认值。源码插件通过`HostServices().Config()`读取当前插件自己的配置；动态插件通过`hostServices`声明`service: config`和`methods: [get]`后读取。确实需要读取宿主公开配置时，使用`HostConfig()`或`hostConfig.get`读取白名单键，目前白名单包括`workspace.basePath`、`i18n.default`和`i18n.enabled`。

## 生产建议

上线前至少检查这些配置：

| 项目 | 建议 |
|------|------|
| `jwt.secret` | 替换默认值，使用随机强密钥 |
| `database.default.link` | 指向生产`PostgreSQL`，不要使用演示连接串 |
| `database.default.debug` | 保持`false`，避免泄露`SQL`细节 |
| `logger.extensions.structured` | 生产环境建议启用结构化日志 |
| `workspace.basePath` | 保持`/admin`或使用独立后台域名的`/`，不要占用`/api`、`/x`、`/x-assets`；本地默认工作台地址是`http://localhost:5666/admin` |
| `scheduler.defaultTimezone` | 按业务所在时区设置 |
| `cluster.redis` | 集群模式使用独立、可靠、带认证的`Redis`实例 |
| `plugin.allowForceUninstall` | 按组织治理要求决定是否允许强制卸载 |
| `plugin.autoEnable` | 生产环境谨慎启用演示数据 |
