---
slug: '/quickstart'
title: 'Quick Start'
sidebar_position: 0
---

Get LinaPro running locally in three commands. When you finish this page the core host service, the management workspace, and the plugin runtime are all up — ready for you to build on, extend, or replace any part.

## Prerequisites

LinaPro runs on macOS, Linux, and Windows. Before you start, make sure the following are on `PATH`:

- **Go** — the core host service (`lina-core`) compiles from source.
- **Node.js** with `pnpm` or `yarn` — for the Vue 3 management workspace (`lina-vben`).
- **Make** — single entrypoint for every common task.
- **Git** — used by the installer and the OpenSpec workflow.

A relational database is required for persistent state. The seeded SQLite file produced by `make mock` is enough for the very first boot; switch to MySQL or PostgreSQL when you're ready.

## 1. Install

Run the official installer in the directory where you want the project created:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

The installer fetches the LinaPro CLI, generates a starter project, and prints the path it dropped you into. `cd` there before continuing.

## 2. Initialise

Generate runtime config, apply database migrations, and — optionally — seed demo data:

```bash
make init confirm=init
make mock confirm=mock   # optional — loads demo users, roles, and menus
```

The `confirm=` argument is a guardrail: commands that touch persistent state refuse to run unless you opt in explicitly. Re-running `make init` is safe — it is idempotent.

## 3. Develop

Bring up the full stack — core host, management workspace, and plugin runtime — with a single command:

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
2. Navigate to **System → Plugin Management** — the official plugins (department, notice, monitor-*) should be listed.
3. Hit `http://localhost:8080/api` to confirm the OpenAPI docs are served.

If any of these fail, jump to [Troubleshooting](#troubleshooting).

## What's next

You're now on the same stack production teams use. Pick the path that matches what you're building:

- **Architecture** — the four layers (`lina-core`, `lina-vben`, `lina-plugins`, `openspec`) and how they interlock.
- **Built-in modules** — User, Role, Menu, Dictionary, Parameter Settings, File, Job Scheduling, Plugin, API Docs, System Info — all RBAC-wired out of the box.
- **Plugins** — write a source plugin for compile-time integration or a WASM plugin for hot-reload distribution. Both run in a sandbox with namespaced database and filesystem access.
- **OpenSpec workflow** — the AI R&D loop (`explore → propose → implement → review → archive`). Every change starts as a spec and lands with E2E tests.

## Troubleshooting

- **`make: command not found`** — install Make via your package manager (`brew install make`, `apt install make`, etc.).
- **Port 5666 or 8080 already in use** — stop the process holding the port, or change the port in the generated project config and re-run `make dev`.
- **Database connection errors** — open the generated config, point the `database` block at your MySQL / PostgreSQL / SQLite instance, then re-run `make init confirm=init`.
- **Anything else** — open an issue on [GitHub](https://github.com/linaproai/linapro/issues); the team monitors it daily.
