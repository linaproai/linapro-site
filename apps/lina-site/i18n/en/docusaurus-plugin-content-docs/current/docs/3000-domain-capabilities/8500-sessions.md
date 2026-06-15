---
slug: '/docs/domain-capability-sessions'
title: 'Sessions'
hide_title: true
description: '`Sessions()` provides online session search and batch-read views for source plugins and dynamic plugins. Source plugins access it through `services.Sessions()`, while dynamic plugins declare `service: sessions` in `plugin.yaml` and access it via `pluginbridge.Default().Sessions()`. Trusted source plugins can revoke sessions through `Admin().Sessions().RevokeSession`.'
keywords:
  - SessionService
  - sessioncap
  - Sessions
  - online sessions
  - SearchSessions
  - BatchGetSessions
  - RevokeSession
  - Token management
  - ClientType
  - DeptName
  - session views
  - session revocation
  - AdminServices
  - plugin capability
  - LinaPro
---

## Overview

Source plugins read online session views through `services.Sessions()`, while dynamic plugins declare `service: sessions` in `plugin.yaml` and access it via `pluginbridge.Default().Sessions()`. When session revocation is needed, trusted source plugins execute the governed admin command through `services.Admin().Sessions().RevokeSession`.

**Capability Phase**: Runtime

**Supported Types**: Source plugins, dynamic plugins

## Capability Design

### Session View Model

Session views are designed for online user monitoring, session governance, and security audit scenarios. They do not expose session storage tables or JWT internals:

| Field | Description |
|-------|-------------|
| `ID` | Session domain identifier |
| `TenantID` | Current tenant identifier |
| `UserID`, `Username` | Session user |
| `ClientType` | Client type, e.g. `web`, `mobile`, `desktop`, `cli` |
| `DeptName` | Department name captured or assembled at login |
| `Ip`, `Browser`, `Os` | Login environment information |
| `LoginAt`, `LastActiveAt` | Login time and most recent activity time |

### Read-write Separation Design

The session capability follows a read-write separation pattern: the standard `Sessions()` provides read-only view capability, while `Admin().Sessions()` provides governed write commands. Session revocation takes effect on the token immediately; after revocation, the corresponding token should no longer pass the host authentication middleware.

### Department Name View

`DeptName` is a view field and may be empty when the org capability is unavailable.

## Interface Definitions

### Source Plugin Interface

| Entry Point | Method | Description |
|-------------|--------|-------------|
| `Sessions()` | `SearchSessions` | Searches visible sessions by username, IP, and pagination |
| `Sessions()` | `BatchGetSessions` | Batch-reads visible session views |
| `Admin().Sessions()` | `RevokeSession` | Revokes a visible online session |

### Dynamic Plugin Interface

Dynamic plugins declare authorized read-only methods through `hostServices.sessions`:

| Dynamic Method | Description |
|----------------|-------------|
| `sessions.search` | Searches visible sessions by username, IP, and pagination |
| `sessions.batch_get` | Batch-reads visible session views |

## Usage

### Source Plugin Usage

Source plugins read and manage sessions through `services.Sessions()`, explicitly passing the domain-required `CapabilityContext`:

```go
// Search online sessions
page, err := services.Sessions().SearchSessions(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-read session views
result, err := services.Sessions().BatchGetSessions(ctx, capabilityCtx, sessionIDs)
```

Trusted source plugins revoke sessions:

```go
err := services.Admin().Sessions().RevokeSession(ctx, capabilityCtx, sessionID)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `sessions` service and authorized methods in `plugin.yaml`:

```yaml
hostServices:
  - service: sessions
    methods:
      - sessions.search
      - sessions.batch_get
```

Dynamic plugins call through `pluginbridge.Default().Sessions()`:

```go
sessionsSvc := pluginbridge.Default().Sessions()

// Search online sessions
page, err := sessionsSvc.SearchSessions(ctx, capabilityCtx, sessioncap.SearchInput{
    Username: keyword,
    Page:     pageRequest,
})

// Batch-read session views
result, err := sessionsSvc.BatchGetSessions(ctx, capabilityCtx, sessionIDs)
```

## Design Constraints

- **Standard capability is read-only.** Session revocation is an admin command and is not part of the standard `Sessions()`.
- **Missing results do not reveal specific reasons.** Batch reads do not distinguish between sessions that do not exist, are not visible, or are denied.
- **Department name is a view field.** `DeptName` may be empty when the org capability is unavailable.
- **Session revocation takes effect on the token immediately.** After revocation, the corresponding token should no longer pass the host authentication middleware.

## Related Services

- [Auth Capability](/docs/domain-capability-auth)
- [Business Context Capability](/docs/domain-capability-bizctx)
- [Org Capability](/docs/domain-capability-org)
