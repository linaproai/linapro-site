---
slug: '/docs/domain-capability-sessions'
title: 'Sessions'
hide_title: true
description: 'Sessions() provides source-code plugins with online session search, batch retrieval, and revocation capabilities. Dynamic plugins declare service: sessions in plugin.yaml and use the pluginbridge.Default().Sessions() client to access published read-only methods.'
keywords:
  - SessionService
  - sessioncap
  - Sessions
  - online sessions
  - SearchSessions
  - BatchGetSessions
  - Revoke
  - token management
  - ClientType
  - DeptName
  - session view
  - session revocation
  - plugin capabilities
  - LinaPro
---

## Overview

Source-code plugins access online session read and revocation capabilities through `services.Sessions()`. Dynamic plugins declare `service: sessions` in `plugin.yaml` and use the `pluginbridge.Default().Sessions()` client to access published read-only methods.

**Capability Phase**: Runtime

**Supported Types**: Source-code plugins, dynamic plugins

## Capability Design

### Session View Model

The session view is designed for online user monitoring, session governance, and security auditing. It does not expose session storage tables or `JWT` internals:

| Field | Description |
|------|------|
| `ID` | Session domain identifier |
| `TenantID` | Current tenant identifier |
| `UserID`, `Username` | Session user |
| `ClientType` | Client type, e.g. `web`, `mobile`, `desktop`, `cli` |
| `DeptName` | Department name captured or assembled at login time |
| `Ip`, `Browser`, `Os` | Login environment information |
| `LoginAt`, `LastActiveAt` | Login time and last active time |

### Read and Write Operations

`Sessions()` provides unified read and revocation operations. `Revoke` and `RevokeMany` are executed after tenant, data scope, target visibility, and audit validation. Session revocation takes immediate effect on the token -- after revocation, the corresponding token should no longer pass host authentication middleware.

### Department Name View

`DeptName` is a view field and may be empty when the organization capability is unavailable.

## Interface Definitions

### Source-Code Plugin Interface

| Method | Description |
|------|------|
| `Current` | Returns the visible session view for the current token |
| `Get` | Retrieves a single visible session view |
| `List` | Searches visible sessions by username, `IP`, and pagination |
| `BatchGet` | Batch-retrieves visible session views |
| `BatchGetUserOnlineStatus` | Batch-retrieves user online status |
| `EnsureVisible` | Validates that a target session set is visible to the current caller |
| `Revoke` | Revokes a visible online session, subject to tenant, data scope, target visibility, and audit validation |
| `RevokeMany` | Batch-revokes visible online sessions; any invisible target rejects the entire operation |

### Dynamic Plugin Interface

Dynamic plugins declare authorized read-only methods through `hostServices.sessions`:

| Dynamic Method | Description |
|----------|------|
| `sessions.current` | Returns the visible session view for the current token |
| `sessions.list` | Searches visible sessions by username, `IP`, and pagination |
| `sessions.batch_get` | Batch-retrieves visible session views |
| `sessions.batch_get_user_online_status` | Batch-retrieves user online status |
| `sessions.visible.ensure` | Validates that a target session set is visible to the current caller |

## Usage

### Source-Code Plugin Usage

Source-code plugins use `services.Sessions()` to read and manage sessions, passing the domain-required `CapabilityContext` explicitly:

```go
// Get the current session view
current, err := services.Sessions().Current(ctx)

// Search online sessions
page, err := services.Sessions().List(ctx, sessioncap.ListInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-retrieve session views
result, err := services.Sessions().BatchGet(ctx, sessionIDs)

// Batch-retrieve user online status
onlineStatus, err := services.Sessions().BatchGetUserOnlineStatus(ctx, userIDs)

// Validate session visibility
err := services.Sessions().EnsureVisible(ctx, sessionIDs)

// Revoke a single session
err := services.Sessions().Revoke(ctx, sessionID)

// Batch-revoke sessions
err := services.Sessions().RevokeMany(ctx, sessionIDs)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `sessions` service and authorized methods in `plugin.yaml`:

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

Dynamic plugins invoke through the `pluginbridge.Default().Sessions()` client:

```go
sessionsSvc := pluginbridge.Default().Sessions()

// Get the current session view
current, err := sessionsSvc.Current(ctx)

// Search online sessions
page, err := sessionsSvc.List(ctx, sessioncap.ListInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-retrieve session views
result, err := sessionsSvc.BatchGet(ctx, sessionIDs)

// Batch-retrieve user online status
onlineStatus, err := sessionsSvc.BatchGetUserOnlineStatus(ctx, userIDs)
```

## Design Constraints

- **Session revocation is subject to governance validation.** `Revoke` and `RevokeMany` are executed after tenant, data scope, target visibility, and audit validation.
- **Missing results do not reveal specific reasons.** Batch retrieval does not distinguish between sessions that do not exist, are invisible, or are denied.
- **Department name is a view field.** `DeptName` may be empty when the organization capability is unavailable.
- **Session revocation takes immediate effect on the token.** After revocation, the corresponding token should no longer pass host authentication middleware.

## Related Services

- [Auth Capability](/docs/domain-capability-auth)
- [Business Context Capability](/docs/domain-capability-bizctx)
- [Org Capability](/docs/domain-capability-org)
