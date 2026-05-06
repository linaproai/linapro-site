---
slug: '/quick/installation'
title: '快速安装'
hide_title: true
description: '本文介绍如何在几分钟内完成 LinaPro 框架的安装与初始配置，包括一键安装脚本的使用方式（支持 macOS、Linux 和 Windows 平台）、自定义安装选项、配置文件设置、数据库连接配置与初始化流程，以及启动开发服务和验证安装结果的完整步骤。在运行安装脚本之前，请先完成环境配置。'
keywords:
  - LinaPro
  - 框架安装
  - 快速开始
  - 安装脚本
  - macOS安装
  - Linux安装
  - Windows安装
  - 一键安装
  - config.yaml
  - 数据库配置
  - 数据库初始化
  - make init
  - make dev
  - 安装验证
  - 开发服务
  - 环境配置
  - Git Bash
  - WSL
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

`LinaPro`提供一键安装脚本，支持`macOS`、`Linux`和`Windows`（`Git Bash`或`WSL`）平台，通常可在几分钟内完成完整安装。

运行安装脚本前，请先参阅[环境配置](/quick/environment)，确保`Go`、`Node.js`、`pnpm`、`MySQL`等必要组件已正确安装。

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


## 启动服务

### 配置数据库连接

安装脚本执行完成后，进入项目目录，将配置模板复制为正式配置文件：

```bash
cd linapro
cp config.yaml.example config.yaml
```

用编辑器打开`config.yaml`，找到数据库连接部分，将其修改为你本地`MySQL`的实际连接信息：

```yaml
database:
  default:
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
```

默认配置使用`root:12345678@127.0.0.1:3306`，如果你的`MySQL`使用了不同的用户名、密码或端口，请在此处更新。

### 初始化数据库

配置完成后，执行以下命令创建数据库表结构并写入初始数据：

```bash
make init confirm=init
```

初始化完成后，数据库中将包含系统所需的基础表结构和默认配置数据。

### 加载演示数据（可选）

配置完成后，执行以下命令加载官方提供的演示数据：

```bash
make mock confirm=mock
```

### 启动开发服务

执行以下命令启动前后端服务：

```bash
make dev
```

服务启动成功后，访问以下地址：

| 服务 | 地址 |
|------|------|
| 默认管理工作台 | `http://localhost:5666` |
| 后端`API`服务 | `http://localhost:8080` |

使用默认账号登录管理工作台：

| 字段 | 值 |
|------|-----|
| 账号 | `admin` |
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
4. 如果问题仍未解决，请前往[社区交流](/community)寻求帮助
