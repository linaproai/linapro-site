---
slug: '/docs/multi-tenant'
title: '多租户能力'
hide_title: true
description: '本文介绍 LinaPro 的多租户能力，包括宿主内置的 bizctx 请求上下文、tenant_id 租户过滤接缝、默认 tenant_id = 0 的平台租户模式、官方 multi-tenant 源码插件提供的租户生命周期和成员关系管理、插件多租户清单字段、全局与租户级插件启用模式、LifecycleGuard 防护和当前 Pool 共享数据库模型的边界，帮助开发者在单租户和多租户场景之间平滑演进。'
keywords:
  - 多租户
  - tenant_id
  - Pool模型
  - 平台租户
  - PLATFORM
  - multi-tenant插件
  - 租户管理
  - 租户成员
  - 租户解析
  - bizctx
  - TenantFilter
  - 插件多租户
  - scope_nature
  - supports_multi_tenant
  - default_install_mode
  - LifecycleGuard
  - 插件治理
  - LinaPro
---

## 概述

`LinaPro`将多租户能力拆成两层：宿主提供稳定的租户上下文和数据过滤接缝，官方`multi-tenant`源码插件提供租户管理界面、租户生命周期、成员关系和租户插件治理等可见能力。

未安装或未启用`multi-tenant`插件时，框架仍保持默认单租户开箱体验。宿主与插件数据统一使用`tenant_id = 0`，表示`PLATFORM`平台租户；已有单租户项目无需为了后续多租户演进而提前改造运行方式。

## 能力分层

| 层级 | 位置 | 职责 |
|------|------|------|
| 宿主基础能力 | `apps/lina-core` | 请求级`bizctx`、租户身份快照、`tenant_id`过滤接缝、插件多租户元数据、平台绕过策略 |
| 官方管理插件 | `apps/lina-plugins/multi-tenant` | 租户主体、租户成员、租户解析策略、租户代管、租户插件启用治理 |
| 租户感知插件 | `apps/lina-plugins/<plugin-id>` | 在清单中声明多租户能力，并在插件数据表中使用`tenant_id`隔离租户数据 |

## 默认平台租户

默认状态下，所有请求都运行在平台上下文中：

- `tenant_id = 0`代表`PLATFORM`平台租户
- 平台请求可以按宿主策略绕过租户过滤
- 插件表如果需要未来支持多租户，应预留`tenant_id`列
- 单租户项目无需启用`multi-tenant`插件即可正常运行

## Pool 共享数据库模型

当前版本采用`Pool`共享数据库模型：不同租户的数据存放在同一套数据库和表结构中，通过`tenant_id`列区分。

| 模型 | 当前状态 | 说明 |
|------|---------|------|
| `Pool`共享表 | 已支持 | 默认模型，适合多数内部系统和`SaaS`早期阶段 |
| 独立`Schema` | 未提供 | 作为后续演进方向保留 |
| 独立数据库 | 未提供 | 作为后续演进方向保留 |

宿主向源码插件发布`TenantFilterService`，插件可以基于当前请求上下文获取租户身份，并对自有表追加租户过滤条件。默认租户过滤列名为`tenant_id`。

## 官方 multi-tenant 插件

官方`multi-tenant`插件是源码插件，提供租户可见管理能力。它自身是平台级治理插件，清单中声明为：

```yaml
id: multi-tenant
type: source                   # 源码插件，随宿主一起编译
scope_nature: platform_only    # 仅在平台上下文治理，不进入租户上下文
supports_multi_tenant: false   # 自身不支持租户级安装，因为它就是租户控制面
default_install_mode: global   # 全局唯一安装，不按租户独立启停
```

启用后，平台管理员可以使用以下能力：

- 租户生命周期管理，包括创建、编辑、删除和状态治理
- 租户成员关系管理，一个用户可以加入多个租户
- 租户选择和切换，用户登录后可进入具体租户上下文
- 平台代管租户，审计中保留真实操作者和被代管租户
- 租户插件治理，支持对租户感知插件进行全局或租户级启用

## 插件多租户清单字段

插件需要在`plugin.yaml`中声明多租户边界：

| 字段 | 可选值 | 说明 |
|------|--------|------|
| `scope_nature` | `platform_only` / `tenant_aware` | 插件是平台级治理能力，还是可按租户上下文治理 |
| `supports_multi_tenant` | `true` / `false` | 是否支持租户级安装、开通和数据隔离 |
| `default_install_mode` | `global` / `tenant_scoped` | 默认是全局启用，还是按租户独立启用 |

示例：

```yaml
id: demo-control
type: source
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: global
```

## 插件启用模式

| 模式 | 适用场景 | 行为 |
|------|---------|------|
| `global` | 平台公共能力、所有租户共享能力 | 插件安装启用一次，对平台或所有租户生效 |
| `tenant_scoped` | 内容、审计、业务模块等租户可独立开通能力 | 插件可以按租户单独启用或停用 |

新租户是否自动启用某个租户感知插件，由平台插件注册表中的开通策略维护，不由`plugin.yaml`直接声明。

## 生命周期防护

插件可以通过`LifecycleGuard`在禁用或卸载前执行防护检查。例如，`multi-tenant`插件可以在仍存在租户数据时阻止卸载，避免误删导致业务不可恢复。

`config.yaml`中的`plugin.allowForceUninstall`控制平台管理员是否允许在`LifecycleGuard`否决后执行带审计的强制卸载：

```yaml
plugin:
  allowForceUninstall: true
```

生产环境应根据组织治理策略谨慎设置该选项。

## 当前边界

当前多租户能力聚焦于内部`BU`、多团队后台、早期`SaaS`和租户级插件治理场景。以下能力尚未作为默认能力提供：

- `Schema per tenant`或`Database per tenant`
- 租户配额、计费和套餐管理
- 租户独立品牌定制
- 通过`rootDomain`自动生成租户域名

如需这些能力，建议在当前`Pool`模型和插件治理能力之上按业务需求扩展。
