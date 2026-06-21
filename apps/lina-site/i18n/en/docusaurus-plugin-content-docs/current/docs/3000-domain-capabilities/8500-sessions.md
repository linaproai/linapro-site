---
slug: '/docs/domain-capability-sessions'
title: 'Sessions'
hide_title: true
description: '`Sessions()` provides online session search and batch reading views for source plugins and dynamic plugins. Source plugins access it through `services.Sessions()`, while dynamic plugins declare `service: sessions` in `plugin.yaml` and access it through the `pluginbridge.Default().Sessions()` client. Trusted source plugins can revoke sessions through `Admin().Sessions().RevokeSession`.'
keywords:
  - SessionService
  - sessioncap
  - Sessions
  - online sessions
  - SearchSessions
  - BatchGetSessions
  - RevokeSession
  - token management
  - ClientType
  - DeptName
  - session view
  - session revocation
  - AdminServices
  - plugin capability
  - LinaPro
---

## Introduction

Source plugins read online session views through `services.Sessions()`. Dynamic plugins declare `service: sessions` in `plugin.yaml` and access it through the `pluginbridge.Default().Sessions()` client. When session revocation is needed, trusted source plugins execute governed management commands through `services.Admin().Sessions().RevokeSession`.

**Capability Phase**: Runtime

**Supported Plugin Types**: Source plugins, Dynamic plugins

## Capability Design

### Session View Model

Session views are used for online user monitoring, session governance, and security auditing scenarios. They do not expose session storage tables or `JWT` internal implementation:

| Field | Description |
|-------|-------------|
| `ID` | Session domain identifier |
| `TenantID` | Current tenant identifier |
| `UserID`, `Username` | Session user |
| `ClientType` | Client type, e.g., `web`, `mobile`, `desktop`, `cli` |
| `DeptName` | Department name captured or assembled at login |
| `Ip`, `Browser`, `Os` | Login environment info |
| `LoginAt`, `LastActiveAt` | Login time and last active time |

### Read-Write Separation Design

The session capability follows a read-write separation pattern: standard `Sessions()` provides read-only view capabilities, while `Admin().Sessions()` provides governed write commands. Session revocation immediately affects tokens — after revocation, the corresponding token should not be able to pass the host authentication middleware.

### Department Name View

`DeptName` is a view field and may be empty when the org capability is not available.

## Interface Definitions

### Source Plugin Interface

| Entry | Method | Description |
|-------|--------|-------------|
| `Sessions()` | `Current` | Returns the visible session view for the current token |
| `Sessions()` | `Search` | Searches visible sessions by username, `IP`, and pagination |
| `Sessions()` | `BatchGet` | Batch-reads visible session views |
| `Sessions()` | `BatchGetUserOnlineStatus` | Batch-reads user online status |
| `Sessions()` | `EnsureVisible` | Validates that target session set is visible to the current call context |
| `Admin().Sessions()` | `Revoke` | Revokes a visible online session |

### Dynamic Plugin Interface

Dynamic plugins declare authorized read-only methods through `hostServices.sessions`:

| Dynamic Method | Description |
|----------------|-------------|
| `sessions.current` | Returns the visible session view for the current token |
| `sessions.search` | Searches visible sessions by username, `IP`, and pagination |
| `sessions.batch_get` | Batch-reads visible session views |
| `sessions.batch_get_user_online_status` | Batch-reads user online status |
| `sessions.visible.ensure` | Validates that target session set is visible to the current call context |

## Capability Usage

### Source Plugin Usage

Source plugins read and manage sessions through `services.Sessions()`, explicitly passing the domain-required `CapabilityContext`:

```go
// Get current session view
current, err := services.Sessions().Current(ctx, capabilityCtx)

// Search online sessions
page, err := services.Sessions().Search(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-read session views
result, err := services.Sessions().BatchGet(ctx, capabilityCtx, sessionIDs)

// Batch-read user online status
onlineStatus, err := services.Sessions().BatchGetUserOnlineStatus(ctx, capabilityCtx, userIDs)

// Validate session visibility
err := services.Sessions().EnsureVisible(ctx, capabilityCtx, sessionIDs)
```

Trusted source plugins revoking sessions:

```go
err := services.Admin().Sessions().Revoke(ctx, capabilityCtx, sessionID)
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

Dynamic plugins call through the `pluginbridge.Default().Sessions()` client:

```go
sessionsSvc := pluginbridge.Default().Sessions()

// Get current session view
current, err := sessionsSvc.Current(ctx, capabilityCtx)

// Search online sessions
page, err := sessionsSvc.Search(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-read session views
result, err := sessionsSvc.BatchGet(ctx, capabilityCtx, sessionIDs)

// Batch-read user online status
onlineStatus, err := sessionsSvc.BatchGetUserOnlineStatus(ctx, capabilityCtx, userIDs)
```

## Design Constraints

- **Standard capability is read-only.** Session revocation is a management command, not part of standard `Sessions()`.
- **Missing results don't reveal specific reasons.** Batch reads don't distinguish between sessions that don't exist, are invisible, or are rejected.
- **Department name is a view field.** `DeptName` may be empty when the org capability is not available.
- **Session revocation immediately affects tokens.** After revocation, the corresponding token should not be able to pass the host authentication middleware.

## Related Services

- [Auth Capability](/docs/domain-capability-auth)
- [Business Context Capability](/docs/domain-capability-bizctx)
- [Org Capability](/docs/domain-capability-org)
