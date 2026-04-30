---
slug: '/quickstart'
title: 'Quick Start'
sidebar_position: 1
description: 'This LinaPro quick start walks new developers through the minimum local setup: installing the framework source, initializing the database, loading optional demo data, starting the Go backend and Vue management workspace, signing in with the default account, and checking the core API and plugin runtime before moving into deeper development manuals.'
keywords:
  - LinaPro
  - quick start
  - AI-native full-stack framework
  - local development
  - installer
  - make init
  - make mock
  - make dev
  - Go backend
  - Vue 3 workspace
  - lina-core
  - lina-vben
  - plugin runtime
  - OpenAPI docs
  - RBAC
  - demo data
---

Get `LinaPro` running locally in three commands. When you finish this page, the core host service, management workspace, and plugin runtime will be ready for development.

## Prerequisites

LinaPro runs on macOS, Linux, and Windows. Before you start, make sure the following are on `PATH`:

- `Go 1.21+` for the core host service.
- `Node.js 18+` with `pnpm 9+` for the `Vue 3` management workspace.
- `Make` as the single entrypoint for common local tasks.
- `Git` for source checkout and the `OpenSpec` workflow.
- `MySQL 8.0+` for persistent state.

The default project expects a relational database. Use a local `MySQL` instance for the most predictable first run.

## 1. Install

Run the official installer in the directory where you want the project created:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

On Windows, run the PowerShell installer:

```powershell
irm https://linapro.ai/install.ps1 | iex
```

The installer downloads the repository source archive, creates a `./linapro` directory by default, and runs an environment check. It does not install dependencies or start services automatically. `cd` into the generated project before continuing.

## 2. Initialise

Generate runtime config, apply database migrations, and — optionally — seed demo data:

```bash
make init confirm=init
make mock confirm=mock   # optional — loads demo users, roles, and menus
```

The `confirm=` argument is a guardrail: commands that touch persistent state refuse to run unless you opt in explicitly. Use `make init confirm=init rebuild=true` only when you intentionally want to rebuild the local database.

## 3. Develop

Bring up the backend host and frontend workspace with a single command:

```bash
make dev
```

Once it settles you'll have:

| Service                              | URL                     |
| ------------------------------------ | ----------------------- |
| Management workspace (`lina-vben`)   | http://localhost:5666   |
| Core API (`lina-core`)               | http://localhost:8080   |

Sign in to the workspace with the default account:

- **Username:** `admin`
- **Password:** `admin123`

:::warning
Change the default password before exposing the workspace to a network. **System → User Management** in the workspace is the place to do it.
:::

## Verify your setup

A quick smoke test before you start building:

1. Open `http://localhost:5666` and sign in with the default account.
2. Navigate to **Extension Center → Plugin Management** and confirm the official plugins are visible.
3. Open `http://localhost:8080/api` to confirm the `OpenAPI` documentation is served.

If any of these fail, jump to [Troubleshooting](#troubleshooting).

## What's next

You're now on the same stack production teams use. Pick the path that matches what you're building:

- Read the [workspace tour](/quick/workspace-tour) to understand the built-in management modules.
- Continue with the [developer manual](/docs) when you need architecture, plugin, testing, or deployment details.
- Use the [community page](/community) when you need help or want to contribute.

## Troubleshooting

- **`make: command not found`** — install Make via your package manager (`brew install make`, `apt install make`, etc.).
- **Port 5666 or 8080 already in use** — stop the process holding the port, or change the port in the generated project config and re-run `make dev`.
- **Database connection errors** — open the generated config, point the `database` block at your local `MySQL` instance, then re-run `make init confirm=init`.
- **Anything else** — open an issue on [GitHub](https://github.com/linaproai/linapro/issues); the team monitors it daily.
