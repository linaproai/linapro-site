---
slug: '/docs/domain-capability-route'
title: 'Route（动态路由）'
hide_title: true
description: '`RouteService`提供附着在当前请求上的动态路由元数据视图。源码插件通过`services.Route()`访问，动态插件通过`plugin.yaml`声明`service: route`后使用`pluginbridge.Default().Route()`客户端访问。它只读取`pluginbridge`运行时写入的`DynamicRouteMetadata`，不负责注册、修改或分发路由。'
keywords:
  - RouteService
  - routecap
  - DynamicRouteMetadata
  - 动态路由
  - pluginbridge
  - Wasm插件
  - 路由元数据
  - PublicPath
  - ResponseBody
  - ResponseContentType
  - APIDocService
  - 审计日志
  - 插件路由
  - capability.Services
  - LinaPro
---

## 基本介绍

`services.Route()`用于读取当前请求上的动态路由元数据。源码插件通过`services.Route()`访问，动态插件通过`plugin.yaml`声明`service: route`后使用`pluginbridge.Default().Route()`客户端访问。该服务主要服务宿主审计、操作记录或源码插件处理动态插件请求后的上下文补充。

源码插件注册自己的路由时不需要使用`RouteService`；路由注册通过`pluginhost.Declarations.HTTP().RegisterRoutes`完成。

**能力阶段**：运行期

**类型支持**：源码插件、动态插件

## 能力设计

### 元数据视图模型

`DynamicRouteMetadata`是附着在当前请求上的只读视图，包含动态路由的完整描述信息：

| 字段 | 说明 |
|------|------|
| `PluginID` | 拥有当前动态路由的插件`ID` |
| `Method` | 动态路由声明的`HTTP`方法 |
| `PublicPath` | 宿主对外暴露并匹配的路径 |
| `Tags` | 动态路由声明的标签 |
| `Summary` | 动态路由声明的摘要 |
| `Meta` | 插件自定义路由元数据 |
| `ResponseBody` | 分发器捕获的原始响应体 |
| `ResponseContentType` | 响应内容类型 |

### 路由分发流程

```mermaid
graph LR
    Request["/x/{pluginId}/..."] --> Bridge["pluginbridge分发"]
    Bridge --> Metadata["写入DynamicRouteMetadata"]
    Metadata --> Route["RouteService读取"]
    Route --> Audit["审计或日志"]
```

### 只读数据语义

调用方不能通过该服务修改动态路由或响应。非动态路由返回`nil`，使用前应判空。`ResponseBody`依赖运行时捕获结果，不应作为业务数据权威来源。

## 接口定义

### 源码插件接口

| 方法 | 说明 |
|------|------|
| `DynamicRouteMetadata` | 从`context.Context`读取动态路由元数据，非动态路由返回`nil` |

### 动态插件接口

动态插件通过`hostServices.route`声明授权的方法：

| 动态方法 | 说明 |
|----------|------|
| `metadata.get` | 读取当前请求上的动态路由元数据 |

## 能力使用

### 源码插件使用

源码插件通过`services.Route()`读取动态路由元数据，典型场景包括审计日志和操作记录：

```go
// 读取动态路由元数据
meta := services.Route().DynamicRouteMetadata(ctx)
if meta != nil {
    // 记录审计日志
    log.Infof("动态插件 %s 的路由 %s %s 被访问", meta.PluginID, meta.Method, meta.PublicPath)
}
```

源码插件注册自己的路由时使用`pluginhost.Declarations.HTTP().RegisterRoutes`：

```go
plugin := pluginhost.NewDeclarations("my-author-my-domain-my-cap")
err := plugin.HTTP().RegisterRoutes(
    pluginhost.ExtensionPointHTTPRouteRegister,
    pluginhost.CallbackExecutionModeBlocking,
    registerRoutes,
)
```

### 动态插件使用

动态插件在`plugin.yaml`中声明`route`服务：

```yaml
hostServices:
  - service: route
    methods:
      - metadata.get
```

动态插件通过`pluginbridge.Default().Route()`客户端调用：

```go
routeSvc := pluginbridge.Default().Route()
meta := routeSvc.DynamicRouteMetadata(ctx)
```

## 设计约束

- **只读数据。** 调用方不能通过该服务修改动态路由或响应。
- **非动态路由返回`nil`。** 使用前应判空。
- **响应体可能为空。** `ResponseBody`依赖运行时捕获结果，不应作为业务数据权威来源。
- **动态插件自身读取请求信封。** 动态插件内的路由信息来自`BridgeRequestEnvelopeV1.Route`，不是通过`hostServices`调用`RouteService`。

## 相关服务

- [接口文档能力](/docs/domain-capability-apidoc)
- [插件可用领域能力概览](/docs/domain-capabilities)
- [业务上下文能力](/docs/domain-capability-bizctx)
