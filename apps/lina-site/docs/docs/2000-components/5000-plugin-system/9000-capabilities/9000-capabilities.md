---
slug: '/docs/plugin-capability-services'
title: '插件可用基础能力概览'
hide_title: true
description: '本文从架构设计角度介绍 LinaPro 主框架向插件暴露的基础能力服务（capability.Services），说明 Services 接口的设计原则、服务分类、获取方式和按场景选服务指南，帮助插件开发者理解各能力服务的定位、边界和协作关系，建立正确的使用心智模型。'
keywords:
  - 插件能力
  - capability.Services
  - pluginhost.Services
  - 基础能力服务
  - 插件开发
  - 服务架构
  - 认证服务
  - 缓存服务
  - 配置服务
  - 国际化服务
  - 租户能力
  - 组织能力
  - 插件生命周期
  - 通知服务
  - 会话服务
  - LinaPro
---

## 基本介绍

`LinaPro`主框架通过`capability.Services`接口向插件暴露一组稳定的基础能力服务。这些服务覆盖了插件开发中最常见的横切关注点：认证与上下文、配置与资源、数据与存储、插件治理、通知、以及组织和租户等框架级能力。

源码插件通过`pluginhost.Services`获取完整的服务目录，它在`capability.Services`基础上扩展了`TenantFilter()`，为携带数据库查询构建器的能力提供了单独的入口。

这套服务架构遵循几个核心设计原则：

- **显式契约，稳定边界。** 每个服务都有明确的合约定义（`contract`包），插件只依赖稳定契约，不依赖宿主内部实现。
- **插件作用域隔离。** 配置、缓存、清单资源等服务自动绑定到当前插件ID，插件间不会互相干扰。
- **能力可选，安全降级。** 组织、租户等框架级能力在提供方不可用时自动降级，插件通过`Available()`检查可用性。
- **只读消费，最小暴露。** 普通插件获取的是只读消费接口，写操作和数据库查询构建器不通过`capability.Services`暴露。

## 获取方式

源码插件在路由注册、钩子回调和定时任务注册时，通过`registrar.Services()`获取完整的服务目录：

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.Services()

    // 通过 services 访问各能力服务
    config := services.Config()
    tenantFilter := services.TenantFilter()
    i18n := services.I18n()
    // ...
    return nil
}
```

`registrar.Services()`返回的`pluginhost.Services`嵌入了`capability.Services`的全部16个服务，并额外提供`TenantFilter()`。在钩子回调和定时任务注册中，同样通过`payload.Services()`或`registrar.Services()`获取。

## 服务分类速查

| 分类 | 服务 | 合约类型 | 一句话描述 |
|------|------|----------|-----------|
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `APIDoc()` | `contract.APIDocService` | API文档本地化，解析路由操作键和翻译文本 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `Auth()` | `contract.AuthService` | 租户Token签发、切换和模拟令牌管理 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `BizCtx()` | `contract.BizCtxService` | 读取当前请求的用户、租户、模拟状态等业务上下文快照 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `I18n()` | `contract.I18nService` | 运行时翻译、获取请求Locale、搜索翻译键 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `Config()` | `contract.ConfigService` | 读取当前插件自己的静态配置 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `HostConfig()` | `contract.HostConfigService` | 读取宿主公开的配置白名单键 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `Manifest()` | `contract.ManifestService` | 读取当前插件`manifest/`下的原始资源文件 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Cache()` | `contract.CacheService` | 插件作用域的运行时缓存 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Session()` | `contract.SessionService` | 在线会话管理：分页查询和踢出会话 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Route()` | `contract.RouteService` | 获取当前动态路由的元数据 |
| <span style={{whiteSpace: 'nowrap'}}>插件治理</span> | `PluginLifecycle()` | `contract.PluginLifecycleService` | 插件生命周期编排：租户级禁用/删除的前置检查和通知 |
| <span style={{whiteSpace: 'nowrap'}}>插件治理</span> | `PluginState()` | `contract.PluginStateService` | 查询插件启用状态 |
| <span style={{whiteSpace: 'nowrap'}}>通知</span> | `Notify()` | `contract.NotifyService` | 发布通知到宿主收件箱 |
| <span style={{whiteSpace: 'nowrap'}}>能力提供方</span> | `Org()` | `orgcap.Service` | 组织能力消费：用户部门、岗位等只读投影 |
| <span style={{whiteSpace: 'nowrap'}}>能力提供方</span> | `Tenant()` | `tenantcap.Service` | 租户能力消费：当前租户、可见性校验、租户切换 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `TenantFilter()` | `contract.TenantFilterService` | 为插件自有表注入`tenant_id`过滤条件 |


## 相关内容

import DocCardList from '@theme/DocCardList';

<DocCardList />
