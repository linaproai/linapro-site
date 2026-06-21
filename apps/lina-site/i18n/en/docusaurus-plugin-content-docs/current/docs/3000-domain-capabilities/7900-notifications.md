---
slug: '/docs/domain-capability-notifications'
title: 'Notifications'
hide_title: true
description: 'The Notifications capability in the latest source code is divided into standard `Notifications()` view reading, trusted source plugin `Admin().Notifications()` management commands, and dynamic plugin `notifications` host service. This page explains the entry differences for notification sending, deletion, and view reading, preventing plugins from continuing to use the legacy root-level `services.Notify()` notation.'
keywords:
  - notification capability
  - notifycap
  - Notifications
  - AdminServices
  - notifications
  - SendInput
  - SendResult
  - SourceType
  - CategoryCode
  - inbox
  - message view
  - dynamic plugin
  - hostServices
  - plugin capability
  - LinaPro
---

## Introduction

The notification capability has three entries:

| Entry | User | Description |
|-------|------|-------------|
| `services.Notifications()` | Source plugin standard capability | Batch-reads visible notification message views |
| `services.Admin().Notifications()` | Trusted source plugin | Sends notifications, deletes notifications, deletes by business source |
| `service: notifications` | Dynamic plugin | Reads message views and sends governed notifications through the `notifications` service |

The latest root capability catalog does not have a `services.Notify()` method. Source plugins that need to send notifications should use management commands, because sending messages produces host write and delivery side effects.

**Capability Phase**: Runtime

**Supported Plugin Types**: Source plugins, Dynamic plugins

## Capability Design

### Data Model

| Type | Field | Description |
|------|-------|-------------|
| `SendInput` | `Recipients` | Target user identifier set |
| `SendInput` | `SourceType`, `SourceID` | Business source type and source record identifier |
| `SendInput` | `Title`, `Content` | Notification title and content |
| `SendInput` | `Category` | Inbox category; falls back to `other` when unspecified |
| `SendInput` | `SenderUserID` | Optional sender; adapter can use `CapabilityContext` operator when omitted |
| `SendResult` | `MessageID`, `DeliveryCount` | Created message identifier and delivery count |

Built-in source types include `notice` and `plugin`; the generic category fallback value is `other`.

### Read-Write Separation Design

The notification capability follows a read-write separation pattern: standard `Notifications()` provides read-only view capabilities, while `Admin().Notifications()` provides governed write commands. Sending is a governed write command because it produces host write and delivery side effects.

### Localization and Sending

Notification content requires plugins to handle localization themselves. Before sending, the i18n capability can be used to resolve copy. The notification capability itself is not responsible for template rendering.

## Interface Definitions

### Source Plugin Interface

| Entry | Method | Description |
|-------|--------|-------------|
| `Notifications()` | `BatchGet` | Batch-reads visible notification message views |
| `Notifications()` | `BatchGetBySource` | Batch-reads visible notification messages by business source |
| `Notifications()` | `EnsureVisible` | Validates that target notification message set is visible to the current call context |
| `Admin().Notifications()` | `Send` | Sends a governed notification |
| `Admin().Notifications()` | `Delete` | Deletes visible notification messages |
| `Admin().Notifications()` | `DeleteBySource` | Deletes notifications by business source |

### Dynamic Plugin Interface

Dynamic plugins declare authorized methods through `hostServices.notifications`:

| Dynamic Method | Description |
|----------------|-------------|
| `messages.batch_get` | Batch-reads visible notification message views |
| `messages.batch_get_by_source` | Batch-reads visible notification messages by business source |
| `messages.visible.ensure` | Validates that target notification message set is visible to the current call context |
| `messages.send` | Sends a governed notification |

## Capability Usage

### Source Plugin Usage

Source plugins read notification views through `services.Notifications()`, explicitly passing the domain-required `CapabilityContext`:

```go
// Batch-read notification messages
result, err := services.Notifications().BatchGet(ctx, capabilityCtx, messageIDs)

// Batch-read notification messages by business source
sourceResult, err := services.Notifications().BatchGetBySource(ctx, capabilityCtx, notifycap.BatchGetBySourceInput{
    SourceType: notifycap.SourceTypePlugin,
    SourceIDs:  []string{reportID1, reportID2},
})

// Validate notification message visibility
err := services.Notifications().EnsureVisible(ctx, capabilityCtx, messageIDs)
```

Trusted source plugins sending and managing notifications:

```go
// Send notification
result, err := services.Admin().Notifications().Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   reportID,
    Title:      "Report Generation Complete",
    Content:    "Your report has been generated, please check",
    Category:   notifycap.CategoryCode("report"),
})

// Delete notification messages
err := services.Admin().Notifications().Delete(ctx, capabilityCtx, messageIDs)

// Delete notifications by business source
err := services.Admin().Notifications().DeleteBySource(ctx, capabilityCtx, notifycap.SourceTypePlugin, []string{reportID})
```

### Dynamic Plugin Usage

Dynamic plugins declare the `notifications` service and authorized methods in `plugin.yaml`:

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

`messages.send` is a resource-type method and must declare `resources[].ref`. The host can use this to restrict the message scenarios or categories a plugin can send. Usage on the dynamic plugin side:

```go
notifySvc := pluginbridge.Default().Notifications()

// Batch-read notification messages
result, err := notifySvc.BatchGet(ctx, capabilityCtx, messageIDs)

// Batch-read notification messages by business source
sourceResult, err := notifySvc.BatchGetBySource(ctx, capabilityCtx, notifycap.BatchGetBySourceInput{
    SourceType: notifycap.SourceTypePlugin,
    SourceIDs:  []string{reportID1, reportID2},
})

// Send notification
sendResult, err := notifySvc.Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   taskID,
    Title:      "Task Complete",
    Content:    "Your export task has been completed",
    Category:   notifycap.CategoryCode("report"),
})
```

## Design Constraints

- **Standard capability is read-only.** `Notifications()` is for reading views, not for sending.
- **Sending is a governed write command.** Source plugins send notifications via `Admin().Notifications().Send`; dynamic plugins via `notifications.messages.send`.
- **Deletion is governed by visibility and source.** Both `DeleteMessages` and `DeleteBySource` should check target visibility in the host domain adapter.
- **Notification content should be self-localized.** The i18n capability can be used to resolve copy before sending; the notification capability is not responsible for template rendering.

## Related Services

- [i18n Capability](/docs/domain-capability-i18n)
- [Business Context Capability](/docs/domain-capability-bizctx)
- [Domain Capabilities Overview](/docs/domain-capabilities)
