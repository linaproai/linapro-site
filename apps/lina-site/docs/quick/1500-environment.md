---
slug: '/quick/environment'
title: '环境配置'
hide_title: true
description: '本文详细介绍运行 LinaPro 框架所需的全部环境依赖，包括 Git、Go（≥ 1.23）、Node.js（≥ 20.19）、pnpm（≥ 10.0）、PostgreSQL（14+）、可选 Redis 集群协调器、跨平台 linactl 开发指令入口，以及 AI 原生工作流推荐的 Claude Code、OpenSpec CLI、goframe-v2 和 find-skills 的版本要求与安装方法，并提供针对 macOS、Linux 和 Windows 平台的安装引导，帮助开发者在正式安装 LinaPro 之前快速完成本地环境准备。'
keywords:
  - LinaPro
  - 环境配置
  - 环境依赖
  - 环境要求
  - Go
  - Node.js
  - pnpm
  - PostgreSQL
  - Git
  - Make
  - linactl
  - Redis
  - Claude Code
  - OpenSpec
  - goframe-v2
  - AI原生工作流
  - 技能安装
  - nvm
  - Homebrew
  - macOS安装
  - Linux安装
  - Windows安装
  - WSL
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## 环境组件

`LinaPro`在源码开发时会依赖以下组件，需在本机预先安装好相关组件。

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| `Git` | - | 版本控制，安装脚本依赖 |
| `Go` | `1.25+` | 后端服务运行时 |
| `Node.js` | `20.19+` | 前端构建环境 |
| `pnpm` | `10.0+` | 前端包管理器 |
| `PostgreSQL` | `14+` | 默认关系型数据库 |

### Git

`macOS` 和大多数 `Linux` 发行版已预装`Git`，可通过`git --version`确认是否可用。如未安装：

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
<TabItem value="windows" label="Windows">

访问 [git-scm.com](https://git-scm.com/download/win) 下载并安装`Git for Windows`。安装完成后，可以在`cmd.exe`、`PowerShell`、`Git Bash`或`WSL`中使用`git`命令。

</TabItem>
</Tabs>

### Go

需要`Go 1.25`及以上版本，可通过`go version`确认当前版本。

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS（Homebrew）
brew install go

# Linux — 下载官方预编译包，将版本号替换为最新稳定版
# 最新版本列表：https://go.dev/dl/
# 以 Linux amd64 go1.25.10 为例：
wget https://go.dev/dl/go1.25.10.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.10.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

</TabItem>
<TabItem value="windows" label="Windows">

访问 [go.dev/dl](https://go.dev/dl/) 下载`Windows`安装包（`.msi`）完成图形化安装，或在`WSL`中按`Linux`步骤安装。

</TabItem>
</Tabs>

### Node.js

推荐通过`nvm`管理`Node.js`版本，最低需要`v20.19.0`。

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# 通过 nvm 安装（推荐）
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/HEAD/install.sh | bash
# 重新加载 shell 后执行
nvm install --lts
nvm use --lts

# 或通过 Homebrew（仅 macOS）
brew install node
```

</TabItem>
<TabItem value="windows" label="Windows">

推荐使用 [nvm-windows](https://github.com/coreybutler/nvm-windows) 或访问 [nodejs.org](https://nodejs.org/) 下载`Windows`版本。也可以在`WSL`中按`Linux`步骤安装`nvm`。

</TabItem>
</Tabs>

安装完成后运行`node --version`，确认输出版本不低于`v20.19.0`。

### pnpm

`pnpm`是`LinaPro`前端工程指定的包管理器，请勿使用`npm`或`yarn`替代。

```bash
npm install -g pnpm
```

安装完成后运行`pnpm --version`，确认输出版本不低于`10.0.0`。

### PostgreSQL

`LinaPro`默认使用`PostgreSQL 14+`作为数据库。运行`make init`或`make dev`之前，请先准备好可连接的`PostgreSQL`实例。

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS（Homebrew）
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

如果本机已安装`Docker`，可以直接启动一个本地`PostgreSQL`容器：

```bash
docker run \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

该示例使用`postgres:postgres@127.0.0.1:5432`连接`linapro`数据库。如果沿用这个密码，请将项目`config.yaml`中的`database.default.link`改为`pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`。

</TabItem>
<TabItem value="windows" label="Windows">

可通过 [PostgreSQL Windows installer](https://www.postgresql.org/download/windows/) 在`Windows`侧安装，也可以使用`Docker Desktop`运行上面的`PostgreSQL`容器。使用`WSL`时，可从`WSL`通过`127.0.0.1`访问`Windows`侧数据库，或直接在`WSL`发行版内安装`PostgreSQL`。

</TabItem>
</Tabs>

`LinaPro`默认使用`postgres:postgres@127.0.0.1:5432`连接`linapro`数据库，对应连接串为`pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`，可在项目的`config.yaml`中修改连接配置。

## 开发技能(Agent Skills)

以下为`LinaPro`推荐安装的`Agent Skills`：

| 技能 | 是否必须 | 作用 |
|------|:-------:|---------|
| `OpenSpec` | 建议 | 可选的规范驱动工作流工具，推荐配合使用以获得最佳体验 |
| `goframe-v2` | 建议 | `GoFrame`专属`AI`技能，提供代码生成、错误诊断和性能优化建议，提升`Go`框架代码生成质量 |
| `find-skills` | 建议 | `AI`技能市场搜索工具，帮助开发者快速查找和评估适合项目的`AI`技能，提升技能选型效率 |

### OpenSpec

`OpenSpec`是可选的规范驱动工作流命令行工具，推荐安装以获得完整的规范驱动工作流体验。安装后，`/opsx:explore`、`/opsx:propose`、`/opsx:apply`和`/opsx:archive`等工作流技能将自动使用`OpenSpec`作为底层引擎。

```bash
npm install -g @fission-ai/openspec@latest
```

### goframe-v2

`goframe-v2`是专为`GoFrame`开发框架提供的`Agent Skill`，内置`GoFrame`编码规范、`ORM`使用模式和最佳实践示例。编写或修改后端`Go`代码时，该技能将自动激活。

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills`是`Agent`技能市场搜索工具，帮助开发者快速查找和评估适合项目的`Agent Skills`，提升技能选型效率。

```bash
npx skills add vercel-labs/skills --skill find-skills -g
```
