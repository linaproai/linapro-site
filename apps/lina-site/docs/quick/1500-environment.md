---
slug: '/quick/environment'
title: '环境配置'
hide_title: true
description: '本文详细介绍运行 LinaPro 框架所需的全部环境依赖，包括 Git、Go（≥ 1.23）、Node.js（≥ 20.19）、pnpm（≥ 10.0）、PostgreSQL（14+）、Make，以及 AI 原生工作流推荐的 Claude Code、OpenSpec CLI 和 goframe-v2 技能的版本要求与安装方法，并提供针对 macOS、Linux 和 Windows（WSL / Git Bash）平台的完整安装引导，帮助开发者在正式安装 LinaPro 之前快速完成本地环境准备。'
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
| `Go` | `1.23+` | 后端服务运行时 |
| `Node.js` | `20.19+` | 前端构建环境 |
| `pnpm` | `10.0+` | 前端包管理器 |
| `PostgreSQL` | `14+` | 默认关系型数据库 |
| `Make` | - | 项目命令入口 |

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
<TabItem value="windows" label="Windows（Git Bash / WSL）">

访问 [git-scm.com](https://git-scm.com/download/win) 下载并安装`Git for Windows`。安装完成后，后续所有命令均在`Git Bash`中执行。

</TabItem>
</Tabs>

### Go

需要`Go 1.23`及以上版本，可通过`go version`确认当前版本。

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS（Homebrew）
brew install go

# Linux — 下载官方预编译包，将版本号替换为最新稳定版
# 最新版本列表：https://go.dev/dl/
sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

</TabItem>
<TabItem value="windows" label="Windows（WSL）">

在`WSL`中按`Linux`步骤安装，或访问 [go.dev/dl](https://go.dev/dl/) 下载`Windows`安装包（`.msi`）完成图形化安装。

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
<TabItem value="windows" label="Windows（WSL）">

在`WSL`中按`Linux`步骤安装`nvm`，或访问 [nodejs.org](https://nodejs.org/) 下载`Windows`版本并配合`WSL`使用。

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
  -e POSTGRES_PASSWORD=12345678 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

该示例使用`postgres:12345678@127.0.0.1:5432`连接`linapro`数据库。如果沿用这个密码，请将项目`config.yaml`中的`database.default.link`改为`pgsql:postgres:12345678@tcp(127.0.0.1:5432)/linapro?sslmode=disable`。

</TabItem>
<TabItem value="windows" label="Windows（WSL）">

在`WSL`中按`Linux`步骤安装，或通过 [PostgreSQL Windows installer](https://www.postgresql.org/download/windows/) 在`Windows`侧安装后，从`WSL`通过`127.0.0.1`访问数据库。

</TabItem>
</Tabs>

`LinaPro`默认使用`postgres:postgres@127.0.0.1:5432`连接`linapro`数据库，对应连接串为`pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`，可在项目的`config.yaml`中修改连接配置。

### Make

`macOS` 和 `Linux` 通常已内置`Make`。如未安装：

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
# macOS（同时会安装 Git 等命令行工具）
xcode-select --install

# Ubuntu / Debian
sudo apt install build-essential

# CentOS / RHEL
sudo yum groupinstall "Development Tools"
```

</TabItem>
<TabItem value="windows" label="Windows（WSL）">

在`WSL`中按对应的`Linux`发行版步骤安装即可。

</TabItem>
</Tabs>

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

`goframe-v2`是专为`LinaPro`后端`Go`代码定制的`Claude Code`技能，内置`GoFrame`编码规范、`ORM`使用模式和最佳实践示例。编写或修改后端`Go`代码时，该技能将自动激活。

```bash
npx skills add github.com/gogf/skills -g
```

### find-skills

`find-skills`是`AI`技能市场搜索工具，帮助开发者快速查找和评估适合项目的`AI`技能，提升技能选型效率。

```bash
npx skills add find-skills -g
```
