---
slug: '/docs/domain-capability-notifications'
title: 'Notifications（通知能力）'
hide_title: true
description: '通知能力在最新源码中分为普通`Notifications()`视图读取、可信源码插件`Admin().Notifications()`管理命令和动态插件`notifications`宿主服务。该页面说明通知发送、删除和视图读取的入口差异，避免插件继续使用旧的根级`services.Notify()`表述。'
keywords:
  - 通知能力
  - notifycap
  - Notifications
  - AdminServices
  - notifications
  - SendInput
  - SendResult
  - SourceType
  - CategoryCode
  - 收件箱
  - 消息视图
  - 动态插件
  - hostServices
  - 插件能力
  - LinaPro
---

## 基本介绍

通知能力有三个入口：

| 入口 | 使用者 | 说明 |
|------|--------|------|
| `services.Notifications()` | 源码插件普通能力 | 批量读取可见通知消息视图 |
| `services.Admin().Notifications()` | 可信源码插件 | 发送通知、删除通知、按业务来源删除通知 |
| `service: notifications` | 动态插件 | 通过`notifications`服务读取消息视图和发送受管控的通知 |

最新根能力目录没有`services.Notify()`方法。需要发送通知的源码插件应使用管理命令，因为发送消息会产生宿主写入和投递副作用。

**能力阶段**：运行期

**类型支持**：源码插件、动态插件

## 能力设计

### 数据模型

| 类型 | 字段 | 说明 |
|------|------|------|
| `SendInput` | `Recipients` | 目标用户标识集合 |
| `SendInput` | `SourceType`、`SourceID` | 业务来源类型和来源记录标识 |
| `SendInput` | `Title`、`Content` | 通知标题和内容 |
| `SendInput` | `Category` | 收件箱分类，未指定时可回退到`other` |
| `SendInput` | `SenderUserID` | 可选发送者，缺省时适配器可使用`CapabilityContext`操作者 |
| `SendResult` | `MessageID`、`DeliveryCount` | 创建的消息标识和投递数量 |

内置来源类型包括`notice`和`plugin`，通用分类回退值为`other`。

### 读写分离设计

通知能力遵循读写分离模式：普通`Notifications()`提供只读视图能力，`Admin().Notifications()`提供受管控的写入命令。发送属于受管控的写命令，因为会产生宿主写入和投递副作用。

### 本地化与发送

通知内容需要插件自行处理本地化。发送前可借助国际化能力解析文案，通知能力本身不负责模板渲染。

## 接口定义

### 源码插件接口

| 入口 | 方法 | 说明 |
|------|------|------|
| `Notifications()` | `BatchGet` | 批量读取可见通知消息视图 |
| `Notifications()` | `BatchGetBySource` | 按业务来源批量读取可见通知消息 |
| `Notifications()` | `EnsureVisible` | 校验目标通知消息集合对当前调用上下文可见 |
| `Admin().Notifications()` | `Send` | 发送受管控的通知 |
| `Admin().Notifications()` | `Delete` | 删除可见通知消息 |
| `Admin().Notifications()` | `DeleteBySource` | 按业务来源删除通知 |

### 动态插件接口

动态插件通过`hostServices.notifications`声明授权的方法：

| 动态方法 | 说明 |
|----------|------|
| `messages.batch_get` | 批量读取可见通知消息视图 |
| `messages.batch_get_by_source` | 按业务来源批量读取可见通知消息 |
| `messages.visible.ensure` | 校验目标通知消息集合对当前调用上下文可见 |
| `messages.send` | 发送受管控的通知 |

## 能力使用

### 源码插件使用

源码插件通过`services.Notifications()`读取通知视图，并显式传入领域要求的`CapabilityContext`：

```go
// 批量读取通知消息
result, err := services.Notifications().BatchGet(ctx, capabilityCtx, messageIDs)

// 按业务来源批量读取通知消息
sourceResult, err := services.Notifications().BatchGetBySource(ctx, capabilityCtx, notifycap.BatchGetBySourceInput{
    SourceType: notifycap.SourceTypePlugin,
    SourceIDs:  []string{reportID1, reportID2},
})

// 校验通知消息可见性
err := services.Notifications().EnsureVisible(ctx, capabilityCtx, messageIDs)
```

可信源码插件发送和管理通知：

```go
// 发送通知
result, err := services.Admin().Notifications().Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   reportID,
    Title:      "报告生成完成",
    Content:    "您的报告已生成，请查看",
    Category:   notifycap.CategoryCode("report"),
})

// 删除通知消息
err := services.Admin().Notifications().Delete(ctx, capabilityCtx, messageIDs)

// 按业务来源删除通知
err := services.Admin().Notifications().DeleteBySource(ctx, capabilityCtx, notifycap.SourceTypePlugin, []string{reportID})
```

### 动态插件使用

动态插件在`plugin.yaml`中声明`notifications`服务和授权方法：

```yaml
hostServices:
  - service: notifications
    methods:
      - messages.batch_get
      - messages.batch_get_by_source
      - messages.visible.ensure
      - messages.send
    resources:
      - ref: inbox:business-alert
```

`messages.send`是资源型方法，必须声明`resources[].ref`。宿主可据此限制插件可发送的消息场景或分类。在动态插件侧使用：

```go
notifySvc := pluginbridge.Default().Notifications()

// 批量读取通知消息
result, err := notifySvc.BatchGet(ctx, capabilityCtx, messageIDs)

// 按业务来源批量读取通知消息
sourceResult, err := notifySvc.BatchGetBySource(ctx, capabilityCtx, notifycap.BatchGetBySourceInput{
    SourceType: notifycap.SourceTypePlugin,
    SourceIDs:  []string{reportID1, reportID2},
})

// 发送通知
sendResult, err := notifySvc.Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   taskID,
    Title:      "任务完成",
    Content:    "您的导出任务已完成",
    Category:   notifycap.CategoryCode("report"),
})
```

## 设计约束

- **普通能力只读。** `Notifications()`用于读取视图，不负责发送。
- **发送属于受管控的写命令。** 源码插件发送通知走`Admin().Notifications().Send`，动态插件走`notifications.messages.send`。
- **删除按可见性和来源治理。** `DeleteMessages`和`DeleteBySource`都应在宿主领域适配器中检查目标可见性。
- **通知内容应自行本地化。** 发送前可以使用国际化能力解析文案，通知能力不负责模板渲染。

## 相关服务

- [国际化能力](/docs/domain-capability-i18n)
- [业务上下文能力](/docs/domain-capability-bizctx)
- [插件可用领域能力概览](/docs/domain-capabilities)
