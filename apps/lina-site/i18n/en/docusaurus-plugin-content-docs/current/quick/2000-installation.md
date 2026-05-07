---
slug: '/quick/installation'
title: 'Installation'
hide_title: true
description: 'This guide explains how to install and initialize LinaPro in a few minutes, including how to use the one-command install script on macOS, Linux, and Windows, how to customize installation options, how to set up config.yaml and the database connection, how to initialize the database, how to load demo data, and how to start the development service and verify that the installation works. Complete the environment setup first.'
keywords:
  - LinaPro
  - framework installation
  - quick start
  - install script
  - macOS installation
  - Linux installation
  - Windows installation
  - one-command install
  - config.yaml
  - database configuration
  - database initialization
  - make init
  - make dev
  - installation verification
  - development service
  - environment setup
  - Git Bash
  - WSL
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

LinaPro provides a one-command install script for `macOS`, `Linux`, and `Windows` (`Git Bash` or `WSL`). A complete installation usually finishes in a few minutes.

Before running the install script, read [Environment Setup](/quick/environment) and make sure required components such as `Go`, `Node.js`, `pnpm`, and `MySQL` are installed correctly.

## One-Command Install

LinaPro provides an official install script that automatically clones the source code, installs dependencies, and initializes the database.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

After installation, the source code is placed in a `./linapro` subdirectory under the current working directory by default.

**Custom install options:**

```bash
# Specify the installation directory
LINAPRO_DIR=~/workspace/linapro curl -fsSL https://linapro.ai/install.sh | bash

# Install a specific version
LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh | bash
```

</TabItem>
<TabItem value="windows" label="Windows (Git Bash / WSL)">

`Windows` users should run the install command in `Git Bash` or `WSL`. Running it directly in native `PowerShell` is not supported.

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

**WSL note:** If you use `WSL 2`, keep the project directory inside the `WSL` filesystem, such as `~/workspace/linapro`, to avoid performance issues caused by cross-filesystem access.

</TabItem>
</Tabs>

## Starting the Service

### Configure the Database Connection

After the install script finishes, enter the project directory and copy the config template as the active config file:

```bash
cd linapro
cp config.yaml.example config.yaml
```

Open `config.yaml` in an editor, find the database connection section, and update it to match your local `MySQL` connection details:

```yaml
database:
  default:
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
```

The default configuration uses `root:12345678@127.0.0.1:3306`. If your `MySQL` instance uses a different username, password, or port, update the value here.

### Initialize the Database

After the config is ready, run the following command to create the database schema and write the initial data:

```bash
make init confirm=init
```

After initialization, the database contains the basic table structure and default configuration data required by the system.

### Load Demo Data (Optional)

After the config is ready, run the following command to load the official demo data:

```bash
make mock confirm=mock
```

### Start the Development Service

Run the following command to start both frontend and backend services:

```bash
make dev
```

When the services start successfully, visit:

| Service | URL |
|---------|-----|
| Default management workspace | `http://localhost:5666` |
| Backend `API` service | `http://localhost:8080` |

Log in to the management workspace with the default account:

| Field | Value |
|-------|-------|
| Account | `admin` |
| Password | `admin123` |

## Common Commands

```bash
make dev               # Start frontend and backend services
make stop              # Stop all local services
make status            # Check service status
make init confirm=init # Reinitialize the database
make mock confirm=mock # Reload demo data
make test              # Run the full E2E test suite
```

## Installation Verification

After the service starts, open `http://localhost:5666` in your browser and log in with `admin / admin123`. If you can enter the management workspace normally, the installation is complete.

If you run into issues, use the following checklist:

1. Confirm that `MySQL` is running and that the database connection in `config.yaml` is correct
2. Check backend log output for service errors
3. Run `make status` to inspect frontend and backend process status
4. If the issue is still unresolved, visit [Community](/community) for help
