---
slug: '/docs/plugin-capability-session'
title: 'SessionService'
hide_title: true
description: 'An architectural overview of LinaPro SessionService — the online session management service covering its design positioning, Session projection model, and role in session governance, helping plugin developers understand how to query and manage online user sessions.'
keywords:
  - SessionService
  - session service
  - online session
  - session management
  - session projection
  - session query
  - session revocation
  - token management
  - plugin capability
  - capability.Services
  - session pagination
  - user session
  - login session
  - session kick
  - LinaPro
---

## Introduction

`SessionService` provides plugins with the ability to query and manage online sessions. Plugins obtain the service via `services.Session()` to paginate through online session lists and revoke sessions by `TokenID`.

A typical consumer is the `linapro-monitor-online` plugin, which provides online user monitoring functionality, displaying currently online users through `SessionService` and allowing administrators to kick sessions.

## Design Philosophy

`SessionService` is built around two core capabilities: **session projection** and **session management**.

**Session projection.** The `Session` struct is a stable projection of an online session, containing the following fields:

| Field | Description |
|-------|-------------|
| `TokenId` | Unique token identifier for the session |
| `TenantId` | Tenant the session belongs to; `0` means platform |
| `UserId` | Authenticated user ID |
| `Username` | Authenticated username |
| `ClientType` | Client type |
| `DeptName` | Department name projection |
| `Ip` | Login IP address |
| `Browser` | Browser fingerprint |
| `Os` | Operating system fingerprint |
| `LoginTime` | Initial login time |
| `LastActiveTime` | Most recent activity time |

**Session filtering.** `ListFilter` supports fuzzy matching by username and login IP, useful for administrators searching for specific users among a large number of sessions.

```mermaid
graph LR
    Admin["Admin"] -->|"Query"| SessionService["SessionService"]
    SessionService -->|"ListPage"| SessionList["Online Session List"]
    SessionService -->|"Revoke"| Revoke["Revoke Session"]
    Revoke --> Kick["User Kicked Out"]
```

## Architectural Position

`SessionService` sits at the query and management layer in the session management chain, complementing the session registration layer of `AuthService`:

```mermaid
graph TB
    subgraph AuthFlow["Auth Flow"]
        AuthService["AuthService"] -->|"Issue Token"| SessionStore["Session Store"]
        AuthService -->|"Revoke Token"| SessionStore
    end

    subgraph ManagementFlow["Management Flow"]
        SessionService["SessionService"] -->|"Query"| SessionStore
        SessionService -->|"Revoke"| SessionStore
    end

    SessionStore -->|"Project"| Session["Session Struct"]
```

- `AuthService` registers and revokes sessions in the authentication flow
- `SessionService` queries and revokes sessions in the management flow
- Both share the same session store, but with different responsibilities

## Key Capabilities

| Method | Description |
|--------|-------------|
| `ListPage` | Paginated query of online sessions, with fuzzy filtering by username and IP |
| `Revoke` | Revoke a single online session by token ID, kicking the user out |

## Design Constraints

- **Sessions are read-only projections.** The `Session` struct is read-only; modifying it does not change session state. Use `Revoke` to terminate a session.
- **Paginated queries are scoped by tenant.** Platform administrators can query sessions across all tenants; tenant administrators can only query sessions within their own tenant.
- **Revocation takes effect immediately.** After `Revoke` is executed, the corresponding token is invalidated instantly, and the client's next request will be rejected.
- **`DeptName` is a projection field.** The department name is projected from the organization capability provider; if the organization capability is unavailable, this field may be empty.

## Related Services

- [AuthService](/docs/plugin-capability-auth) - AuthService registers sessions in the authentication flow; SessionService queries sessions in the management flow
- [BizCtxService](/docs/plugin-capability-bizctx) - Current request session information is projected into BizCtx
- [OrgService](/docs/plugin-capability-org) - The DeptName field in Session is projected from the organization capability
