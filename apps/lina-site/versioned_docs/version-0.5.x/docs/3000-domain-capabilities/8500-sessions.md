---
slug: '/docs/domain-capability-sessions'
title: 'Sessions（会话管理）'
hide_title: true
description: '`Sessions()`为源码插件提供在线会话搜索、批量读取和吊销能力。动态插件通过`plugin.yaml`声明`service: sessions`后使用`pluginbridge.Default().Sessions()`客户端访问已发布的只读方法。'
keywords:
  - SessionService
  - sessioncap
  - Sessions
  - 在线会话
  - SearchSessions
  - BatchGetSessions
  - Revoke
  - Token管理
  - ClientType
  - DeptName
  - 会话视图
  - 会话吊销
  - 插件能力
  - LinaPro
---

## 基本介绍

源码插件通过`services.Sessions()`访问在线会话的读取和吊销能力，动态插件通过`plugin.yaml`声明`service: sessions`后使用`pluginbridge.Default().Sessions()`客户端访问已发布的只读方法。

**能力阶段**：运行期

**类型支持**：源码插件、动态插件

## 能力设计

### 会话视图模型

会话视图用于在线用户监控、会话治理和安全审计场景，不暴露会话存储表或`JWT`内部实现：

| 字段 | 说明 |
|------|------|
| `ID` | 会话领域标识 |
| `TenantID` | 当前租户标识 |
| `UserID`、`Username` | 会话用户 |
| `ClientType` | 客户端类型，例如`web`、`mobile`、`desktop`、`cli` |
| `DeptName` | 登录时捕获或装配的部门名称 |
| `Ip`、`Browser`、`Os` | 登录环境信息 |
| `LoginAt`、`LastActiveAt` | 登录时间和最近活跃时间 |

### 读写操作

`Sessions()`统一提供读取和吊销操作。`Revoke`和`RevokeMany`经过租户、数据范围、目标可见性和审计校验后执行。会话吊销即时影响令牌，吊销后对应令牌应无法继续通过宿主认证中间件。

### 部门名称视图

`DeptName`是视图字段，组织能力不可用时可能为空。

## 接口定义

### 源码插件接口

| 方法 | 说明 |
|------|------|
| `Current` | 返回当前令牌的可见会话视图 |
| `Get` | 读取单个可见会话视图 |
| `List` | 按用户名、`IP`和分页条件搜索可见会话 |
| `BatchGet` | 批量读取可见会话视图 |
| `BatchGetUserOnlineStatus` | 批量读取用户在线状态 |
| `EnsureVisible` | 校验目标会话集合对当前调用上下文可见 |
| `Revoke` | 吊销一个可见在线会话，经过租户、数据范围、目标可见性和审计校验 |
| `RevokeMany` | 批量吊销可见在线会话，任何不可见目标会拒绝整个操作 |

### 动态插件接口

动态插件通过`hostServices.sessions`声明授权的只读方法：

| 动态方法 | 说明 |
|----------|------|
| `sessions.current` | 返回当前令牌的可见会话视图 |
| `sessions.list` | 按用户名、`IP`和分页条件搜索可见会话 |
| `sessions.batch_get` | 批量读取可见会话视图 |
| `sessions.batch_get_user_online_status` | 批量读取用户在线状态 |
| `sessions.visible.ensure` | 校验目标会话集合对当前调用上下文可见 |

## 能力使用

### 源码插件使用

源码插件通过`services.Sessions()`读取和管理会话，并显式传入领域要求的`CapabilityContext`：

```go
// 获取当前会话视图
current, err := services.Sessions().Current(ctx)

// 搜索在线会话
page, err := services.Sessions().List(ctx, sessioncap.ListInput{
    Username: keyword,
    Page:     pageRequest,
})

// 批量读取会话视图
result, err := services.Sessions().BatchGet(ctx, sessionIDs)

// 批量读取用户在线状态
onlineStatus, err := services.Sessions().BatchGetUserOnlineStatus(ctx, userIDs)

// 校验会话可见性
err := services.Sessions().EnsureVisible(ctx, sessionIDs)

// 吊销单个会话
err := services.Sessions().Revoke(ctx, sessionID)

// 批量吊销会话
err := services.Sessions().RevokeMany(ctx, sessionIDs)
```

### 动态插件使用

动态插件在`plugin.yaml`中声明`sessions`服务和授权方法：

```yaml
hostServices:
  - service: sessions
    methods:
      - sessions.current
      - sessions.search
      - sessions.batch_get
      - sessions.batch_get_user_online_status
      - sessions.visible.ensure
```

动态插件通过`pluginbridge.Default().Sessions()`客户端调用：

```go
sessionsSvc := pluginbridge.Default().Sessions()

// 获取当前会话视图
current, err := sessionsSvc.Current(ctx)

// 搜索在线会话
page, err := sessionsSvc.List(ctx, sessioncap.ListInput{
    Username: keyword,
    Page:     pageRequest,
})

// 批量读取会话视图
result, err := sessionsSvc.BatchGet(ctx, sessionIDs)

// 批量读取用户在线状态
onlineStatus, err := sessionsSvc.BatchGetUserOnlineStatus(ctx, userIDs)
```

## 设计约束

- **会话吊销经过治理校验。** `Revoke`和`RevokeMany`经过租户、数据范围、目标可见性和审计校验后执行。
- **缺失结果不透露具体原因。** 批量读取不会区分会话不存在、不可见或被拒绝。
- **部门名称是视图字段。** 组织能力不可用时，`DeptName`可能为空。
- **会话吊销即时影响令牌。** 吊销后对应令牌应无法继续通过宿主认证中间件。

## 相关服务

- [Auth能力](/docs/domain-capability-auth)
- [业务上下文能力](/docs/domain-capability-bizctx)
- [Org能力](/docs/domain-capability-org)
