---
slug: '/quick/installation'
title: '快速安装'
hide_title: true
description: '本文介绍如何在几分钟内完成 LinaPro 框架的安装与初始配置，包括通过 Git 克隆仓库获取源码、按需初始化官方插件子模块、准备默认 PostgreSQL 数据库、复制并调整配置文件、使用跨平台 linactl 或兼容 make 入口初始化数据库、加载演示数据、启动开发服务和验证安装结果的完整步骤。克隆仓库之前，请先完成环境配置。'
keywords:
  - LinaPro
  - 框架安装
  - 快速开始
  - git clone
  - 克隆仓库
  - 源码下载
  - 稳定版本
  - config.yaml
  - 数据库配置
  - 数据库初始化
  - make init
  - make dev
  - 安装验证
  - 开发服务
  - 环境配置
  - 演示数据
  - make mock
  - linactl
  - 官方插件子模块
---

克隆仓库前，请先参阅[环境配置](/quick/environment)，确保`Go`、`Node.js`、`pnpm`、`PostgreSQL`等必要组件已正确安装。

## 克隆仓库

使用以下命令获取框架源码。

安装最新实验版本：

```bash
git clone --depth 1 https://github.com/linaproai/linapro.git linapro
```
或者指定稳定发布版本，如 v0.1.0：
```bash
git clone --depth 1 https://github.com/linaproai/linapro.git linapro --branch v0.1.0
```

## 启动服务

### 准备 PostgreSQL

`LinaPro`默认使用`PostgreSQL 14+`作为数据库。`make init`和`make dev`不会启动或管理数据库，请先准备可连接的`PostgreSQL`实例。本地开发可以使用以下容器：

```bash
docker run \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=linapro \
  postgres:14-alpine
```

如果本机`5432`端口已被占用，可以将容器映射到其他本机端口，例如`15432:5432`。

### 配置数据库连接

克隆完成后，进入项目目录，将配置模板复制为正式配置文件：

```bash
cd linapro
cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml
```

默认配置使用`postgres:postgres@127.0.0.1:5432`连接`linapro`数据库，如果你的`PostgreSQL`使用了不同的用户名、密码、主机、端口或数据库名，请在此处更新。用编辑器打开`config.yaml`，找到数据库连接部分，将其修改为你本地`PostgreSQL`的实际连接信息：

```yaml title="apps/lina-core/manifest/config/config.yaml"
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```



### 初始化数据库

配置完成后，执行以下命令创建数据库表结构并写入初始数据：

```bash
make init confirm=init
```

`Windows`用户如果使用`PowerShell`，可执行：

```powershell
.\make init confirm=init
```

初始化完成后，数据库中将包含系统所需的基础表结构和默认配置数据。

### 加载演示数据（可选）

配置完成后，执行以下命令加载官方提供的演示数据：

```bash
make mock confirm=mock
```

### 运行环境检查（可选）

执行以下命令检查当前开发环境是否满足要求：

```bash
make env.check
```

如果不满足，则可通过以下指令安装完整的开发环境资源，可能需要较长时间：

```bash
make env.setup
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

## 工具集成

由于当前`AI Coding`工具百花齐放，且各个工具的技能、规范、提示词文件存储路径不尽相同，因此我们提供了一个通用的终端交互式指令，帮助大家将框架提供的`Skills`、项目规范和提示词快速集成到自己熟悉的`AI Coding`工具中，以降低大家的心智负担。

```bash
make agents
```

:::info 提示
需要在进行代码开发之前执行，以便驱动你的`AI Coding`工具能够更高质量地工作。
:::

## 常用命令

以下为开发过程中较频繁使用的指令（完整的开发工具指令集请参考[开发指令](../docs/5000-tools/1000-commands.md)章节）：

```bash
make dev          # 启动前后端服务
make stop         # 停止所有本地服务
make status       # 查看服务运行状态
make build        # 编译生成可发布的二进制/WASM文件
make image        # 构建 Docker 镜像
```

> `Windows cmd.exe`可直接使用`make <指令>`；`PowerShell`使用`.\make <指令>`或`.\make.cmd <指令>`。



## 安装验证

服务启动后，如果能够正常进入管理工作台，说明安装已成功完成。
如果遇到问题，可以通过以下步骤排查：

1. 确认`PostgreSQL`服务已启动且`config.yaml`中的数据库连接配置正确
2. 查看后端日志输出，确认服务是否有异常
3. 执行`make status`检查前后端进程状态
4. 如果问题仍未解决，请前往[社区交流](/community)寻求帮助
