---
slug: '/quick/installation'
title: '快速安装'
hide_title: true
description: '本文介绍如何在几分钟内完成 LinaPro 框架的安装与初始配置，包括通过 Git 克隆仓库获取源码（支持最新实验版本与指定稳定发布版本）、配置文件设置、数据库连接配置与初始化流程，以及加载演示数据、启动开发服务和验证安装结果的完整步骤。克隆仓库之前，请先完成环境配置。'
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
---

克隆仓库前，请先参阅[环境配置](/quick/environment)，确保`Go`、`Node.js`、`pnpm`、`PostgreSQL`等必要组件已正确安装。

## 克隆仓库

使用以下命令获取框架源码：

```bash
# 安装最新实验版本
git clone --depth 1 https://github.com/linaproai/linapro.git linapro

# 或者指定稳定发布版本，如 v0.1.0
git clone --depth 1 --branch v0.1.0 https://github.com/linaproai/linapro.git linapro
```

克隆完成后，初始化官方插件子模块（`lina-plugins`是独立的`Git`子仓库，默认不会随主仓库自动拉取）：

```bash
cd linapro
git submodule update --init --recursive
```

:::info 说明
官方插件目录`apps/lina-plugins/`以`Git submodule`形式管理，与主框架解耦，使主框架更精简轻量。如果只需要运行主框架而不需要官方插件，可以跳过此步骤。
:::

## 启动服务

### 准备 PostgreSQL

`LinaPro`默认使用`PostgreSQL 14+`作为数据库。`make init`和`make dev`不会启动或管理数据库，请先准备可连接的`PostgreSQL`实例。本地开发可以使用以下容器：

```bash
docker run --name linapro-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=linapro \
  -p 5432:5432 \
  --health-cmd pg_isready \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 5 \
  -d postgres:14-alpine
```

如果本机`5432`端口已被占用，可以将容器映射到其他本机端口，例如`15432:5432`，并同步修改`database.default.link`中的端口。

### 配置数据库连接

克隆完成后，进入项目目录，将配置模板复制为正式配置文件：

```bash
cd linapro
cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml
```

用编辑器打开`config.yaml`，找到数据库连接部分，将其修改为你本地`PostgreSQL`的实际连接信息：

```yaml title="apps/lina-core/manifest/config/config.yaml"
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```

默认配置使用`postgres:postgres@127.0.0.1:5432`连接`linapro`数据库，如果你的`PostgreSQL`使用了不同的用户名、密码、主机、端口或数据库名，请在此处更新。

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
make build             # 编译生成可发布的二进制文件
make image             # 构建 Docker 镜像
```

## 安装验证

服务启动后，打开浏览器访问`http://localhost:5666`，使用`admin / admin123`登录，如果能够正常进入管理工作台，说明安装已成功完成。

如果遇到问题，可以通过以下步骤排查：

1. 确认`PostgreSQL`服务已启动且`config.yaml`中的数据库连接配置正确
2. 查看后端日志输出，确认服务是否有异常
3. 执行`make status`检查前后端进程状态
4. 如果问题仍未解决，请前往[社区交流](/community)寻求帮助
