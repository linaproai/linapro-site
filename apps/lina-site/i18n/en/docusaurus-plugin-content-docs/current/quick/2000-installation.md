---
slug: '/quick/installation'
title: 'Installation'
hide_title: true
description: 'Get LinaPro up and running in minutes. This guide covers system requirements, the one-command install script for macOS, Linux, and Windows, database initialization, using Claude Code to complete dependency setup, and how to start the service and verify a successful installation.'
keywords:
  - LinaPro
  - installation
  - quick start
  - install script
  - macOS installation
  - Linux installation
  - Windows installation
  - system requirements
  - Go
  - Node.js
  - pnpm
  - MySQL
  - Claude Code
  - make dev
  - database initialization
  - one-click install
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

LinaPro provides a one-command install script that supports macOS, Linux, and Windows (Git Bash or WSL). The full installation typically completes in a few minutes.

## One-Command Install

The official install script handles source checkout, dependency installation, and database initialization automatically.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

After installation, the source code is placed in a `./linapro` subdirectory of your current working directory by default.

**Custom install options:**

```bash
# Install to a specific directory
LINAPRO_DIR=~/workspace/linapro curl -fsSL https://linapro.ai/install.sh | bash

# Install a specific version
LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh | bash

# Skip demo data initialization
LINAPRO_SKIP_MOCK=1 curl -fsSL https://linapro.ai/install.sh | bash
```

</TabItem>
<TabItem value="windows" label="Windows (Git Bash / WSL)">

Windows users should run the install command inside Git Bash or WSL. Running it in native PowerShell is not supported.

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

**WSL note:** If you're using WSL 2, keep the project directory inside the WSL filesystem (e.g., `~/workspace/linapro`) to avoid cross-filesystem performance issues.

</TabItem>
</Tabs>

The install script runs the following steps in sequence:

1. Check dependencies (`Go`, `Node.js`, `pnpm`, `MySQL`)
2. Clone the LinaPro source repository
3. Install backend and frontend dependencies
4. Initialize the database schema and seed data (`make init confirm=init`)
5. Load demo data (skip with `LINAPRO_SKIP_MOCK=1`)

## Setting Up with Claude Code

After the install script finishes, open the project directory in Claude Code and let AI help you complete the remaining environment setup.

**Install Claude Code:**

```bash
npm install -g @anthropic-ai/claude-code
```

**Open the project:**

```bash
cd linapro
claude
```

Once inside Claude Code, just tell AI what you want to do — for example:

```
Check whether the development environment is configured correctly and start the dev service.
```

Claude Code automatically reads `CLAUDE.md` to understand the project structure and conventions, then helps you with environment verification, configuration adjustments, and service startup.

:::info
LinaPro is an AI-native framework. The `CLAUDE.md` at the project root contains complete engineering rules and the development workflow guide — Claude Code uses this as its operating context when assisting with development tasks.
:::

## Starting the Development Service

Once setup is complete, start the frontend and backend with a single command:

```bash
make dev
```

When the service is up, open the following URLs:

| Service | URL |
|---------|-----|
| Default management workspace | `http://localhost:5666` |
| Backend API service | `http://localhost:8080` |

Log in to the management workspace with the default credentials:

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `admin123` |

## Common Commands

```bash
make dev               # Start the frontend and backend services
make stop              # Stop all local services
make status            # Check service process status
make init confirm=init # Re-initialize the database
make mock confirm=mock # Reload demo data
make test              # Run the full E2E test suite
```

## Verifying the Installation

After starting the service, open `http://localhost:5666` in a browser and log in with `admin / admin123`. If you can access the management workspace normally, the installation is complete.

If you run into issues, try these troubleshooting steps:

1. Confirm that MySQL is running and that the database connection settings in `config.yaml` are correct
2. Check the backend log output for any errors
3. Run `make status` to verify that both frontend and backend processes are running
4. If the issue persists, visit [Community](/community) for help
