---
slug: '/docs/domain-capability-sessions'
title: 'Sessions（会话管理）'
hide_title: true
description: '`Sessions()`为源码插件和动态插件提供在线会话搜索和批量读取视图。源码插件通过`services.Sessions()`访问，动态插件通过`plugin.yaml`声明`service: sessions`后使用`pluginbridge.Default().Sessions()`客户端访问。可信源码插件可通过`Admin().Sessions().RevokeSession`吊销会话。'
keywords:
  - SessionService
  - sessioncap
  - Sessions
  - 在线会话
  - SearchSessions
  - BatchGetSessions
  - RevokeSession
  - Token管理
  - ClientType
  - DeptName
  - 会话视图
  - 会话吊销
  - AdminServices
  - 插件能力
  - LinaPro
---

## 基本介绍

源码插件通过`services.Sessions()`读取在线会话视图，动态插件通过`plugin.yaml`声明`service: sessions`后使用`pluginbridge.Default().Sessions()`客户端访问。需要吊销会话时，可信源码插件通过`services.Admin().Sessions().RevokeSession`执行受管控的管理命令。

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

### 读写分离设计

会话能力遵循读写分离模式：普通`Sessions()`提供只读视图能力，`Admin().Sessions()`提供受管控的写入命令。会话吊销即时影响令牌，吊销后对应令牌应无法继续通过宿主认证中间件。

### 部门名称视图

`DeptName`是视图字段，组织能力不可用时可能为空。

## 接口定义

### 源码插件接口

| 入口 | 方法 | 说明 |
|------|------|------|
| `Sessions()` | `SearchSessions` | 按用户名、`IP`和分页条件搜索可见会话 |
| `Sessions()` | `BatchGetSessions` | 批量读取可见会话视图 |
| `Admin().Sessions()` | `RevokeSession` | 吊销一个可见在线会话 |

### 动态插件接口

动态插件通过`hostServices.sessions`声明授权的只读方法：

| 动态方法 | 说明 |
|----------|------|
| `sessions.search` | 按用户名、`IP`和分页条件搜索可见会话 |
| `sessions.batch_get` | 批量读取可见会话视图 |

## 能力使用

### 源码插件使用

源码插件通过`services.Sessions()`读取和管理会话，并显式传入领域要求的`CapabilityContext`：

```go
// 搜索在线会话
page, err := services.Sessions().SearchSessions(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// 批量读取会话视图
result, err := services.Sessions().BatchGetSessions(ctx, capabilityCtx, sessionIDs)
```

可信源码插件吊销会话：

```go
err := services.Admin().Sessions().RevokeSession(ctx, capabilityCtx, sessionID)
```

### 动态插件使用

动态插件在`plugin.yaml`中声明`sessions`服务和授权方法：

```yaml
hostServices:
  - service: sessions
    methods:
      - sessions.search
      - sessions.batch_get
```

动态插件通过`pluginbridge.Default().Sessions()`客户端调用：

```go
sessionsSvc := pluginbridge.Default().Sessions()

// 搜索在线会话
page, err := sessionsSvc.SearchSessions(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// 批量读取会话视图
result, err := sessionsSvc.BatchGetSessions(ctx, capabilityCtx, sessionIDs)
```

## 设计约束

- **普通能力只读。** 会话吊销属于管理命令，不在普通`Sessions()`中。
- **缺失结果不透露具体原因。** 批量读取不会区分会话不存在、不可见或被拒绝。
- **部门名称是视图字段。** 组织能力不可用时，`DeptName`可能为空。
- **会话吊销即时影响令牌。** 吊销后对应令牌应无法继续通过宿主认证中间件。

## 相关服务

- [Auth能力](/docs/domain-capability-auth)
- [业务上下文能力](/docs/domain-capability-bizctx)
- [Org能力](/docs/domain-capability-org)
