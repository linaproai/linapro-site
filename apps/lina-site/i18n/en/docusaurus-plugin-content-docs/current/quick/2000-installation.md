---
slug: '/quick/installation'
title: 'Installation'
hide_title: true
description: 'how to install and initialize the LinaPro framework in a few minutes. It covers cloning the source repository with Git, initializing official plugin submodules when needed, preparing the default PostgreSQL database, copying and adjusting the configuration file, initializing the database through the cross-platform linactl or compatible make entry, loading demo data, starting the development service, integrating AI Coding tools, and verifying that the installation is working correctly.'
keywords:
  - LinaPro
  - framework installation
  - quick start
  - git clone
  - clone repository
  - source download
  - stable release
  - config.yaml
  - database configuration
  - database initialization
  - make init
  - make dev
  - installation verification
  - development service
  - environment setup
  - demo data
  - make mock
  - linactl
  - official plugin submodules
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Before cloning the repository, read [Environment Setup](/quick/environment) and make sure required components such as `Go`, `Node.js`, `pnpm`, and `PostgreSQL` are installed correctly.

## Clone the Repository

Use the following commands to fetch the framework source code:

```bash
# Install the latest experimental version
git clone --depth 1 https://github.com/linaproai/linapro.git linapro

# Or pin to a specific stable release, e.g., v0.1.0
git clone --depth 1 --branch v0.1.0 https://github.com/linaproai/linapro.git linapro
```

## Start the Services

### Prepare PostgreSQL

`LinaPro` uses `PostgreSQL 14+` as its default database. `make init` and `make dev` do not start or manage the database, so prepare a reachable `PostgreSQL` instance first. For local development, you can use this container:

```bash
docker run \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

If local port `5432` is already occupied, map the container to another host port, such as `15432:5432`.

### Configure the Database Connection

After cloning, enter the project directory and copy the config template as the active config file:

```bash
cd linapro
cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml
```

Open `config.yaml` in an editor, find the database connection section, and update it to match your local `PostgreSQL` connection details:

```yaml title="apps/lina-core/manifest/config/config.yaml"
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```

The default configuration connects to the `linapro` database with `postgres:postgres@127.0.0.1:5432`. If your `PostgreSQL` instance uses a different username, password, host, port, or database name, update the value here.

### Initialize the Database

After the config is ready, run the following command to create the database schema and write the initial data:

```bash
make init confirm=init
```

If you use `PowerShell` on `Windows`, run:

```powershell
.\make init confirm=init
```

After initialization, the database contains the basic table structure and default configuration data required by the system.

### Load Demo Data (Optional)

After the config is ready, run the following command to load the official demo data:

```bash
make mock confirm=mock
```

### Check the Environment (Optional)

Run the following command to check whether the current development environment meets the requirements:

```bash
make env.check
```

If requirements are missing, you can install the complete development environment resources (frontend and test resource components only) with the following command. This may take some time:

```bash
make env.setup
```

This command automatically downloads the `Playwright` browser for end-to-end testing. If the `Chromium` download is slow, you can use a mirror to speed it up:

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>
```bash
PLAYWRIGHT_DOWNLOAD_HOST=https://cdn.npmmirror.com/binaries/playwright make env.setup
```
</TabItem>
<TabItem value="windows" label="Windows">
```bash
$env:PLAYWRIGHT_DOWNLOAD_HOST="https://cdn.npmmirror.com/binaries/playwright" make env.setup
```
</TabItem>
</Tabs>

### Start the Development Services

Run the following command to start both frontend and backend services:

```bash
make dev
```

When the services start successfully, visit:

| Service | URL |
|---------|-----|
| Frontend dev server | `http://localhost:5666` |
| Backend `API` service | `http://localhost:9120` |
| Default admin workspace | `http://localhost:5666/admin` |

Log in to the admin workspace with the default account:

| Field | Value |
|-------|-------|
| Account | `admin` |
| Password | `admin123` |

## Tool Integration

Today's `AI Coding` tools use different conventions for storing skills, project rules, and prompt files. LinaPro provides a general interactive terminal command that integrates the framework's `Skills`, project conventions, and prompts into the `AI Coding` tool you already use, reducing the amount of setup you need to remember.

```bash
make agents
```

:::info Tip
Run this before code development so your `AI Coding` tool can work with higher-quality project context.
:::

## Common Commands

The following commands are used frequently during development. For the complete development command set, see [Development Commands](../docs/5000-tools/1000-commands.md).

```bash
make dev          # Start frontend and backend services
make stop         # Stop all local services
make status       # Check service status
make build        # Build distributable binary/WASM artifacts
make image        # Build the Docker image
```

> `Windows cmd.exe` can run `make <command>` directly. In `PowerShell`, use `.\make <command>` or `.\make.cmd <command>`.

## Installation Verification

After the service starts, if you can enter the admin workspace normally, the installation is complete.
If you run into issues, use the following checklist:

1. Confirm that `PostgreSQL` is running and that the database connection in `config.yaml` is correct
2. Check backend log output for service errors
3. Run `make status` to inspect frontend and backend process status
4. If the issue is still unresolved, visit [Community](/community) for help
