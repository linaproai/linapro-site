---
slug: '/docs/dynamic-configuration'
title: '框架动态配置'
hide_title: true
description: '主框架运行时可热更新参数详解，涵盖 sys_config 数据表驱动的系统内置参数与业务插件自定义参数的分类管理、后端运行时参数、公开前端展示参数和 TZ 环境变量，说明参数缓存机制、集群同步策略和典型使用场景，帮助运维人员在不重启进程的情况下动态调整系统行为。'
keywords:
  - 动态配置
  - 运行时参数
  - sys_config
  - 热更新
  - 缓存机制
  - 集群同步
  - 系统内置参数
  - 业务自定义参数
  - JWT有效期
  - 会话超时
  - 上传大小限制
  - IP黑名单
  - 日志保留
  - 定时任务
  - 前端展示配置
  - 应用名称
  - Logo配置
  - 登录页配置
  - 主题配置
  - 水印配置
  - 时区配置
---

## 基本介绍

除`config.yaml`静态配置外，主框架还通过数据库`sys_config`表维护一组**运行时可热更新参数**。这些参数修改后无需重启进程即可生效，适合在管理后台直接调整。

运行时参数在进程内使用内存缓存，`TTL`为`1`小时。缓存刷新采用**修订号驱动机制**，而非单纯依赖`TTL`过期：修改参数时立即推进修订号，后续读取时检测到修订号变更即自动刷新缓存，新鲜度预算为`10`秒。

单机模式下通过内存修订号检测变更；集群模式下通过`Redis`协调的修订号同步机制保持各节点一致，修改后立即发布跨节点修订推进，其他节点在下次读取时同步更新。

> 关于框架的缓存设计请参考章节：[缓存架构与一致性设计](/docs/cache-design)。

## 参数分类与管理

`sys_config`表中的运行时参数按治理属性分为两大类：**系统内置参数**和**业务自定义参数**。

### 系统内置参数

系统内置参数由主框架预定义，具有以下治理特征：

- **键名受保护**：不可重命名或删除，`CRUD`接口会拒绝相关操作
- **内置默认值**：每个参数都有框架预设的默认值
- **值校验规则**：写入时由框架自动执行格式和范围校验

系统内置参数共`19`个，按用途分为两个子类：

| 子类 | 数量 | 用途 |
|------|------|------|
| 后端运行时参数 | `7`个 | 控制进程运行时行为，如`JWT`有效期、会话超时等 |
| 公开前端展示参数 | `12`个 | 控制品牌展示、登录页面、主题布局和水印等 |

### 业务自定义参数

业务自定义参数由插件或管理员自由创建，用于存储业务配置。其特征为：

- **键名自由**：可使用任意键名（包括`sys.`前缀），如`sys.index.skinName`
- **无内置默认值**：创建时需显式指定值
- **无自动校验**：框架不对其值格式做校验
- **完全可管理**：支持自由创建、修改和删除

> 两类参数共享同一张`sys_config`表，区别仅在于治理策略。业务插件通过`hostconfigcap`能力读写运行时配置，无需维护独立的参数注册表。

## 运行时参数概览

### 后端运行时参数

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `sys.jwt.expire` | 时长字符串 | `"24h"` | 覆盖`config.yaml`中的`jwt.expire`，控制`Token`有效期 |
| `sys.session.timeout` | 时长字符串 | `"24h"` | 覆盖`config.yaml`中的`session.timeout`，控制会话超时 |
| `sys.upload.maxSize` | 正整数（`MB`） | `"100"` | 覆盖`config.yaml`中的`upload.maxSize`，控制上传大小上限 |
| `sys.login.blackIPList` | 分号分隔的`IP`或`CIDR` | 空 | 登录`IP`黑名单，支持精确地址和`CIDR`段，例如`192.168.1.100;10.0.0.0/8` |
| `sys.log.retentionDays` | 正整数（天） | `"90"` | 日志文件最大保留天数，超期日志由清理任务自动删除 |
| `sys.cron.shell.enabled` | 布尔字符串 | `"true"` | 全局`Shell`类型定时任务开关；`Windows`下强制禁用，不受此值影响 |
| `sys.cron.log.retention` | `JSON`对象 | `{"mode":"days","value":30}` | 定时任务日志默认清理策略 |

`sys.cron.log.retention`支持三种清理模式：

| 模式 | 说明 |
|------|------|
| `days` | 删除超过`N`天的日志；`value`为正整数 |
| `count` | 仅保留最近`N`条日志；`value`为正整数 |
| `none` | 禁用自动清理；`value`强制为`0` |

### 前端展示运行时参数

以下参数通过公开前端配置接口返回给管理工作台，用于品牌展示、登录页面和界面布局：

| 键 | 默认值 | 说明 |
|----|------|--------|
| `sys.app.name` | `"LinaPro.AI"` | 应用名称，最长`120`字符 |
| `sys.app.logo` | `"/logo.webp"` | 亮色模式`Logo`路径 |
| `sys.app.logoDark` | `"/logo.webp"` | 暗色模式`Logo`路径 |
| `sys.user.defaultAvatar` | `"/avatar.webp"` | 用户默认头像路径 |
| `sys.auth.pageTitle` | 框架描述文案 | 登录页标题，最长`120`字符 |
| `sys.auth.pageDesc` | 框架描述文案 | 登录页描述，最长`500`字符 |
| `sys.auth.loginSubtitle` | 登录引导文案 | 登录表单副标题，最长`120`字符 |
| `sys.auth.loginPanelLayout` | `"panel-right"` | 登录面板布局：`panel-left`、`panel-center`、`panel-right` |
| `sys.ui.theme.mode` | `"light"` | 主题模式：`light`、`dark`、`auto` |
| `sys.ui.layout` | `"sidebar-nav"` | 全局布局：`sidebar-nav`、`sidebar-mixed-nav`、`header-nav`、`header-sidebar-nav`、`header-mixed-nav`、`mixed-nav`、`full-content` |
| `sys.ui.watermark.enabled` | `"false"` | 全局水印开关 |
| `sys.ui.watermark.content` | `"LinaPro"` | 水印文本内容，最长`120`字符 |

## 参数管理方式

### 管理接口

系统内置参数和业务自定义参数统一通过`sysconfig`服务的`CRUD`接口管理：

| 操作 | 说明 |
|------|------|
| `List` | 列表查询，支持分页和筛选 |
| `GetByKey` | 按键名查询单个参数 |
| `Create` | 创建参数，系统内置参数自动标记`isBuiltin=1` |
| `Update` | 更新参数值，系统内置参数禁止重命名键名 |
| `Delete` | 删除参数，系统内置参数禁止删除 |
| `Export` / `Import` | 批量导出导入，导入时对系统内置参数执行值校验 |

### 租户覆盖

配置值支持租户级覆盖。当存在租户上下文时，框架优先使用租户级别的值，回退到平台级别。`API`响应中包含回退元数据：

- `isFallback`：当前值是否来自平台级回退
- `canEdit`：当前租户是否可编辑
- `canOverride`：是否支持租户级覆盖
- `overrideMode`：覆盖模式

### 插件访问

业务插件通过`hostconfigcap`能力访问运行时配置：

- **只读访问**：`hostconfigcap.Service`提供`Get`、`Exists`、`String`、`Bool`、`Int`、`Duration`等方法
- **管理访问**：`hostconfigcap.AdminService`提供`BatchGetRuntimeConfig`和`SetRuntimeConfigJSON`方法

插件写入配置后会自动触发修订号推进，确保集群缓存一致性。

## TZ环境变量

主框架在检测前端时区时读取`TZ`环境变量，回退链为：`TZ`环境变量 → 进程所在时区 → `Asia/Shanghai`。容器化部署时可通过设置`TZ`确保前端展示与业务所在时区一致。
