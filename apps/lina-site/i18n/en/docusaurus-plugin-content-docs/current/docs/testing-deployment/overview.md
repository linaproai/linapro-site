---
slug: '/docs/testing-deployment'
title: '🧪 Testing and Deployment'
sidebar_position: 0
description: 'This testing and deployment page summarizes the LinaPro local command workflow, safety confirmations for database-changing commands, installer smoke tests, E2E suite expectations, build preparation, service ports, operational checks, and deployment-minded practices that should be verified before sharing a production change.'
keywords:
  - LinaPro
  - testing
  - deployment
  - local commands
  - make init
  - make mock
  - make dev
  - make test
  - make test-install
  - E2E tests
  - Playwright
  - database initialization
  - build preparation
  - service ports
  - operational checks
  - production readiness
---

Use local commands consistently so setup, verification, and review are repeatable.

## Common Commands

| Command | Purpose |
| --- | --- |
| `make init confirm=init` | Initialize database schema and seed data. |
| `make init confirm=init rebuild=true` | Rebuild the local database intentionally. |
| `make mock confirm=mock` | Load optional demo data after initialization. |
| `make dev` | Start backend and frontend services locally. |
| `make stop` | Stop local services. |
| `make status` | Show local service status. |
| `make test-install` | Run installer smoke tests. |
| `make test` | Run the full `E2E` test suite. |

Commands that change persistent state use explicit confirmation arguments so accidental data changes fail early.

## Default Local Ports

| Service | Port |
| --- | --- |
| Management workspace | `5666` |
| Core host `API` | `8080` |

## Verification Checklist

Before preparing a change for review:

1. Run the narrowest relevant test first.
2. Run `make test` when shared workflows, permissions, plugins, or user-visible behavior changed.
3. Check the management workspace and `API` docs manually when route, menu, or permission behavior changed.
4. Record any skipped verification with a concrete reason.

Deployment-specific documentation will expand as public release packaging stabilizes.
