---
slug: '/quick/installation'
title: '框架安装'
sidebar_position: 2
hide_title: true
description: '本文介绍如何在几分钟内完成 LinaPro 框架的安装与初始配置，包括环境要求、一键安装脚本（支持 macOS、Linux 和 Windows 平台）、数据库初始化、使用 Claude Code 完成依赖安装，以及启动服务和验证安装结果的完整流程。'
keywords:
  - LinaPro
  - 框架安装
  - 快速开始
  - 安装脚本
  - macOS安装
  - Linux安装
  - Windows安装
  - 环境要求
  - Go
  - Node.js
  - pnpm
  - MySQL
  - Claude Code
  - make dev
  - 数据库初始化
  - 一键安装
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

`LinaPro`提供一键安装脚本，支持`macOS`、`Linux`和`Windows`（`Git Bash`或`WSL`）平台，通常可在几分钟内完成完整安装。

## 环境要求

在开始安装前，请确认本机已具备以下运行环境：

| 依赖 | 最低版本 | 说明 |
|------|---------|------|
| `Go` | 1.22+ | 后端运行时编译环境 |
| `Node.js` | 20+ | 前端构建环境 |
| `pnpm` | 8+ | 前端包管理工具 |
| `MySQL` | 8.0+ | 数据库，需提前启动并创建好数据库 |
| `Git` | 任意稳定版 | 源码克隆工具 |

:::tip
推荐使用[`nvm`](https://github.com/nvm-sh/nvm)管理`Node.js`版本，使用[`gvm`](https://github.com/moovweb/gvm)或官方安装包管理`Go`版本。
:::

## 一键安装

`LinaPro`提供官方安装脚本，执行后会自动完成源码克隆、依赖安装和数据库初始化。

<Tabs groupId="platform">
<TabItem value="mac-linux" label="macOS / Linux" default>

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

安装完成后，源码默认位于当前工作目录下的`./linapro`子目录中。

**自定义安装选项：**

```bash
# 指定安装目录
LINAPRO_DIR=~/workspace/linapro curl -fsSL https://linapro.ai/install.sh | bash

# 安装指定版本
LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh | bash

# 跳过演示数据初始化
LINAPRO_SKIP_MOCK=1 curl -fsSL https://linapro.ai/install.sh | bash
```

</TabItem>
<TabItem value="windows" label="Windows（Git Bash / WSL）">

`Windows`用户请在`Git Bash`或`WSL`环境中执行安装命令，不支持在原生`PowerShell`中直接运行。

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

**WSL 用户提示：** 如果使用`WSL 2`，建议将项目目录放在`WSL`文件系统内（如`~/workspace/linapro`），避免跨文件系统访问导致的性能问题。

</TabItem>
</Tabs>

安装脚本会依次执行以下步骤：

1. 检查依赖环境（`Go`、`Node.js`、`pnpm`、`MySQL`）
2. 克隆`LinaPro`源码仓库
3. 安装后端和前端依赖
4. 初始化数据库结构和基础数据（`make init confirm=init`）
5. 加载演示数据（可通过`LINAPRO_SKIP_MOCK=1`跳过）

## 使用 Claude Code 完成准备工作

安装脚本执行完成后，推荐使用`Claude Code`打开项目目录，让`AI`帮助完成后续环境配置和准备工作。

**安装 Claude Code：**

```bash
npm install -g @anthropic-ai/claude-code
```

**打开项目目录：**

```bash
cd linapro
claude
```

进入`Claude Code`后，直接告诉`AI`你想做什么，例如：

```
帮我检查一下开发环境是否配置正确，并启动开发服务
```

`Claude Code`会自动读取项目配置文件（`CLAUDE.md`），了解项目结构和规范，然后帮助你完成环境验证、配置调整和服务启动等工作。

:::info
`LinaPro`是一款`AI`原生框架，项目根目录下的`CLAUDE.md`包含了完整的工程规则和研发工作流说明，`Claude Code`会以此为依据辅助你完成开发工作。
:::

## 启动开发服务

完成上述准备工作后，执行以下命令启动前后端服务：

```bash
make dev
```

服务启动成功后，访问以下地址：

| 服务 | 地址 |
|------|------|
| 默认管理工作台 | `http://localhost:5666` |
| 后端`API`服务 | `http://localhost:8080` |
| 接口文档 | `http://localhost:8080/api.json` |

使用默认账号登录管理工作台：

| 字段 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `admin123` |

## 常用命令

```bash
make dev               # 启动前后端服务
make stop              # 停止所有本地服务
make status            # 查看服务运行状态
make init confirm=init # 重新初始化数据库
make mock confirm=mock # 重新加载演示数据
make test              # 运行完整 E2E 测试套件
```

## 安装验证

服务启动后，打开浏览器访问`http://localhost:5666`，使用`admin / admin123`登录，如果能够正常进入管理工作台，说明安装已成功完成。

如果遇到问题，可以通过以下步骤排查：

1. 确认`MySQL`服务已启动且`config.yaml`中的数据库连接配置正确
2. 查看后端日志输出，确认服务是否有异常
3. 执行`make status`检查前后端进程状态
4. 如果问题仍未解决，请前往[社区交流](/community/channels)寻求帮助
