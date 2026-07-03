---
slug: '/docs/domain-capability-notifications'
title: 'Notifications（通知能力）'
hide_title: true
description: '通知能力通过`Notifications()`统一提供视图读取、发送、删除和标记操作；动态插件通过`notifications`宿主服务访问已发布的方法子集。该页面说明通知能力的接口和使用方式。'
keywords:
  - 通知能力
  - notifycap
  - Notifications
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

通知能力有两个入口：

| 入口 | 使用者 | 说明 |
|------|--------|------|
| `services.Notifications()` | 源码插件 | 读取通知视图、发送通知、删除通知、标记已读 |
| `service: notifications` | 动态插件 | 通过`notifications`宿主服务读取消息视图和发送受管控的通知 |

根能力目录没有`services.Notify()`方法。所有通知操作统一通过`services.Notifications()`访问。

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

### 读写操作

`Notifications()`统一提供读取和写入操作。发送、删除等写操作经过收件人、来源、分类、资源授权、租户边界和审计校验后执行。

### 本地化与发送

通知内容需要插件自行处理本地化。发送前可借助国际化能力解析文案，通知能力本身不负责模板渲染。

## 接口定义

### 源码插件接口

| 方法 | 说明 |
|------|------|
| `Get` | 读取单个可见通知消息视图 |
| `BatchGet` | 批量读取可见通知消息视图 |
| `List` | 按分页条件搜索可见通知消息 |
| `BatchGetBySource` | 按业务来源批量读取可见通知消息 |
| `EnsureVisible` | 校验目标通知消息集合对当前调用上下文可见 |
| `Send` | 发送受管控的通知，经过收件人、来源、分类和资源授权校验 |
| `Delete` | 删除可见通知消息 |
| `DeleteBySource` | 按业务来源删除通知 |
| `MarkRead` | 标记通知消息为已读 |
| `MarkUnread` | 标记通知消息为未读 |

### 动态插件接口

动态插件通过`hostServices.notifications`声明授权的方法：

| 动态方法 | 说明 |
|----------|------|
| `messages.get` | 读取单个可见通知消息视图 |
| `messages.batch_get` | 批量读取可见通知消息视图 |
| `messages.list` | 按分页条件搜索可见通知消息 |
| `messages.batch_get_by_source` | 按业务来源批量读取可见通知消息 |
| `messages.visible.ensure` | 校验目标通知消息集合对当前调用上下文可见 |
| `messages.send` | 发送受管控的通知 |
| `messages.mark_read` | 标记通知消息为已读 |
| `messages.mark_unread` | 标记通知消息为未读 |

## 能力使用

### 源码插件使用

源码插件通过`services.Notifications()`访问通知能力：

```go
// 批量读取通知消息
result, err := services.Notifications().BatchGet(ctx, messageIDs)

// 按业务来源批量读取通知消息
sourceResult, err := services.Notifications().BatchGetBySource(ctx, notifycap.BatchGetBySourceInput{
    SourceType: notifycap.SourceTypePlugin,
    SourceIDs:  []string{reportID1, reportID2},
})

// 校验通知消息可见性
err := services.Notifications().EnsureVisible(ctx, messageIDs)

// 发送通知
result, err := services.Notifications().Send(ctx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   reportID,
    Title:      "报告生成完成",
    Content:    "您的报告已生成，请查看",
    Category:   notifycap.CategoryCode("report"),
})

// 删除通知消息
err := services.Notifications().Delete(ctx, messageIDs)

// 按业务来源删除通知
err := services.Notifications().DeleteBySource(ctx, notifycap.SourceTypePlugin, []string{reportID})
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

- **发送经过治理校验。** `Send`经过收件人、来源、分类和资源授权校验后执行。
- **删除按可见性和来源治理。** `Delete`和`DeleteBySource`都应在宿主领域适配器中检查目标可见性。
- **已读标记为轻量操作。** `MarkRead`和`MarkUnread`用于通知状态管理。
- **通知内容应自行本地化。** 发送前可以使用国际化能力解析文案，通知能力不负责模板渲染。

## 相关服务

- [国际化能力](/docs/domain-capability-i18n)
- [业务上下文能力](/docs/domain-capability-bizctx)
- [插件可用领域能力概览](/docs/domain-capabilities)
