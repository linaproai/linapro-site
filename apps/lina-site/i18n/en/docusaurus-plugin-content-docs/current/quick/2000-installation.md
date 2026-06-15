---
slug: '/quick/installation'
title: 'Quick Installation'
hide_title: true
description: 'How to install and configure the LinaPro framework in minutes, including cloning the repository via Git, initializing official plugin submodules as needed, preparing a default PostgreSQL database, copying and adjusting configuration files, initializing the database with cross-platform linactl or compatible make commands, loading demo data, starting the development server, and verifying the installation.'
keywords:
  - LinaPro
  - framework installation
  - quick start
  - git clone
  - clone repository
  - source download
  - stable version
  - config.yaml
  - database configuration
  - database initialization
  - make db.init
  - make dev
  - installation verification
  - development server
  - environment setup
  - demo data
  - make db.mock
  - linactl
  - official plugin submodules
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Before cloning the repository, please refer to [Environment Setup](/quick/environment) to ensure that Go, Node.js, pnpm, PostgreSQL, and other required components are properly installed.

## Clone the Repository

Use the following commands to obtain the framework source code.

Install the latest experimental version:

```bash
git clone https://github.com/linaproai/linapro.git linapro
```
Or specify a stable release version, such as `v0.1.0`:
```bash
git clone https://github.com/linaproai/linapro.git linapro --branch v0.1.0
```

## Start the Service

### Prepare PostgreSQL

LinaPro uses PostgreSQL 14+ as its database by default. `make db.init` and `make dev` do not start or manage the database, so please prepare a connectable PostgreSQL instance first. For local development, you can use the following container:

```bash
docker run \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

If port 5432 on your machine is already in use, you can map the container to a different local port, such as `15432:5432`.

### Configure the Database Connection

After cloning, navigate to the project directory and copy the configuration template to create the actual configuration file:

```bash
cd linapro
cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml
```

The default configuration connects to the `linapro` database using `postgres:postgres@127.0.0.1:5432`. If your PostgreSQL uses a different username, password, host, port, or database name, update it here. Open `config.yaml` in an editor, locate the database connection section, and change it to match your local PostgreSQL connection details:

```yaml title="apps/lina-core/manifest/config/config.yaml"
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```



### Initialize the Database

Once configuration is complete, run the following command to create the database schema and write initial data:

```bash
make db.init confirm=init
```

For Windows users running PowerShell:

```powershell
.\make db.init confirm=init
```

After initialization, the database will contain the base table structures and default configuration data required by the system.

### Load Demo Data (Optional)

After configuration, run the following command to load the provided demo data:

```bash
make db.mock confirm=mock
```

### Run Environment Check (Optional)

Run the following command to verify that your development environment meets the requirements:

```bash
make env.check
```

If requirements are not met, you can install the full development environment resources (frontend and test tooling only) with the following command. This may take a while:

```bash
make env.setup
```

This automatically downloads the Playwright tool for subsequent E2E tests. If downloading Chromium is slow, you can use a regional mirror:

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




### Start the Development Server

Run the following command to start the frontend and backend services:

```bash
make dev
```

After the services start successfully, visit the following addresses:

| Service | Address |
|------|------|
| Frontend dev server | `http://localhost:5666` |
| Backend API server | `http://localhost:9120` |
| Default admin workspace | `http://localhost:5666/admin` |

Log in to the admin workspace with the default credentials:

| Field | Value |
|------|-----|
| Username | `admin` |
| Password | `admin123` |

## Tool Integration

Given the current diversity of AI coding tools and the varying locations where they store skills, rules, and prompt files, we provide a universal interactive terminal command to help you quickly integrate the framework's skills, project rules, and prompts into your preferred AI coding tool, reducing cognitive overhead.

```bash
make agents
```

:::info Note
Run this before starting code development so that your AI coding tool can work more effectively.
:::

## Common Commands

The following are frequently used commands during development (for the complete set, see the [Development Commands](../docs/5000-tools/1000-commands.md) section):

```bash
make dev          # Start frontend and backend services
make stop         # Stop all local services
make status       # Check service running status
make build        # Compile release binary/WASM files
make image        # Build Docker image
```

> On Windows `cmd.exe`, use `make <command>` directly. In PowerShell, use `.\make <command>` or `.\make.cmd <command>`.



## Installation Verification

After the services start, if you can successfully access the admin workspace, the installation is complete.

If you encounter issues, try the following troubleshooting steps:

1. Confirm that PostgreSQL is running and the database connection in `config.yaml` is correct
2. Check the backend log output for any errors
3. Run `make status` to check frontend and backend process status
4. If the issue persists, visit [Community](/community) for help
