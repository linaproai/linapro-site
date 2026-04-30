---
slug: '/docs/config-management'
title: '配置管理'
sidebar_position: 3
hide_title: true
description: '本文详细介绍 LinaPro 核心宿主服务的配置文件 config.yaml，提供完整的配置模板和各配置项的说明，涵盖 HTTP 服务、日志、数据库、JWT 认证、会话管理、定时调度、国际化、集群部署、文件上传和插件管理等所有配置分组，帮助开发者快速理解和调整框架的运行时行为。'
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

# 数据库配置
# https://goframe.org/docs/core/gdb-config-file
database:
  default:
    # 数据库连接串格式：mysql:<user>:<pass>@tcp(<host>:<port>)/<dbname>?参数
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
    # 是否开启 SQL 调试日志（生产环境建议关闭）
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
    - locale: zh-TW
      nativeName: 繁體中文

# 集群部署配置
cluster:
  # 是否开启集群多节点模式
  # 单机部署设为 false；多节点水平扩展设为 true
  enabled: false
  election:
    # 主节点选举锁租约时长
    # 主节点下线后，其他节点最多等待此时长后完成重新选主
    lease: 30s
    # 租约续约间隔，建议为 lease 的 1/3
    renewInterval: 10s

# 文件上传配置
upload:
  # 上传文件保存目录（相对于可执行文件的路径）
  path: "temp/upload"
  # 单个上传文件大小上限（单位：MB）
  maxSize: 20

# 插件配置
plugin:
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
    # - id: "org-center"
    # - id: "content-notice"
    # - id: "monitor-online"
    # - id: "monitor-server"
    # - id: "monitor-operlog"
    # - id: "monitor-loginlog"
    # - id: "plugin-demo-dynamic"
    #   withMockData: true
```

## 配置分组说明

| 配置分组 | 说明 |
|---------|------|
| `server` | `HTTP`服务监听地址、路由表输出、接口文档路径 |
| `logger` | 日志文件路径、日志级别、结构化日志开关 |
| `database` | 数据库连接串、`SQL`调试日志 |
| `jwt` | 签名密钥、`Token`有效期 |
| `session` | 会话超时时间、清理任务间隔 |
| `monitor` | 服务器指标采集间隔 |
| `health` | 健康探针数据库超时 |
| `shutdown` | 优雅关停超时时间 |
| `scheduler` | 定时任务默认时区 |
| `i18n` | 默认语言、多语言开关、语言列表 |
| `cluster` | 集群模式开关、选主租约配置 |
| `upload` | 上传目录、文件大小限制 |
| `plugin` | 动态插件目录、自动启用插件清单 |

## 生产环境注意事项

**必须修改的配置项：**

1. `jwt.secret`：必须替换为随机生成的强密码，至少 32 位，切勿使用默认值
2. `database.default.link`：替换为生产数据库连接信息
3. `database.default.debug`：设为`false`，避免`SQL`日志泄露

**建议调整的配置项：**

1. `logger.structured`：生产环境设为`true`，便于日志采集系统（如`ELK`、`Loki`）解析
2. `logger.path`：设置日志文件目录，避免日志丢失
3. `scheduler.defaultTimezone`：根据业务实际所在时区调整
4. `plugin.autoEnable`：按需列出生产环境需要自动启用的插件

## 多环境配置

`GoFrame`框架支持通过`GF_GFCLI_BUILD_COND_FILE`环境变量或目录约定实现多环境配置，常见做法是在`manifest/config/`目录下维护多个配置文件，通过构建时或启动时选择加载：

```text
manifest/config/
  config.yaml           # 开发环境配置（含默认值）
  config.prod.yaml      # 生产环境配置（覆盖敏感配置）
```

也可以通过环境变量覆盖配置文件中的特定项：

```bash
# 通过环境变量覆盖 JWT 密钥（生产环境推荐）
GF_DATABASE_DEFAULT_LINK="mysql:..." go run main.go
```
