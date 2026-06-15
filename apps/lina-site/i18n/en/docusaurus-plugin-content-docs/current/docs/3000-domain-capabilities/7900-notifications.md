---
slug: '/docs/domain-capability-notifications'
title: 'Notifications'
hide_title: true
description: 'The notifications capability in the latest source code is split into the standard `Notifications()` view read capability, trusted source plugin `Admin().Notifications()` management commands, and the dynamic plugin `notifications` host service. This page explains the entry point differences for notification sending, deletion, and view reading, to help plugins move away from the legacy root-level `services.Notify()` usage.'
keywords:
  - notifications capability
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

## Overview

The notifications capability has three entry points:

| Entry Point | User | Description |
|-------------|------|-------------|
| `services.Notifications()` | Source plugin standard capability | Batch reads visible notification message views |
| `services.Admin().Notifications()` | Trusted source plugins | Sends notifications, deletes notifications, deletes by business source |
| `service: notifications` | Dynamic plugins | Reads message views and sends managed notifications through the `notifications` service |

The latest root capability catalog does not have a `services.Notify()` method. Source plugins that need to send notifications should use the management commands, because sending messages produces host write and delivery side effects.

**Capability phase**: Runtime

**Type support**: Source plugins, dynamic plugins

## Capability Design

### Data Model

| Type | Field | Description |
|------|-------|-------------|
| `SendInput` | `Recipients` | Set of target user identifiers |
| `SendInput` | `SourceType`, `SourceID` | Business source type and source record identifier |
| `SendInput` | `Title`, `Content` | Notification title and body |
| `SendInput` | `Category` | Inbox category; falls back to `other` when unspecified |
| `SendInput` | `SenderUserID` | Optional sender; adapter may use the `CapabilityContext` operator when omitted |
| `SendResult` | `MessageID`, `DeliveryCount` | Created message identifier and delivery count |

Built-in source types include `notice` and `plugin`. The generic category fallback value is `other`.

### Read/Write Separation Design

The notifications capability follows a read/write separation pattern: the standard `Notifications()` provides a read-only view capability, while `Admin().Notifications()` provides managed write commands. Sending is a managed write command because it produces host write and delivery side effects.

### Localization and Sending

Notification content requires plugins to handle localization themselves. Before sending, you can use the internationalization capability to resolve copy; the notifications capability itself does not handle template rendering.

## Interface Definitions

### Source Plugin Interface

| Entry Point | Method | Description |
|-------------|--------|-------------|
| `Notifications()` | `BatchGetMessages` | Batch reads visible notification message views |
| `Admin().Notifications()` | `Send` | Sends a managed notification |
| `Admin().Notifications()` | `DeleteMessages` | Deletes visible notification messages |
| `Admin().Notifications()` | `DeleteBySource` | Deletes notifications by business source |

### Dynamic Plugin Interface

Dynamic plugins declare authorized methods through `hostServices.notifications`:

| Dynamic Method | Description |
|----------------|-------------|
| `messages.batch_get` | Batch reads visible notification message views |
| `messages.send` | Sends a managed notification |

## Usage

### Source Plugin Usage

Source plugins read notification views through `services.Notifications()`, passing the domain-required `CapabilityContext` explicitly:

```go
// Batch read notification messages
result, err := services.Notifications().BatchGetMessages(ctx, capabilityCtx, messageIDs)
```

Trusted source plugins send and manage notifications:

```go
// Send notification
result, err := services.Admin().Notifications().Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   reportID,
    Title:      "Report Generation Complete",
    Content:    "Your report has been generated, please check it",
    Category:   notifycap.CategoryCode("report"),
})

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
      - messages.send
    resources:
      - ref: inbox:business-alert
```

`messages.send` is a resource-type method and must declare `resources[].ref`. The host can use this to restrict the message scenarios or categories a plugin can send. Usage on the dynamic plugin side:

```go
notifySvc := pluginbridge.Default().Notifications()

// Batch read notification messages
result, err := notifySvc.BatchGetMessages(ctx, capabilityCtx, messageIDs)

// Send notification
sendResult, err := notifySvc.Send(ctx, capabilityCtx, notifycap.SendInput{
    Recipients: []string{userID},
    SourceType: notifycap.SourceTypePlugin,
    SourceID:   taskID,
    Title:      "Task Complete",
    Content:    "Your export task has completed",
    Category:   notifycap.CategoryCode("report"),
})
```

## Design Constraints

- **Standard capability is read-only.** `Notifications()` is for reading views and is not responsible for sending.
- **Sending is a managed write command.** Source plugins send notifications via `Admin().Notifications().Send`; dynamic plugins use `notifications.messages.send`.
- **Deletion is governed by visibility and source.** Both `DeleteMessages` and `DeleteBySource` should check target visibility in the host domain adapter.
- **Notification content should be localized by the plugin.** You can use the internationalization capability to resolve copy before sending; the notifications capability does not handle template rendering.

## Related Services

- [I18n Capability](/docs/domain-capability-i18n)
- [Business Context Capability](/docs/domain-capability-bizctx)
- [Domain Capabilities Overview](/docs/domain-capabilities)
