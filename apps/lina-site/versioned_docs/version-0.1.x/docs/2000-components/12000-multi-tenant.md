---
slug: '/docs/multi-tenant'
title: '原生多租户能力'
hide_title: true
description: '本文从组件设计角度介绍 LinaPro 多租户能力，说明主框架 bizctx、TenantFilterService、tenant_id 过滤接缝、默认平台租户、官方 multi-tenant 源码插件、租户代管、插件多租户清单字段和当前 Pool 共享表模型之间如何协作。'
keywords:
  - 多租户
  - tenant_id
  - Pool模型
  - 平台租户
  - PLATFORM
  - bizctx
  - TenantFilterService
  - TenantFilterColumn
  - CurrentContext
  - PlatformBypass
  - 租户代管
  - multi-tenant插件
  - 租户管理
  - 租户成员
  - 租户插件治理
  - scope_nature
  - supports_multi_tenant
  - default_install_mode
  - LifecycleGuard
  - LinaPro
---

## 基本介绍

`LinaPro`把多租户能力拆成“主框架基础接缝”和“官方租户控制面”两层。主框架负责在请求上下文、插件服务契约和数据过滤中提供稳定基础；官方`multi-tenant`源码插件负责租户主体、成员关系、租户解析、租户代管和租户级插件治理。

未安装或未启用`multi-tenant`插件时，框架仍以单租户方式开箱运行。此时主框架默认使用`tenant_id = 0`表示平台租户，已有项目无需为了未来多租户演进而提前改变部署模型。

## 能力分层

| 层级 | 位置 | 职责 |
|------|------|------|
| **主框架基础能力** | `apps/lina-core` | 请求级`bizctx`、身份快照、`tenant_id`过滤接缝、平台绕过策略、插件多租户元数据 |
| **租户控制面** | `apps/lina-plugins/multi-tenant` | 租户生命周期、成员关系、租户解析、租户切换、租户代管、租户插件治理 |
| **租户感知插件** | `apps/lina-plugins/<plugin-id>` | 声明多租户能力，在自有表中使用`tenant_id`隔离数据 |

这种分层保证了主框架始终稳定，租户业务规则可以通过官方插件或业务插件继续扩展。

## 默认平台租户

默认上下文中，`tenant_id = 0`代表`PLATFORM`平台租户：

- 单租户项目可以直接使用平台租户运行。
- 主框架基础表和插件表都可以保留`tenant_id`字段，为后续多租户演进留出空间。
- 平台请求是否绕过租户过滤由主框架策略控制，不应由插件自行硬编码。
- 启用`multi-tenant`插件后，用户可以进入具体租户上下文，也可以由平台管理员执行租户代管。

## 请求上下文

主框架通过`bizctx`向源码插件暴露只读请求快照。核心字段包括：

| 字段 | 说明 |
|------|------|
| `UserID` | 当前认证用户编号 |
| `Username` | 当前认证用户名 |
| `TenantID` | 当前请求所属租户编号 |
| `ActingUserID` | 租户代管时的真实平台操作者 |
| `ActingAsTenant` | 当前请求是否以租户视角执行 |
| `IsImpersonation` | 当前令牌是否代表代管请求 |
| `PlatformBypass` | 当前请求是否处于平台绕过策略中 |

插件不需要访问主框架内部上下文对象，只通过`pluginservice`发布的契约读取必要信息。

## 租户过滤服务

主框架向源码插件发布`TenantFilterService`，用于对插件自有表追加租户过滤条件。默认过滤列名由`TenantFilterColumn`定义：

```go
const TenantFilterColumn = "tenant_id"
```

服务契约提供两个核心方法：

| 方法 | 说明 |
|------|------|
| `Context(ctx)` | 返回当前请求的租户、操作者、代管状态和平台绕过元数据，结果类型为`TenantFilterContext` |
| `Apply(ctx, model, qualifier)` | 向模型追加`tenant_id`过滤条件；平台绕过时直接返回原模型；`qualifier`为联合查询中的表名或别名，单表查询传空字符串 |

单表列表查询示例：

```go
func ListArticles(ctx context.Context, filter contract.TenantFilterService) ([]*entity.Article, error) {
    model := filter.Apply(ctx, dao.ContentArticle.Ctx(ctx), "")

    var rows []*entity.Article
    err := model.Scan(&rows)
    return rows, err
}
```

联合查询时通过`qualifier`指定表名或别名，避免列名歧义：

```go
// 使用表名限定租户过滤列，生成 content_article.tenant_id = ?
model = filter.Apply(ctx, dao.ContentArticle.Ctx(ctx), "content_article")
```

写入操作需要手动读取租户编号时，可通过`Context`方法获取：

```go
tenantID := filter.Context(ctx).TenantID
```

租户感知插件应复用这套服务，而不是在各处手写`Where("tenant_id", ...)`，否则平台绕过和租户代管场景容易出现行为不一致。

## Pool共享表模型

当前版本采用`Pool`共享表模型：不同租户的数据存放在同一套数据库和同一组表中，通过`tenant_id`列区分。

| 模型 | 当前状态 | 说明 |
|------|----------|------|
| `Pool`共享表 | 已支持 | 默认模型，适合内部多团队后台和早期`SaaS`场景 |
| `Schema per tenant` | 未内置 | 后续可在现有治理模型之上扩展 |
| `Database per tenant` | 未内置 | 后续可按更强隔离需求扩展 |

插件自有表如果需要支持多租户，应在建表时包含`tenant_id`列，并为租户过滤建立合适索引。

## 官方multi-tenant插件

官方`multi-tenant`插件是源码插件，清单声明为平台级治理插件：

```yaml
id: multi-tenant
type: source
scope_nature: platform_only
supports_multi_tenant: false
default_install_mode: global
```

这些字段表示它本身不是租户级业务插件，而是租户控制面：

| 字段 | 值 | 含义 |
|------|----|------|
| `scope_nature` | `platform_only` | 仅在平台上下文中治理租户 |
| `supports_multi_tenant` | `false` | 插件自身不按租户安装 |
| `default_install_mode` | `global` | 全局唯一安装和启用 |

启用后，平台管理员可以管理租户、成员关系、租户代管入口以及租户感知插件的启用状态。

## 插件多租户声明

业务插件通过`plugin.yaml`声明自己的多租户边界：

| 字段 | 可选值 | 说明 |
|------|--------|------|
| `scope_nature` | `platform_only` / `tenant_aware` | 插件只属于平台上下文，还是可以进入租户上下文 |
| `supports_multi_tenant` | `true` / `false` | 是否支持租户级安装、开通和数据隔离 |
| `default_install_mode` | `global` / `tenant_scoped` | 默认全局启用，还是按租户独立启停 |

示例：

```yaml
id: content-article
type: source
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: tenant_scoped
```

清单字段会被主框架启动一致性检查和插件治理流程使用。字段组合不合法时，主框架会在启动或插件扫描阶段暴露明确错误。

## 插件启用模式

| 模式 | 行为 | 适用场景 |
|------|------|----------|
| `global` | 插件安装启用一次，对平台或所有租户统一生效 | 平台公共能力、全局监控、统一通知 |
| `tenant_scoped` | 插件可按租户独立启用或停用 | 内容、审计、业务模块、租户可选能力 |

新租户是否自动开通某个插件，不由`plugin.yaml`直接决定，而是由主框架治理记录和`multi-tenant`插件的开通策略维护。

## 租户代管

租户代管用于平台管理员进入某个租户视角排查问题或协助操作。代管请求需要同时保留两类身份：

| 身份 | 用途 |
|------|------|
| 当前租户身份 | 决定数据访问范围和租户视角 |
| 真实操作者身份 | 写入审计记录，避免把平台管理员操作误记为租户用户 |

因此，业务插件记录审计信息时应优先使用`CurrentContext`中的`ActingUserID`、`ActingAsTenant`和`IsImpersonation`等字段，而不是只读取当前用户编号。

## 生命周期防护

插件可以通过生命周期防护阻止危险操作。例如，`multi-tenant`插件可以在仍存在租户数据或关键治理关系时阻止禁用或卸载。

`config.yaml`中的`plugin.allowForceUninstall`控制平台管理员是否可以在生命周期防护否决后执行带审计的强制卸载：

```yaml
plugin:
  allowForceUninstall: true
```

生产环境应谨慎开启强制卸载能力，并把卸载前的数据保留、数据清理和反向依赖检查纳入操作流程。

## 当前边界

当前多租户能力聚焦于内部`BU`、多团队后台、早期`SaaS`和租户级插件治理场景。以下能力尚未作为默认能力提供：

- `Schema per tenant`或`Database per tenant`。
- 租户配额、计费和套餐管理。
- 租户独立品牌定制。
- 通过`rootDomain`自动生成租户域名。

这些能力可以在现有`Pool`模型、插件治理和租户上下文之上按业务需求扩展。

