---
slug: '/quick/environment'
title: 'Environment Setup'
hide_title: true
description: 'This guide covers every environment dependency required to run LinaPro from source, including Git, Go (>= 1.23), Node.js (>= 20.19), pnpm (>= 10.0), PostgreSQL (14+), Make, and the recommended AI-native workflow tools Claude Code, OpenSpec CLI, and the goframe-v2 skill. It also provides installation guidance for macOS, Linux, and Windows (WSL / Git Bash) so you can prepare your local machine before installing LinaPro.'
keywords:
  - LinaPro
  - environment setup
  - environment dependencies
  - system requirements
  - Go
  - Node.js
  - pnpm
  - PostgreSQL
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

## Environment Components

LinaPro depends on the following components during source development. Install them locally before continuing.

| Component | Version requirement | Notes |
|-----------|--------------------|-------|
| `Git` | - | Version control; required by the install script |
| `Go` | `1.25+` | Backend service runtime |
| `Node.js` | `20.19+` | Frontend build environment |
| `pnpm` | `10.0+` | Frontend package manager |
| `PostgreSQL` | `14+` | Default relational database |

### Git

`Git` is preinstalled on `macOS` and most `Linux` distributions. Run `git --version` to confirm that it is available. If it is not installed:

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

Download and install `Git for Windows` from [git-scm.com](https://git-scm.com/download/win). After installation, run all subsequent commands in `Git Bash`.

</TabItem>
</Tabs>

### Go

`Go 1.23` or later is required. Run `go version` to check your current version.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS (Homebrew)
brew install go

# Linux — download the official prebuilt package and replace the version with the latest stable release
# Latest versions: https://go.dev/dl/
sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install it inside `WSL` using the `Linux` steps, or download the `Windows` installer (`.msi`) from [go.dev/dl](https://go.dev/dl/) and complete the graphical setup.

</TabItem>
</Tabs>

### Node.js

Using `nvm` to manage `Node.js` versions is recommended. The minimum required version is `v20.19.0`.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# Install via nvm (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
# Reload your shell, then run
nvm install --lts
nvm use --lts

# Or via Homebrew (macOS only)
brew install node
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install `nvm` inside `WSL` using the `Linux` steps, or download the `Windows` build from [nodejs.org](https://nodejs.org/) and use it together with `WSL`.

</TabItem>
</Tabs>

After installation, run `node --version` and confirm that the output is at least `v20.19.0`.

### pnpm

`pnpm` is the package manager used by LinaPro's frontend project. Do not replace it with `npm` or `yarn`.

```bash
npm install -g pnpm
```

After installation, run `pnpm --version` and confirm that the output is at least `10.0.0`.

### PostgreSQL

LinaPro uses `PostgreSQL 14+` as its default database. Before running `make db.init` or `make dev`, prepare a reachable `PostgreSQL` instance.

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS (Homebrew)
brew install postgresql@14
brew services start postgresql@14

# Ubuntu / Debian
sudo apt install postgresql
sudo systemctl enable --now postgresql

# CentOS / RHEL
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql
```

</TabItem>
<TabItem value="docker" label="Docker">

If `Docker` is already installed locally, you can start a local `PostgreSQL` container directly:

```bash
docker run \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=12345678 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

This example connects to the `linapro` database with `postgres:12345678@127.0.0.1:5432`. If you keep this password, update `database.default.link` in the project's `config.yaml` to `pgsql:postgres:12345678@tcp(127.0.0.1:5432)/linapro?sslmode=disable`.

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install `PostgreSQL` inside `WSL` using the `Linux` steps, or install it on the `Windows` side with the [PostgreSQL Windows installer](https://www.postgresql.org/download/windows/) and connect from `WSL` through `127.0.0.1`.

</TabItem>
</Tabs>

By default, LinaPro connects to the `linapro` database with `postgres:postgres@127.0.0.1:5432`, using the link `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`. You can change the connection settings in the project's `config.yaml`.

### Make

`Make` is usually built into `macOS` and `Linux`. If it is not installed:

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS (also installs Git and other command-line tools)
xcode-select --install

# Ubuntu / Debian
sudo apt install build-essential

# CentOS / RHEL
sudo yum groupinstall "Development Tools"
```

</TabItem>
<TabItem value="windows" label="Windows (WSL)">

Install it inside `WSL` using the command for your Linux distribution.

</TabItem>
</Tabs>

## Development Skills (Agent Skills)

LinaPro recommends installing the following `Agent Skills`:

| Skill | Required | Purpose |
|-------|:--------:|---------|
| `OpenSpec` | Recommended | Optional spec-driven workflow tool — recommended for the best experience |
| `goframe-v2` | Recommended | GoFrame-specific `AI` skill that provides code generation, diagnostics, and performance optimization guidance to improve generated `Go` code quality |
| `find-skills` | Recommended | `AI` skill marketplace search tool for quickly finding and evaluating skills that fit the project |

### OpenSpec

`OpenSpec` is an optional spec-driven workflow command-line tool. Install it to unlock the full spec-driven workflow experience. Once installed, workflow skills such as `/opsx:explore`, `/opsx:propose`, `/opsx:apply`, and `/opsx:archive` will automatically use it as their underlying engine.

```bash
npm install -g @fission-ai/openspec@latest
```

### goframe-v2

`goframe-v2` is a `Claude Code` skill tailored for LinaPro's backend `Go` code. It includes `GoFrame` coding rules, `ORM` usage patterns, and best-practice examples. The skill activates automatically when you write or modify backend `Go` code.

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills` is an `AI` skill marketplace search tool that helps developers quickly find and evaluate `AI` skills suited to the project, improving skill selection efficiency.

```bash
npx skills add vercel-labs/skills --skill find-skills -g
```
