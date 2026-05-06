---
slug: '/quick/environment'
title: 'Environment Setup'
hide_title: true
description: 'A complete guide to setting up the environment required to run LinaPro. Covers version requirements and cross-platform installation instructions for Git, Go (≥ 1.23), Node.js (≥ 20.19), pnpm (≥ 10.0), MySQL (8.0+), and Make, plus the AI-native workflow prerequisites — Claude Code, OpenSpec CLI, and the goframe-v2 skill — on macOS, Linux, and Windows (WSL / Git Bash). Follow this guide before running the LinaPro install script.'
keywords:
  - LinaPro
  - environment setup
  - environment dependencies
  - system requirements
  - Go
  - Node.js
  - pnpm
  - MySQL
  - Git
  - Make
  - Claude Code
  - OpenSpec
  - goframe-v2
  - AI-native workflow
  - skill installation
  - nvm
  - Homebrew
  - macOS installation
  - Linux installation
  - Windows installation
  - WSL
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

The following components must be installed on your machine before running the LinaPro install script. The script checks for each one at startup and exits if any required dependency is missing.

| Component | Minimum version | Purpose |
|-----------|----------------|---------|
| `Git` | — | Version control; required by the installer |
| `Go` | 1.23 | Backend runtime |
| `Node.js` | 20.19 | Frontend build environment |
| `pnpm` | 10.0 | Frontend package manager |
| `MySQL` | 8.0 | Relational database |
| `Make` | — | Project command entrypoint |
| `Claude Code` | latest | AI coding assistant; powers the OpenSpec workflow |
| `OpenSpec` | latest | Spec-driven workflow CLI |
| `goframe-v2` skill | latest | GoFrame-specific AI skill |

:::info
The **GoFrame CLI** (`gf`) does not need to be installed manually. When you run `make dev` or any backend build command, the Makefile detects whether `gf` is present and installs it automatically via `wget`.
:::

### Git

Git is pre-installed on macOS and most Linux distributions. Run `git --version` to confirm. If it's missing:

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS
brew install git

# Ubuntu / Debian
sudo apt install git

# CentOS / RHEL
sudo yum install git
```

</TabItem>
<TabItem value="windows" label="Windows (Git Bash / WSL)">

Download and install [Git for Windows](https://git-scm.com/download/win). All subsequent commands should be run inside Git Bash.

</TabItem>
</Tabs>

### Go

Go 1.25 or later is required. Run `go version` to check your current version.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS (Homebrew)
brew install go

# Linux — download the official prebuilt archive
# Find the latest release at https://go.dev/dl/
sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Follow the Linux steps inside WSL, or download the Windows installer (`.msi`) from [go.dev/dl](https://go.dev/dl/) for a GUI-based setup.

</TabItem>
</Tabs>

### Node.js

Node.js 20.19 or later is required. Using `nvm` to manage versions is recommended.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# Install via nvm (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
# Reload your shell, then run:
nvm install --lts
nvm use --lts

# Or via Homebrew (macOS only)
brew install node
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install `nvm` inside WSL following the Linux steps above, or download the Windows package from [nodejs.org](https://nodejs.org/) and use it alongside WSL.

</TabItem>
</Tabs>

After installation, run `node --version` and confirm the output is at least `v20.19.0`.

### pnpm

`pnpm` is the designated package manager for LinaPro's frontend. Do not substitute `npm` or `yarn`.

```bash
npm install -g pnpm
```

Run `pnpm --version` to confirm the version is 10.0.0 or higher.

### MySQL

MySQL 8.0 or later is required.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS (Homebrew)
brew install mysql
brew services start mysql

# Ubuntu / Debian
sudo apt install mysql-server
sudo systemctl enable --now mysql

# CentOS / RHEL
sudo yum install mysql-server
sudo systemctl enable --now mysqld
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install MySQL inside WSL using the Linux steps above, or use the [MySQL Installer](https://dev.mysql.com/downloads/installer/) on the Windows side and connect from WSL via `127.0.0.1`.

</TabItem>
</Tabs>

LinaPro connects to MySQL using `root:12345678@127.0.0.1:3306` by default. You can change these settings in the project's `config.yaml`.

### Make

Make is typically built into macOS and Linux. If it's not available:

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS — also installs Git and other command-line tools
xcode-select --install

# Ubuntu / Debian
sudo apt install build-essential

# CentOS / RHEL
sudo yum groupinstall "Development Tools"
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Run the appropriate Linux command for your WSL distribution.

</TabItem>
</Tabs>

### Claude Code

LinaPro's AI-native workflow runs entirely through Claude Code. Install it before opening the project.

```bash
npm install -g @anthropic-ai/claude-code
```

Run `claude --version` to confirm the installation. The first time you launch `claude`, you'll be prompted to authorize your account.

The following Agent Skills extend Claude Code for LinaPro development:

| Skill | Required | Install command | Purpose |
|-------|:--------:|----------------|---------|
| `OpenSpec` | Yes | `npm install -g @fission-ai/openspec@latest` | Spec-driven workflow CLI; powers LinaPro's core workflow engine |
| `goframe-v2` | Recommended | `npx skills add github.com/gogf/skills -g` | GoFrame-specific AI skill for code generation, diagnostics, and optimization |
| `find-skills` | Recommended | `npx skills add find-skills -g` | AI skill marketplace search tool for discovering and evaluating skills |

### OpenSpec

OpenSpec is the command-line tool that powers LinaPro's spec-driven workflow. The `/opsx:explore`, `/opsx:propose`, `/opsx:apply`, and `/opsx:archive` skills depend on it.

```bash
npm install -g @fission-ai/openspec@latest
```

### goframe-v2 Skill

`goframe-v2` is a Claude Code skill tailored for LinaPro's Go backend. It bundles GoFrame coding conventions, ORM patterns, and best-practice examples. The skill activates automatically whenever you write or modify backend Go code.

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills` is an AI skill marketplace search tool that helps developers quickly find and evaluate AI skills suited to their project, improving skill selection efficiency.

```bash
npx skills add find-skills -g
```
